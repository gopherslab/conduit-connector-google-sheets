package destination

import (
	"context"
	"fmt"
	"net/http"

	"github.com/conduitio/conduit-connector-google-sheets/destination/model"
	u "github.com/conduitio/conduit-connector-google-sheets/destination/update"

	"google.golang.org/api/option"
	sheets "google.golang.org/api/sheets/v4"
	sdk "github.com/conduitio/conduit-connector-sdk"
	"golang.org/x/oauth2"
)

type Destination struct {
	sdk.UnimplementedDestination

	Config
	client   *http.Client
	service  *sheets.Service
	toUpdate bool
	object   model.SheetObject
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
	d.object = model.SheetObject{
		SpreadsheetID: destinationConfig.GoogleSpreadsheetID,
		SheetRange:    destinationConfig.SheetRange,
		InputOption:   destinationConfig.ValueInputOption,
	}

	fmt.Printf("*************+===========%#v\n\n======:", d.object)

	return nil
}

// Open makes sure everything is prepared to receive records.
func (d *Destination) Open(ctx context.Context) error {
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
	fmt.Println("*******************", d.object.SpreadsheetID, d.object.SheetRange, d.object.InputOption)

	err := u.UpdateSpreadsheetRecord(ctx, d.service, d.object, r.Payload.Bytes())
	if err != nil {
		return err
	}

	return nil
}

// Teardown gracefully disconnects the client
func (d *Destination) Teardown(ctx context.Context) error {
	fmt.Println("************************" + "Inside Teardown method" + "**************************")
	d.service = nil
	d.client = nil

	return nil // TODO
}
