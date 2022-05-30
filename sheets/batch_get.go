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
	// spreadsheet ID of the Google sheet
	spreadsheetID string
	// gid of the sheet extracted from the sheet URL <url>#gid=<gid>
	sheetID int64
	// instance of sheets service, used to interact with Google Sheets APIs
	sheetSvc *sheets.Service
	// If rate limit is exceeded, nextRun is used to skip hitting API till the specified time.
	// Exponential backoff is used to decide nextRun time on rate-limit error, uses retryCount*pollingPeriod seconds as duration
	nextRun time.Time
	// polling period defined in the config, it is used to implement exponential backoff
	pollingPeriod time.Duration
	// the count of unsuccessful retries made after getting 429(rate-limit exceeded) http error status.
	retryCount int64
	// dateTimeRenderOption Determines how dates, times, and durations in the response should be rendered.
	// This is ignored if responseValueRenderOption is FORMATTED_VALUE.
	// The default dateTime render option is FORMATTED_STRING for the connector.
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
		return nil, fmt.Errorf("error creating sheets service client: %w", err)
	}
	return &BatchReader{
		spreadsheetID:        args.SpreadsheetID,
		sheetID:              args.SheetID,
		pollingPeriod:        args.PollingPeriod,
		sheetSvc:             sheetService,
		dateTimeRenderOption: args.DateTimeRenderOption,
		valueRenderOption:    args.ValueRenderOption,
	}, nil
}

// GetSheetRecords returns the list of records up to a maximum of 1000 rows(default limit)
// added after the row offset of last successfully read record
func (b *BatchReader) GetSheetRecords(ctx context.Context, offset int64) ([]sdk.Record, error) {
	if b.nextRun.After(time.Now()) {
		return nil, nil
	}

	res, err := b.sheetSvc.Spreadsheets.Values.BatchGetByDataFilter(b.spreadsheetID, b.getDataFilter(offset)).Context(ctx).Do()
	if err != nil {
		if googleapi.IsNotModified(err) {
			return nil, nil
		}
		if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == http.StatusTooManyRequests {
			b.retryCount++
			duration := time.Duration(b.retryCount * int64(b.pollingPeriod)) // exponential back off
			b.nextRun = time.Now().Add(duration)
			sdk.Logger(ctx).Error().Err(gerr).
				Int64("retry_count", b.retryCount).
				Float64("wait_duration", duration.Seconds()).
				Msg("exponential back off, rate limit exceeded")
			return nil, nil
		}
		return nil, fmt.Errorf("error getting sheet(gid:%v) values, %w", b.sheetID, err)
	}

	b.retryCount = 0
	return b.valueRangesToRecords(res.ValueRanges, offset)
}

func (b *BatchReader) getDataFilter(offset int64) *sheets.BatchGetValuesByDataFilterRequest {
	dataFilters := make([]*sheets.DataFilter, 0)
	dataFilters = append(dataFilters, &sheets.DataFilter{
		GridRange: &sheets.GridRange{
			SheetId:       b.sheetID,
			StartRowIndex: offset,
		},
	})
	return &sheets.BatchGetValuesByDataFilterRequest{
		DataFilters:          dataFilters,
		DateTimeRenderOption: b.dateTimeRenderOption,
		MajorDimension:       majorDimension,
		ValueRenderOption:    b.valueRenderOption,
	}
}

func (b *BatchReader) valueRangesToRecords(valueRanges []*sheets.MatchedValueRange, offset int64) ([]sdk.Record, error) {
	records := make([]sdk.Record, 0)

	// As we can fetch multiple ranges in one BatchGetByDataFilter request
	// iterate over all the value ranges fetched from the Google sheet BatchGet API request
	// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/batchGetByDataFilter#response-body
	for _, valueRange := range valueRanges {
		rowValues := valueRange.ValueRange.Values
		// Iterate over the Rows of the value range
		// Data is of format: [][]interface{} => ([ [ROW1 => A1,B1,C1..], [ROW2 => A2, B2, C2,...],...])
		for index, rowValue := range rowValues {
			if len(rowValue) == 0 {
				continue
			}
			rawData, err := json.Marshal(rowValue)
			if err != nil {
				return records, fmt.Errorf("error marshaling the map: %w", err)
			}
			rowOffset := offset + int64(index) + 1
			lastRowPosition := position.SheetPosition{
				RowOffset:     rowOffset,
				SpreadsheetID: b.spreadsheetID,
				SheetID:       b.sheetID,
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
