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

package iterator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/conduitio/conduit-connector-google-sheets/source/position"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"google.golang.org/api/option"
	sheets "google.golang.org/api/sheets/v4"
	"gopkg.in/tomb.v2"
)

type SheetsIterator struct {
	client        *http.Client
	nextRun       time.Time
	rowOffset     int64
	tomb          *tomb.Tomb
	ticker        *time.Ticker
	caches        chan []sdk.Record
	buffer        chan sdk.Record
	spreadsheetID string
	sheetID       int64
	retryCount    int64
	pollingPeriod time.Duration
}

// NewSheetsIterator creates a new instance of sheets iterator and starts polling google sheets api for new changes
// using the row offset of last successful row read in a separate go routine, row offset is received in sheet position
func NewSheetsIterator(ctx context.Context,
	client *http.Client,
	tp position.SheetPosition,
	spreadsheetID string,
	sheetID int64,
	pollingPeriod time.Duration) (*SheetsIterator, error) {
	tmbWithCtx, ctx := tomb.WithContext(ctx)
	cdc := &SheetsIterator{
		client:        client,
		rowOffset:     tp.RowOffset,
		tomb:          tmbWithCtx,
		caches:        make(chan []sdk.Record, 1),
		buffer:        make(chan sdk.Record, 1),
		ticker:        time.NewTicker(pollingPeriod),
		spreadsheetID: spreadsheetID,
		sheetID:       sheetID,
		pollingPeriod: pollingPeriod,
	}

	cdc.tomb.Go(cdc.startIterator(ctx))
	cdc.tomb.Go(cdc.flush)

	return cdc, nil
}

// startIterator is the go routine function used to poll the google sheets API for new changes at regular intervals
func (c *SheetsIterator) startIterator(ctx context.Context) func() error {
	return func() error {
		defer close(c.caches)
		for {
			select {
			case <-c.tomb.Dying():
				return c.tomb.Err()
			case <-c.ticker.C:
				records, err := c.getSheetRecords(ctx)
				if err != nil {
					return err
				}
				if len(records) == 0 {
					continue
				}
				select {
				case c.caches <- records:
					pos, err := position.ParseRecordPosition(records[len(records)-1].Position)
					if err != nil {
						return err
					}
					c.rowOffset = pos.RowOffset
				case <-c.tomb.Dying():
					return c.tomb.Err()
				}
			}
		}
	}
}

// flush is the go routine, responsible for getting the array of records in caches channel
// and pushing them into read buffer to be returned by Next function
func (c *SheetsIterator) flush() error {
	defer close(c.buffer)
	for {
		select {
		case <-c.tomb.Dying():
			return c.tomb.Err()
		case cache := <-c.caches:
			for _, record := range cache {
				select {
				case c.buffer <- record:
				case <-c.tomb.Dying():
					return c.tomb.Err()
				}

			}
		}
	}
}

// HasNext returns whether there are any more records to be returned
func (c *SheetsIterator) HasNext(_ context.Context) bool {
	return len(c.buffer) > 0 || !c.tomb.Alive() // return true if tomb is dead, call to Next() will return error
}

// Next returns the next record in buffer and error in case there are no more records
// and there was an error leading to tomb dying or context was cancelled
func (c *SheetsIterator) Next(ctx context.Context) (sdk.Record, error) {
	select {
	case rec := <-c.buffer:
		return rec, nil
	case <-c.tomb.Dying():
		return sdk.Record{}, c.tomb.Err()
	case <-ctx.Done():
		return sdk.Record{}, ctx.Err()
	}
}

// Stop stops the go routines
func (c *SheetsIterator) Stop() {
	c.ticker.Stop()
	c.tomb.Kill(errors.New("iterator stopped"))
}

// getSheetRecords returns the list of records up to a maximum of 1000 rows(default limit)
// added after the row offset of last successfully read record
func (c *SheetsIterator) getSheetRecords(ctx context.Context) ([]sdk.Record, error) {
	if c.nextRun.After(time.Now()) {
		return nil, nil
	}

	sheetService, err := sheets.NewService(ctx, option.WithHTTPClient(c.client))
	if err != nil {
		return nil, err
	}
	var s sheets.DataFilter
	dataFilters := make([]*sheets.DataFilter, 0)
	s.GridRange = &sheets.GridRange{
		SheetId:       c.sheetID,
		StartRowIndex: c.rowOffset,
	}
	dataFilters = append(dataFilters, &s)
	valueRenderOption := ""
	dateTimeRenderOption := "FORMATTED_STRING"
	rbt := &sheets.BatchGetValuesByDataFilterRequest{
		ValueRenderOption:    valueRenderOption,
		DataFilters:          dataFilters,
		DateTimeRenderOption: dateTimeRenderOption,
	}
	res, err := sheetService.Spreadsheets.Values.BatchGetByDataFilter(c.spreadsheetID, rbt).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	valueRange := res.ValueRanges[0].ValueRange

	if res.HTTPStatusCode == http.StatusTooManyRequests {
		c.retryCount++
		duration := time.Duration(c.retryCount * int64(c.pollingPeriod)) // exponential back off
		c.nextRun = time.Now().Add(duration)
		sdk.Logger(ctx).Error().
			Int64("retry_count", c.retryCount).
			Float64("wait_duration", duration.Seconds()).
			Msg("exponential back off, rate limit exceeded")
		return nil, nil
	}

	if res.HTTPStatusCode != http.StatusOK {
		return nil, fmt.Errorf("non 200 status code received(%v)", res.HTTPStatusCode)
	}

	responseData := valueRange.Values

	records := make([]sdk.Record, 0, len(responseData))
	for index, val := range responseData {

		rawData, err := json.Marshal(val)
		if err != nil {
			return records, fmt.Errorf("error marshaling the map: %w", err)
		}
		rowOffset := c.rowOffset + int64(index) + 1
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
	c.retryCount = 0
	return records, nil
}
