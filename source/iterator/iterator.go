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
	scfg "github.com/conduitio/conduit-connector-google-sheets/source/sourceconfig"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"google.golang.org/api/option"
	sheets "google.golang.org/api/sheets/v4"
	"gopkg.in/tomb.v2"
)

type SheetsIterator struct {
	client           *http.Client
	nextRun          time.Time
	rowOffset        int64
	lastModifiedTime time.Time
	tomb             *tomb.Tomb
	ticker           *time.Ticker
	caches           chan []sdk.Record
	buffer           chan sdk.Record
	cfg              scfg.Config
}

func NewSheetsIterator(ctx context.Context, config scfg.Config,
	client *http.Client, tp *position.SheetPosition) (*SheetsIterator, error) {
	tmbWithCtx, ctx := tomb.WithContext(ctx)

	cdc := &SheetsIterator{
		client:           client,
		lastModifiedTime: tp.NextRun,
		rowOffset:        tp.RowOffset,
		tomb:             tmbWithCtx,
		cfg:              config,
		caches:           make(chan []sdk.Record, 1),
		buffer:           make(chan sdk.Record, 1),
		ticker:           time.NewTicker(config.IterationInterval),
	}

	cdc.tomb.Go(cdc.startIterator(ctx))
	cdc.tomb.Go(cdc.flush)

	return cdc, nil
}

func (c *SheetsIterator) startIterator(ctx context.Context) func() error {
	sdk.Logger(ctx).Info().Msg("++++++++starting iterator+++++++++++++")
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
					c.lastModifiedTime = pos.NextRun
				case <-c.tomb.Dying():
					return c.tomb.Err()
				}
			}
		}
	}
}

func (c *SheetsIterator) flush() error {
	defer close(c.buffer)
	for {
		select {
		case <-c.tomb.Dying():
			return c.tomb.Err()
		case cache := <-c.caches:
			for _, record := range cache {
				c.buffer <- record
			}
		}
	}
}

func (c *SheetsIterator) HasNext(_ context.Context) bool {
	return len(c.buffer) > 0 || !c.tomb.Alive()
}

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

func (c *SheetsIterator) Stop() {
	c.ticker.Stop()
	c.tomb.Kill(errors.New("iterator stopped"))
}

func (c *SheetsIterator) getSheetRecords(ctx context.Context) ([]sdk.Record, error) {
	if c.nextRun.After(time.Now()) {
		return nil, nil
	}

	sheetService, err := sheets.NewService(ctx, option.WithHTTPClient(c.client))
	if err != nil {
		return nil, err
	}

	var s sheets.DataFilter
	dataFilters := []*sheets.DataFilter{}
	s.GridRange = &sheets.GridRange{
		SheetId:       c.cfg.GoogleSheetID,
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

	res, err := sheetService.Spreadsheets.Values.BatchGetByDataFilter(c.cfg.GoogleSpreadsheetID, rbt).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	valueRange := res.ValueRanges[0].ValueRange
	if (res.HTTPStatusCode != http.StatusOK) || res == nil {
		c.nextRun = time.Now().Add(c.cfg.IterationInterval)
		return nil, nil
	}

	responseData := valueRange.Values
	for index, value := range responseData {
		if len(value) == 0 {
			responseData = responseData[:index]
			break
		}
	}

	records := make([]sdk.Record, 0, len(responseData))
	for index, val := range responseData {
		rawData, err := json.Marshal(val)
		if err != nil {
			return records, fmt.Errorf("error marshaling the map: %w", err)
		}
		lastRowPosition := position.SheetPosition{
			RowOffset: int64(index + 1),
			NextRun:   time.Now(),
		}

		c.rowOffset = int64(index)
		records = append(records, sdk.Record{
			Position:  lastRowPosition.RecordPosition(),
			Metadata:  nil,
			CreatedAt: time.Now(),
			Key:       sdk.RawData([]byte(fmt.Sprintf("A%d", index+1))),
			Payload:   sdk.RawData(rawData),
		})
	}
	return records, nil
}
