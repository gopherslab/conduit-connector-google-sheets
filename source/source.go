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

	"github.com/conduitio/conduit-connector-google-sheets/source/iterator"
	"github.com/conduitio/conduit-connector-google-sheets/source/position"
	sc "github.com/conduitio/conduit-connector-google-sheets/source/sourceconfig"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"golang.org/x/oauth2"
)

type Source struct {
	sdk.UnimplementedSource

	// token      *oauth2.Token
	client     *http.Client
	iterator   Iterator
	configData sc.Config
}

type Iterator interface {
	HasNext(_ context.Context) bool
	Next(ctx context.Context) (sdk.Record, error)
	Stop()
}

func NewSource() sdk.Source {
	return &Source{}
}

func (s *Source) Configure(ctx context.Context, cfg map[string]string) error {
	sdk.Logger(ctx).Info().Msg("inside configure")
	sheetsConfig, err := sc.Parse(cfg)
	if err != nil {
		return err
	}

	s.configData = sc.Config{
		Config:            sheetsConfig.Config,
		GoogleSheetID:     sheetsConfig.GoogleSheetID,
		IterationInterval: sheetsConfig.IterationInterval,
	}

	token := &oauth2.Token{
		AccessToken:  sheetsConfig.GoogleAccessToken,
		TokenType:    "Bearer",
		RefreshToken: sheetsConfig.AuthRefreshToken,
	}

	var authCfg *oauth2.Config
	s.client = authCfg.Client(context.Background(), token)
	return nil
}

// Open prepare the plugin to start sending records from the given position
func (s *Source) Open(ctx context.Context, rp sdk.Position) error {
	sdk.Logger(ctx).Info().Msg("inside open method")
	record, err := position.ParseRecordPosition(rp)
	if err != nil {
		return fmt.Errorf("couldn't parse position: %w", err)
	}

	// s.iterator, err = iterator.NewCDCIterator(ctx, s.client,
	// 	s.configData.GoogleSpreadsheetID,
	// 	s.configData.GoogleSheetID,
	// 	s.configData.IterationInterval,
	// 	record,
	// )

	// s.iterator, err = iterator.NewSheetIterator(ctx, s.client, s.configData, record)
	// if err != nil {
	// 	return fmt.Errorf("couldn't create a iterator: %w", err)
	// }

	s.iterator, err = iterator.NewSheetsIterator(ctx,
		s.configData, s.client, &record)
	if err != nil {
		return fmt.Errorf("couldn't create a iterator: %w", err)
	}

	return nil
}

// Read gets the next object
func (s *Source) Read(ctx context.Context) (sdk.Record, error) {
	sdk.Logger(ctx).Info().Msg("reading data from client")
	if !s.iterator.HasNext(ctx) {
		fmt.Println("++++++++Error on line 107 of source file+++++++++++")

		return sdk.Record{}, sdk.ErrBackoffRetry
	}

	r, err := s.iterator.Next(ctx)
	if err != nil {
		return sdk.Record{}, err
	}
	return r, nil
}

func (s *Source) Teardown(ctx context.Context) error {
	sdk.Logger(ctx).Info().Msg("shutting down google-sheet client")
	if s.iterator != nil {
		s.iterator.Stop()
		s.iterator = nil
	}
	return nil
}

func (s *Source) Ack(ctx context.Context, position sdk.Position) error {
	sdk.Logger(ctx).Info().
		Str("position", string(position)).
		Msg("position ack received")
	return nil
}
