package update

import (
	"context"
	"fmt"

	"github.com/conduitio/conduit-connector-google-sheets/destination/model"

	sheets "google.golang.org/api/sheets/v4"
)

func UpdateSpreadsheetRecord(ctx context.Context, service *sheets.Service, details model.SheetObject, data []byte) error {
	var updateData = [][]interface{}{}
	updateData = append(updateData, []interface{}{data})

	rb := &sheets.ValueRange{
		MajorDimension: "ROWS",
		Range:          details.SheetRange,
		Values:         updateData,
	}

	fmt.Println("Rb Data: ====================", rb)
	fmt.Printf("Rb Data11111: ====================: %#v", rb)

	resp, err := service.Spreadsheets.Values.Update(details.SpreadsheetID,
		details.SheetRange, rb).ValueInputOption(details.InputOption).Context(ctx).Do()
	if err != nil {
		return err
	}

	// TODO: Change code below to process the `resp` object:
	fmt.Printf("%#v\n", resp)

	return nil
}
