package destination

// import (
// 	"context"
// 	"testing"

// 	sdk "github.com/conduitio/conduit-connector-sdk"
// )

// func TestDstination(t *testing.T) {
// 	cfg := map[string]string{
// 		"access_token":       "access_token_value",
// 		"refresh_token":      "refresh_token_value",
// 		"spreadsheet_id":     "google_spreadsheet_id",
// 		"sheet_range":        "Sheet2",
// 		"value_input_option": "USER_ENTERED",
// 		"insert_data_option": "INSERT_ROWS",
// 	}

// 	ctx := context.Background()
// 	dest := &Destination{}
// 	err := dest.Configure(ctx, cfg)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	err = dest.Open(ctx)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	records := []sdk.Record{
// 		{}, {},
// 	}
// 	ackFunc := []sdk.AckFunc{}
// 	for i, v := range records {
// 		dest.WriteAsync(ctx, v, ackFunc[i])
// 	}

// 	err = dest.Flush(ctx)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// }