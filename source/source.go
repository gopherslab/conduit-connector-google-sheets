package source

import (
	"context"
	"fmt"
	"net/http"

	"github.com/conduitio/conduit-connector-google-sheets/source/iterator"
	"github.com/conduitio/conduit-connector-google-sheets/source/position"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"golang.org/x/oauth2"
)

type Source struct {
	sdk.UnimplementedSource

	client     *http.Client
	iterator   Iterator
	configData Config
}

type Iterator interface {
	HasNext(ctx context.Context) bool
	Next(ctx context.Context) (sdk.Record, error)
	Stop()
}

func NewSource() sdk.Source {
	return &Source{}
}

func (s *Source) Configure(ctx context.Context, cfg map[string]string) error {
	sheetsConfig, err := Parse(cfg) // config.Parse(cfg)
	if err != nil {
		return err
	}

	// s.configData = config.Config{
	// 	GoogleSpreadsheetID: sheetsConfig.GoogleSpreadsheetID,
	// 	GoogleSheetID:       sheetsConfig.GoogleSheetID,
	// 	IterationInterval:   sheetsConfig.IterationInterval,
	// }

	s.configData = Config{
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
	record, err := position.ParseRecordPosition(rp)
	if err != nil {
		return fmt.Errorf("couldn't parse position: %w", err)
	}

	s.iterator, err = iterator.NewCDCIterator(ctx, s.client,
		s.configData.GoogleSpreadsheetID,
		s.configData.GoogleSheetID,
		s.configData.IterationInterval,
		record,
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

func (s *Source) Teardown(ctx context.Context) error {
	if s.iterator != nil {
		s.iterator = nil
	}
	return nil
}

func (s *Source) Ack(ctx context.Context, position sdk.Position) error {
	return nil
}
