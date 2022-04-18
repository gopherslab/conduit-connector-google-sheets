package source

import (
	"context"
	"fmt"
	"net/http"

	"github.com/conduitio/conduit-connector-google-sheets/config"
	"github.com/conduitio/conduit-connector-google-sheets/source/iterator"
	"github.com/conduitio/conduit-connector-google-sheets/source/position"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"golang.org/x/oauth2"
)

type Source struct {
	sdk.UnimplementedSource

	client     *http.Client
	iterator   Iterator
	configData config.Config
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
	config2, err := config.Parse(cfg)
	if err != nil {
		return err
	}

	s.configData = config.Config{GoogleSpreadsheetId: config2.GoogleSpreadsheetId}

	token := &oauth2.Token{
		AccessToken: config2.GoogleAccessToken,
		TokenType:   "Bearer",
	}

	var authCfg *oauth2.Config
	s.client = authCfg.Client(ctx, token)

	return nil
}

// Open prepare the plugin to start sending records from the given position
func (s *Source) Open(ctx context.Context, rp sdk.Position) error {
	p, err := position.ParseRecordPosition(rp)
	if err != nil {
		return fmt.Errorf("couldn't parse position: %w", err)
	}
	// var err error
	sdk.Logger(ctx).Info().Msg("Last Position Value in Open: " + string(rp))

	s.iterator, err = iterator.NewCDCIterator(ctx, s.client, s.configData.GoogleSpreadsheetId, p.Key)
	if err != nil {
		return fmt.Errorf("couldn't create a iterator: %w", err)
	}

	return nil
}

// Read gets the next object
func (s *Source) Read(ctx context.Context) (sdk.Record, error) {
	if !s.iterator.HasNext(ctx) {
		sdk.Logger(ctx).Info().Msg("This is in hasnext error block")
		return sdk.Record{}, sdk.ErrBackoffRetry
	}

	sdk.Logger(ctx).Info().Msg("This is entering in next")
	r, err := s.iterator.Next(ctx)
	if err != nil {
		return sdk.Record{}, err
	}
	return r, nil
}

func (s *Source) Teardown(ctx context.Context) error {
	sdk.Logger(ctx).Info().Msg("This is entering in stop")
	if s.iterator != nil {
		// s.iterator.Stop()
		// s.client = nil
		s.iterator = nil
	}

	return nil
}

func (s *Source) Ack(ctx context.Context, position sdk.Position) error {
	sdk.Logger(ctx).Info().Msg("This is ack")
	// sdk.Logger(ctx).Info().Msg(fmt.Sprintf("Position: %s", string(position)))

	return nil
}
