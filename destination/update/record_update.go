package update

import (
	"context"
	"fmt"
	"net/http"

	sheets "google.golang.org/api/sheets/v4"
)

// type UpdateRecordDetails struct {
// 	spreadsheetID string
// 	sheetRange    string
// 	inputOption   string
// }

func UpdateSpreadsheetRecord(ctx context.Context, service *sheets.Service, client *http.Client, gSheetID string, sheetRange string, inputValue string, data []byte) error {
	// service, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	// if err != nil {
	// 	return err
	// }

	// The ID of the spreadsheet to update.
	// spreadsheetId := "1gQjm4hnSdrMFyPjhlwSGLBbj0ACOxFQJpVST1LmW6Hg" // TODO: Update placeholder value.

	// The A1 notation of the values to update.
	// range2 := "Sheet1!A:A9" // TODO: Update placeholder value.

	// How the input data should be interpreted.
	// valueInputOption := "USER_ENTERED" // TODO: Update placeholder value.

	// updateData := [][]interface{}{
	// 	{"Yuvi"},
	// }

	var updateData = [][]interface{}{}
	updateData = append(updateData, []interface{}{data})

	rb := &sheets.ValueRange{
		MajorDimension: "ROWS",
		Range:          sheetRange,
		Values:         updateData,
	}

	resp, err := service.Spreadsheets.Values.Update(gSheetID,
		sheetRange, rb).ValueInputOption(inputValue).Context(ctx).Do()
	if err != nil {
		return err
	}

	// TODO: Change code below to process the `resp` object:
	fmt.Printf("%#v\n", resp)

	return nil
}
