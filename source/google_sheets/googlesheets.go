/*
Copyright © 2022 Meroxa, Inc.

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

package googlesheets

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/conduitio/conduit-connector-google-sheets/source/config"
	"github.com/conduitio/conduit-connector-google-sheets/source/position"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"google.golang.org/api/option"
	sheets "google.golang.org/api/sheets/v4"
)

type SheetsClient struct {
	client        *http.Client
	nextRun       time.Time
	spreadsheetID string
	sheetID       int64
	retryCount    int64
	pollingPeriod time.Duration
}

func NewClient(sClient *http.Client, config config.Config, tp position.SheetPosition) *SheetsClient {
	return &SheetsClient{
		client:        sClient,
		spreadsheetID: config.GoogleSpreadsheetID,
		sheetID:       config.GoogleSheetID,
		pollingPeriod: config.PollingPeriod,
	}
}

// getSheetRecords returns the list of records up to a maximum of 1000 rows(default limit)
// added after the row offset of last successfully read record
func (g *SheetsClient) GetSheetRecords(ctx context.Context, offset int64) ([]sdk.Record, error) {
	if g.nextRun.After(time.Now()) {
		return nil, nil
	}

	sheetService, err := sheets.NewService(ctx, option.WithHTTPClient(g.client))
	if err != nil {
		return nil, err
	}
	var s sheets.DataFilter
	dataFilters := make([]*sheets.DataFilter, 0)
	s.GridRange = &sheets.GridRange{
		SheetId:       g.sheetID,
		StartRowIndex: offset,
	}
	dataFilters = append(dataFilters, &s)
	valueRenderOption := ""
	dateTimeRenderOption := "FORMATTED_STRING"
	rbt := &sheets.BatchGetValuesByDataFilterRequest{
		ValueRenderOption:    valueRenderOption,
		DataFilters:          dataFilters,
		DateTimeRenderOption: dateTimeRenderOption,
	}
	res, err := sheetService.Spreadsheets.Values.BatchGetByDataFilter(g.spreadsheetID, rbt).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	if res.HTTPStatusCode == http.StatusTooManyRequests {
		g.retryCount++
		duration := time.Duration(g.retryCount * int64(g.pollingPeriod)) // exponential back off
		g.nextRun = time.Now().Add(duration)
		sdk.Logger(ctx).Error().
			Int64("retry_count", g.retryCount).
			Float64("wait_duration", duration.Seconds()).
			Msg("exponential back off, rate limit exceeded")
		return nil, nil
	}

	if res.HTTPStatusCode != http.StatusOK {
		return nil, fmt.Errorf("non 200 status code received(%v)", res.HTTPStatusCode)
	}

	records := make([]sdk.Record, 0)

	for i := range res.ValueRanges {
		valueRange := res.ValueRanges[i].ValueRange

		values := valueRange.Values

		for index, val := range values {
			if len(val) == 0 {
				break
			}

			rawData, err := json.Marshal(val)
			if err != nil {
				return records, fmt.Errorf("error marshaling the map: %w", err)
			}
			rowOffset := offset + int64(index) + 1
			lastRowPosition := position.SheetPosition{
				RowOffset: rowOffset,
			}

			records = append(records, sdk.Record{
				Position:  lastRowPosition.RecordPosition(),
				Metadata:  nil,
				CreatedAt: time.Now(),
				Key:       sdk.RawData(fmt.Sprintf("A%d", rowOffset)),
				Payload:   sdk.RawData(rawData),
			})
		}
	}
	g.retryCount = 0
	return records, nil
}
