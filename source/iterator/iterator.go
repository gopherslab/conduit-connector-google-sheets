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
	"errors"
	"fmt"
	"time"

	"github.com/conduitio/conduit-connector-google-sheets/sheets"
	"github.com/conduitio/conduit-connector-google-sheets/source/position"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"gopkg.in/tomb.v2"
)

type SheetsIterator struct {
	sheets    *sheets.BatchReader
	rowOffset int64
	tomb      *tomb.Tomb
	ticker    *time.Ticker
	caches    chan []sdk.Record
	buffer    chan sdk.Record
}

// NewSheetsIterator creates a new instance of sheets iterator and starts polling google sheets api for new changes
// using the row offset of last successful row read in a separate go routine, row offset is received in sheet position
func NewSheetsIterator(ctx context.Context,
	tp position.SheetPosition,
	args sheets.BatchReaderArgs,
) (*SheetsIterator, error) {
	tmbWithCtx, _ := tomb.WithContext(ctx)
	sheets, err := sheets.NewBatchReader(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("error initializing sheets BatchReader: %w", err)
	}

	cdc := &SheetsIterator{
		sheets:    sheets,
		rowOffset: tp.RowOffset,
		tomb:      tmbWithCtx,
		caches:    make(chan []sdk.Record, 1),
		buffer:    make(chan sdk.Record, 1),
		ticker:    time.NewTicker(args.PollingPeriod),
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
				records, err := c.sheets.GetSheetRecords(ctx, c.rowOffset)
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
func (c *SheetsIterator) HasNext() bool {
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
