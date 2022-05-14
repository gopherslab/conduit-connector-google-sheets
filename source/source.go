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

package source

import (
	"context"
	"fmt"
	"net/http"

	sc "github.com/conduitio/conduit-connector-google-sheets/source/config"
	"github.com/conduitio/conduit-connector-google-sheets/source/iterator"
	"github.com/conduitio/conduit-connector-google-sheets/source/position"

	sdk "github.com/conduitio/conduit-connector-sdk"
)

type Source struct {
	sdk.UnimplementedSource

	client     *http.Client
	iterator   Iterator
	configData sc.Config
}

type Iterator interface {
	HasNext(ctx context.Context) bool
	Next(ctx context.Context) (sdk.Record, error)
	Stop()
}

func NewSource() sdk.Source {
	return &Source{}
}

// Configure validates the passed config and prepares the source connector
func (s *Source) Configure(ctx context.Context, cfg map[string]string) error {
	sheetsConfig, err := sc.Parse(cfg)
	if err != nil {
		return err
	}
	s.configData = sheetsConfig

	s.client = s.configData.Client
	return nil
}

// Open prepare the plugin to start sending records from the given position
func (s *Source) Open(ctx context.Context, rp sdk.Position) error {
	pos, err := position.ParseRecordPosition(rp)
	if err != nil {
		return fmt.Errorf("couldn't parse position: %w", err)
	}

	s.iterator, err = iterator.NewSheetsIterator(ctx, s.client, pos,
		s.configData.GoogleSpreadsheetID,
		s.configData.GoogleSheetID,
		s.configData.PollingPeriod,
	)

	if err != nil {
		return fmt.Errorf("couldn't create a iterator: %w", err)
	}
	return nil
}

// Read gets the next object
func (s *Source) Read(ctx context.Context) (sdk.Record, error) {
	if !s.iterator.HasNext(ctx) {
		return sdk.Record{}, sdk.ErrBackoffRetry
	}

	r, err := s.iterator.Next(ctx)
	if err != nil {
		return sdk.Record{}, err
	}
	return r, nil
}

// Teardown is called by the conduit server to stop the source connector
// all the cleanup should be done in this function
func (s *Source) Teardown(_ context.Context) error {
	if s.iterator != nil {
		s.iterator.Stop()
		s.iterator = nil
	}
	return nil
}

// Ack is called by the conduit server after the record has been successfully processed by all destination connectors
func (s *Source) Ack(ctx context.Context, tp sdk.Position) error {
	pos, err := position.ParseRecordPosition(tp)
	if err != nil {
		sdk.Logger(ctx).Error().Err(err).Msg("invalid position received")
	}
	sdk.Logger(ctx).Info().Int64("row_offset", pos.RowOffset).Msg("message ack received")
	return nil
}
