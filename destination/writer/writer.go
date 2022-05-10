/*
Copyright © 2022 Meroxa, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package writer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	dcfg "github.com/conduitio/conduit-connector-google-sheets/destination/destinationconfig"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"google.golang.org/api/option"
	sheets "google.golang.org/api/sheets/v4"
)

func Writer(ctx context.Context, record []sdk.Record, cfg dcfg.Config, client *http.Client) error {
	var dataFormat [][]interface{}

	for _, dataValue := range record {
		rowArr := make([]interface{}, 0)
		err := json.Unmarshal(dataValue.Payload.Bytes(), &rowArr)
		if err != nil {
			return fmt.Errorf("unable to marshal the record %w", err)
		}
		dataFormat = append(dataFormat, rowArr)
	}

	sheetService, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("unable to create google-sheet service %w", err)
	}

	sheetValueFormat := &sheets.ValueRange{
		MajorDimension: "ROWS",
		Range:          cfg.SheetRange,
		Values:         dataFormat,
	}

	_, err = sheetService.Spreadsheets.Values.Append(
		cfg.GoogleSpreadsheetID, cfg.SheetRange,
		sheetValueFormat).ValueInputOption(
		cfg.ValueInputOption).InsertDataOption(
		cfg.InsertDataOption).Context(ctx).Do()

	if err != nil {
		return fmt.Errorf("error pushing records to google-sheets %w", err)
	}
	return nil
}
