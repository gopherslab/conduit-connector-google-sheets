package insert

import (
	"context"
	"fmt"
	"log"

	sheets "google.golang.org/api/sheets/v4"
)

type AppendRecordDetails struct {
	spreadsheetID string
	sheetRange    string
	inputOption   string
}

func AddSpreadsheetRecord(srv *sheets.Service, ctx context.Context) {
	spreadsheetId := "1gQjm4hnSdrMFyPjhlwSGLBbj0ACOxFQJpVST1LmW6Hg" // TODO: Update placeholder value.

	// The A1 notation of a range to search for a logical table of data.
	// Values will be appended after the last row of the table.
	range2 := "Sheet2" // TODO: Update placeholder value.

	// How the input data should be interpreted.
	valueInputOption := "USER_ENTERED" // TODO: Update placeholder value.

	// How the input data should be inserted.
	insertDataOption := "INSERT_ROWS" // TODO: Update placeholder value.

	Data := [][]interface{}{
		{"", "", "", "24", "Probation", "Fresher"},
	}

	rb := &sheets.ValueRange{
		MajorDimension: "ROWS",
		Range:          range2,
		Values:         Data,
	}

	resp, err := srv.Spreadsheets.Values.Append(spreadsheetId, 
		range2, rb).ValueInputOption(valueInputOption).InsertDataOption(insertDataOption).Context(ctx).Do()
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Change code below to process the `resp` object:
	fmt.Printf("%#v\n", resp)
}
