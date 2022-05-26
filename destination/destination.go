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

package destination

import (
	"context"
	"fmt"
	"sync"

	"github.com/conduitio/conduit-connector-google-sheets/sheets"
	sdk "github.com/conduitio/conduit-connector-sdk"
)

// Destination connector
type Destination struct {
	sdk.UnimplementedDestination
	// buffer holds the records for asynchronous write to google sheets
	buffer []sdk.Record
	// ackCache holds the ack functions for the records currently buffered
	// i-th index of ackCache is the ack function of record buffered at i-th index
	ackCache []sdk.AckFunc
	// err holds the last error encountered by the connector
	err error
	// config holds the destination config
	config Config
	// writer is the instance of sheets writer, which is a wrapper over sheets write API
	writer *sheets.Writer

	mux *sync.Mutex
}

func NewDestination() sdk.Destination {
	return &Destination{}
}

// Configure parses and initializes the config.
func (d *Destination) Configure(ctx context.Context,
	cfg map[string]string) error {
	sheetsConfig, err := Parse(cfg)
	if err != nil {
		return fmt.Errorf("failed parsing the config: %w", err)
	}

	d.config = Config{
		Config:           sheetsConfig.Config,
		SheetName:        sheetsConfig.SheetName,
		BufferSize:       sheetsConfig.BufferSize,
		ValueInputOption: sheetsConfig.ValueInputOption,
	}
	d.mux = &sync.Mutex{}
	return nil
}

// Open makes sure everything is prepared to receive records.
func (d *Destination) Open(ctx context.Context) error {
	// initializing the buffer
	d.buffer = make([]sdk.Record, 0, d.config.BufferSize)
	d.ackCache = make([]sdk.AckFunc, 0, d.config.BufferSize)

	writer, err := sheets.NewWriter(
		ctx,
		d.config.OAuthConfig,
		d.config.OAuthToken,
		d.config.GoogleSpreadsheetID,
		d.config.SheetName,
		d.config.ValueInputOption,
		d.config.MaxRetries,
	)
	if err != nil {
		return fmt.Errorf("unable to init writer: %w", err)
	}
	d.writer = writer
	return nil
}

// WriteAsync writes a record into a Destination. Typically, Destination maintains an in-memory
// buffer and doesn't actually perform write until the buffer has enough
// records in it. This is done for performance reasons.
func (d *Destination) WriteAsync(ctx context.Context,
	r sdk.Record, ack sdk.AckFunc) error {
	// If either Destination or Writer have encountered an error, there's no point in
	// accepting more records. We better signal the error up the stack and force
	// the server to maybe re-instantiate plugin or do something else about it.
	if d.err != nil {
		return d.err
	}

	if len(r.Payload.Bytes()) == 0 {
		return nil
	}

	d.mux.Lock()
	defer d.mux.Unlock()

	d.buffer = append(d.buffer, r)
	d.ackCache = append(d.ackCache, ack)

	if len(d.buffer) >= int(d.config.BufferSize) {
		err := d.Flush(ctx)
		if err != nil {
			return fmt.Errorf("failed flushing the records: %w", err)
		}
	}

	return d.err
}

// Flush writes the records when the buffer threshold is hit and after successful pushing the data
// empties the record buffer and acknowledgment buffer for new records.
func (d *Destination) Flush(ctx context.Context) error {
	bufferedRecords := d.buffer
	d.buffer = d.buffer[:0]

	err := d.writer.Write(ctx, bufferedRecords)
	if err != nil {
		d.err = err
	}

	// call all the written records ackFunctions
	for _, ack := range d.ackCache {
		err := ack(d.err)
		if err != nil {
			return fmt.Errorf("failed acknowledgement: %w", err)
		}
	}
	d.ackCache = d.ackCache[:0]

	return nil
}

// Teardown writes all the pending records to sheets and gracefully disconnects the client
func (d *Destination) Teardown(ctx context.Context) error {
	defer func() {
		d.writer = nil
	}()
	if d.mux != nil {
		d.mux.Lock()
		defer d.mux.Unlock()
		return d.Flush(ctx)
	}
	return nil
}
