package source

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/conduitio/conduit-connector-google-sheets/config"
	sdk "github.com/conduitio/conduit-connector-sdk"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Source struct {
	sdk.UnimplementedSource

	client     *http.Client
	service    *sheets.Service
	configData config.Config
}

func NewSource() sdk.Source {
	return &Source{}
}

func (s *Source) Configure(ctx context.Context, cfg map[string]string) error {
	config2, err := config.Parse(cfg)
	if err != nil {
		return err
	}

	s.configData = config.Config{
		GoogleSpreadsheetId:    config2.GoogleSpreadsheetId,
		GoogleSpreadsheetRange: config2.GoogleSpreadsheetRange,
	}

	token := &oauth2.Token{
		AccessToken:  config2.GoogleAccessToken,
		TokenType:    "Bearer",
		RefreshToken: config2.RefreshToken,
	}

	var authCfg *oauth2.Config
	s.client = authCfg.Client(ctx, token)

	return nil
}

// Open prepare the plugin to start sending records from the given position
func (s *Source) Open(ctx context.Context, rp sdk.Position) error {
	service, err := sheets.NewService(ctx, option.WithHTTPClient(s.client))
	if err != nil {
		return err
	}

	s.service = service
	return nil
}

// Read gets the next object
func (s *Source) Read(ctx context.Context) (sdk.Record, error) {
	r := sdk.Record{}
	sheetData, err := s.service.Spreadsheets.Values.Get(s.configData.GoogleSpreadsheetId, s.configData.GoogleSpreadsheetRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
		return r, err
	}

	if len(sheetData.Values) == 0 {
		log.Fatalf("spreadheet has no value:%v", err)
		return r, err
	} else {
		for _, row := range sheetData.Values {
			fmt.Printf("%s\n", row)
		}
	}

	return r, nil
}

func (s *Source) Teardown(ctx context.Context) error {
	return nil
}

func (s *Source) Ack(ctx context.Context, position sdk.Position) error {
	return nil
}
