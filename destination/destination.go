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

package destination

import (
	"context"
	"net/http"
	"sync"

	dConfig "github.com/conduitio/conduit-connector-google-sheets/destination/config"
	"github.com/conduitio/conduit-connector-google-sheets/destination/writer"

	sdk "github.com/conduitio/conduit-connector-sdk"
)

// Destination connector
type Destination struct {
	sdk.UnimplementedDestination

	Buffer            []sdk.Record
	AckCache          []sdk.AckFunc
	Error             error
	Client            *http.Client
	Mutex             *sync.Mutex
	DestinationConfig dConfig.Config
}

func NewDestination() sdk.Destination {
	return &Destination{}
}

// Configure parses and initializes the config.
func (d *Destination) Configure(ctx context.Context,
	cfg map[string]string) error {
	sheetsConfig, err := dConfig.Parse(cfg)
	if err != nil {
		return err
	}

	d.DestinationConfig = dConfig.Config{
		Config:           sheetsConfig.Config,
		SheetRange:       sheetsConfig.SheetRange,
		BufferSize:       sheetsConfig.BufferSize,
		ValueInputOption: sheetsConfig.ValueInputOption,
		InsertDataOption: sheetsConfig.InsertDataOption,
	}

	return nil
}

// Open makes sure everything is prepared to receive records.
func (d *Destination) Open(ctx context.Context) error {
	d.Mutex = &sync.Mutex{}

	// initializing the buffer
	d.Buffer = make([]sdk.Record, 0, d.DestinationConfig.BufferSize)
	d.AckCache = make([]sdk.AckFunc, 0, d.DestinationConfig.BufferSize)

	d.Client = d.DestinationConfig.Client

	return nil
}

// WriteAsync writes a record into a Destination. Typically Destination maintains an in-memory
// buffer and doesn't actually perform a write until the buffer has enough
// records in it. This is done for performance reasons.
func (d *Destination) WriteAsync(ctx context.Context,
	r sdk.Record, ack sdk.AckFunc) error {
	if d.Error != nil {
		return d.Error
	}

	if len(r.Payload.Bytes()) == 0 {
		return nil
	}

	d.Mutex.Lock()
	defer d.Mutex.Unlock()

	d.Buffer = append(d.Buffer, r)
	d.AckCache = append(d.AckCache, ack)

	if len(d.Buffer) >= int(d.DestinationConfig.BufferSize) {
		err := d.Flush(ctx)
		if err != nil {
			return err
		}
	}

	return d.Error
}

// Flush writes the records when the buffer threshold is hit and after successful pushing the data 
// empties the record buffer and acknowledgment buffer for new records.
func (d *Destination) Flush(ctx context.Context) error {
	bufferedRecords := d.Buffer
	d.Buffer = d.Buffer[:0]

	err := writer.Writer(ctx, bufferedRecords,
		d.DestinationConfig, d.Client)
	if err != nil {
		d.Error = err
	}

	// call all the written records ackFunctions
	for _, ack := range d.AckCache {
		err := ack(d.Error)
		if err != nil {
			return err
		}
	}
	d.AckCache = d.AckCache[:0]

	return nil
}

// Teardown gracefully disconnects the client
func (d *Destination) Teardown(ctx context.Context) error {
	if d.Mutex != nil {
		d.Mutex.Lock()
		defer d.Mutex.Unlock()
		if err := d.Flush(ctx); err != nil {
			return err
		}
	}
	d.Client = nil
	return nil
}
