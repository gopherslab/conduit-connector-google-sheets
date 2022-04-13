package iterator

import (
	"context"
	"fmt"
	"net/http"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type CDCIterator struct {
	service     *sheets.Service
	sheetsValue *sheets.ValueRange
}

var counter int

func NewCDCIterator(ctx context.Context, client *http.Client, spreadsheetId string, sheetRange string) (*CDCIterator, error) {
	var err error
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	sheetData, err := srv.Spreadsheets.Values.Get(spreadsheetId, sheetRange).Do()
	if err != nil {
		return nil, err
	}

	counter = 0
	counter = len(sheetData.Values)

	fmt.Println("Data from GSA: ", sheetData)

	c := &CDCIterator{
		service:     srv,
		sheetsValue: sheetData,
	}

	return c, nil

}

func (i *CDCIterator) HasNext(ctx context.Context) bool {
	if counter > 0 {
		counter -= 1
		return true

	}

	return false
}

func (i *CDCIterator) Next(ctx context.Context) (sdk.Record, error) {
	// read object
	rawData, err := i.sheetsValue.MarshalJSON()
	if err != nil {
		return sdk.Record{}, fmt.Errorf("could not read the object's body: %w", err)
	}

	// create the record
	output := sdk.Record{
		Metadata: map[string]string{
			"range":     i.sheetsValue.Range,
			"dimension": i.sheetsValue.MajorDimension,
		},
		// Position:  p.ToRecordPosition(),
		Payload: sdk.RawData(rawData),
		// Key:       sdk.RawData(*key),
		// CreatedAt: *object.LastModified,
	}

	return output, nil

}

func (i *CDCIterator) Stop() {
	// under development
}
