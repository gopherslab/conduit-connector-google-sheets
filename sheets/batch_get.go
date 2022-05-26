/*
Copyright © 2022 Meroxa, Inc. & Gophers Lab Technologies Pvt. Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sheets

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/conduitio/conduit-connector-google-sheets/source/position"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"golang.org/x/oauth2"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const majorDimension = "ROWS"

type BatchReader struct {
	// If an error encountered, after how much time duration, the api should hit next.
	nextRun       time.Time
	spreadsheetID string
	sheetID       int64
	// retry mechanism after getting http error status. maximum is 3.
	retryCount    int64
	sheets        *sheets.Service
	pollingPeriod time.Duration
	// dateTimeRenderOption Determines how dates, times, and durations in the response should be rendered.
	// This is ignored if responseValueRenderOption is FORMATTED_VALUE.
	// The default dateTime render option is SERIAL_NUMBER.
	dateTimeRenderOption string
	// valueRenderOption Determines how values in the response should be rendered.
	// The default render option is FORMATTED_VALUE.
	valueRenderOption string
}

type BatchReaderArgs struct {
	OAuthConfig          *oauth2.Config
	OAuthToken           *oauth2.Token
	SpreadsheetID        string
	SheetID              int64
	DateTimeRenderOption string
	ValueRenderOption    string
	PollingPeriod        time.Duration
}

func NewBatchReader(ctx context.Context, args BatchReaderArgs) (*BatchReader, error) {
	sheetService, err := sheets.NewService(ctx, option.WithHTTPClient(args.OAuthConfig.Client(ctx, args.OAuthToken)))
	if err != nil {
		return nil, err
	}
	return &BatchReader{
		spreadsheetID:        args.SpreadsheetID,
		sheetID:              args.SheetID,
		pollingPeriod:        args.PollingPeriod,
		sheets:               sheetService,
		dateTimeRenderOption: args.DateTimeRenderOption,
		valueRenderOption:    args.ValueRenderOption,
	}, nil
}

// GetSheetRecords returns the list of records up to a maximum of 1000 rows(default limit)
// added after the row offset of last successfully read record
func (g *BatchReader) GetSheetRecords(ctx context.Context, offset int64) ([]sdk.Record, error) {
	if g.nextRun.After(time.Now()) {
		return nil, nil
	}

	res, err := g.sheets.Spreadsheets.Values.BatchGetByDataFilter(g.spreadsheetID, g.getDataFilter(offset)).Context(ctx).Do()
	if err != nil {
		if googleapi.IsNotModified(err) {
			return nil, nil
		}
		if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == http.StatusTooManyRequests {
			g.retryCount++
			duration := time.Duration(g.retryCount * int64(g.pollingPeriod)) // exponential back off
			g.nextRun = time.Now().Add(duration)
			sdk.Logger(ctx).Error().
				Int64("retry_count", g.retryCount).
				Float64("wait_duration", duration.Seconds()).
				Msg("exponential back off, rate limit exceeded")
			return nil, nil
		}
		return nil, err
	}

	g.retryCount = 0
	return g.valueRangesToRecords(res.ValueRanges, offset)
}

func (g *BatchReader) getDataFilter(offset int64) *sheets.BatchGetValuesByDataFilterRequest {
	dataFilters := make([]*sheets.DataFilter, 0)
	dataFilters = append(dataFilters, &sheets.DataFilter{
		GridRange: &sheets.GridRange{
			SheetId:       g.sheetID,
			StartRowIndex: offset,
		},
	})
	return &sheets.BatchGetValuesByDataFilterRequest{
		DataFilters:          dataFilters,
		DateTimeRenderOption: g.dateTimeRenderOption,
		MajorDimension:       majorDimension,
		ValueRenderOption:    g.valueRenderOption,
	}
}

func (g *BatchReader) valueRangesToRecords(valueRanges []*sheets.MatchedValueRange, offset int64) ([]sdk.Record, error) {
	records := make([]sdk.Record, 0)
	for i := range valueRanges {
		values := valueRanges[i].ValueRange.Values
		// values := valueRange.Values
		for index, val := range values {
			if len(val) == 0 {
				continue
			}
			rawData, err := json.Marshal(val)
			if err != nil {
				return records, fmt.Errorf("error marshaling the map: %w", err)
			}
			rowOffset := offset + int64(index) + 1
			lastRowPosition := position.SheetPosition{
				RowOffset:     rowOffset,
				SpreadsheetID: g.spreadsheetID,
				SheetID:       g.sheetID,
			}

			records = append(records, sdk.Record{
				Position:  lastRowPosition.RecordPosition(),
				Metadata:  nil,
				CreatedAt: time.Now(),
				Key:       sdk.RawData(fmt.Sprintf("%d", rowOffset)),
				Payload:   sdk.RawData(rawData),
			})
		}
	}
	return records, nil
}
