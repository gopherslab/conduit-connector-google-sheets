package destination

import (
	"context"
	"fmt"
	"net/http"

	updater "github.com/conduitio/conduit-connector-google-sheets/destination/update"

	"google.golang.org/api/option"
	sheets "google.golang.org/api/sheets/v4"

	// "github.com/conduitio/conduit-connector-google-sheets/destination/insert"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"golang.org/x/oauth2"
)

type Destination struct {
	sdk.UnimplementedDestination

	Config
	client   *http.Client
	service  *sheets.Service
	toUpdate bool
	// object   writer.RecordDetails
	obj DestObject
	// sWriter  writer.GoogleSheetWriter
}

type DestObject struct {
	spreadsheetID string
	sheetRange    string
	inputOption   string
}

func NewDestination() sdk.Destination {
	return &Destination{}
}

// Configure parses and initializes the config.
func (d *Destination) Configure(ctx context.Context, cfg map[string]string) error {
	fmt.Println("************************" + "Inside Configure method" + "**************************")

	destinationConfig, err := Parse(cfg)
	if err != nil {
		return err
	}

	token := &oauth2.Token{
		AccessToken:  destinationConfig.GoogleAccessToken,
		TokenType:    "Bearer",
		RefreshToken: destinationConfig.AuthRefreshToken,
	}

	var authCfg *oauth2.Config
	d.client = authCfg.Client(context.Background(), token)
	d.toUpdate = true
	d.obj = DestObject{
		spreadsheetID: destinationConfig.GoogleSpreadsheetID,
		sheetRange:    destinationConfig.SheetRange,
		inputOption:   destinationConfig.ValueInputOption,
	}

	fmt.Printf("*************+===========%#v\n\n======:", d.obj)

	return nil
}

// Open makes sure everything is prepared to receive records.
func (d *Destination) Open(ctx context.Context) error {
	// sheetWriter, err := writer.NewWriter(ctx, d.client, d.object)
	// if err != nil {
	// 	return err
	// }

	// d.sWriter = *sheetWriter
	fmt.Println("************************" + "Inside open method" + "**************************")
	var err error
	d.service, err = sheets.NewService(ctx, option.WithHTTPClient(d.client))
	if err != nil {
		return err
	}

	// if d.toUpdate {
	// 	updater.UpdateSpreadsheetRecord(ctx, d.client, d.object)
	// }

	// recordAdder.AddSpreadsheetRecord(, ctx)
	return nil
}

func (d *Destination) WriteAsync(ctx context.Context, r sdk.Record, ackFunc sdk.AckFunc) error {
	fmt.Println("************************" + "Inside WriteAsyn method" + "**************************")

	fmt.Printf("=================%#v\n\n=======================", r)
	fmt.Printf("=================%s\n\n=======================", r.Payload)
	fmt.Println("*******************", d.obj.spreadsheetID, d.obj.sheetRange, d.obj.inputOption)

	err := updater.UpdateSpreadsheetRecord(ctx, d.service, d.client,
		d.obj.spreadsheetID, d.obj.sheetRange, d.obj.inputOption,
		r.Payload.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (d *Destination) Write(ctx context.Context, r sdk.Record) error {
	fmt.Println("************************" + "Inside Write method" + "**************************")

	// fmt.Printf("=================%#v\n\n=======================", r)
	// fmt.Printf("=================%s\n\n=======================", r.Payload)
	// fmt.Println("*******************", d.obj.spreadsheetID, d.obj.sheetRange, d.obj.inputOption)

	// err := updater.UpdateSpreadsheetRecord(ctx, d.service, d.client,
	// 	d.obj.spreadsheetID, d.obj.sheetRange, d.obj.inputOption,
	// 	r.Payload.Bytes())
	// if err != nil {
	// 	return err
	// }

	return nil
}

// Teardown gracefully disconnects the client
func (d *Destination) Teardown(ctx context.Context) error {
	fmt.Println("************************" + "Inside Teardown method" + "**************************")

	d.client = nil
	d.service = nil
	return nil // TODO
}
