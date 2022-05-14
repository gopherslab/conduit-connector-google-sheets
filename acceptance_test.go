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
package googlesheets

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/conduitio/conduit-connector-google-sheets/destination"
	"github.com/conduitio/conduit-connector-google-sheets/source"
	"github.com/conduitio/conduit-connector-google-sheets/source/position"
	"google.golang.org/api/option"
	sheets "google.golang.org/api/sheets/v4"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"go.uber.org/goleak"
)

func TestAcceptance(t *testing.T){
	sourceConfig := map[string]string{
		"gsheets.accessToken":   "access_token",
		"gsheets.refreshToken":  "refresh_token",
		"gsheets.expiry":        "expiry",
		"gsheets.spreadsheetId": "spreadsheet_id",
		"sheet_id":              "sheet_id",
		"polling_period":        "6s", // Configurable polling period
	}

	destConfig := map[string]string{
		"gsheets.accessToken":   "access_token",
		"gsheets.refreshToken":  "refresh_token",
		"gsheets.expiry":        "expiry",
		"gsheets.spreadsheetId": "spreadsheet_id",
		"sheet_range":           "Sheet1",
		"insert_data_option":    "INSERT_ROW",
		"value_input_option":    "USER_ENTERED",
		"buffer_size":           "10",
	}

	sdk.AcceptanceTest(t, sdk.ConfigurableAcceptanceTestDriver{
		Config: sdk.ConfigurableAcceptanceTestDriverConfig{
			Connector: sdk.Connector{ // Note that this variable should rather be created globally in `connector.go`
				NewSpecification: Specification,
				NewSource:        source.NewSource,
				NewDestination:   destination.NewDestination,
			},
			SourceConfig:      sourceConfig,
			DestinationConfig: destConfig,
			GoleakOptions:     []goleak.Option{goleak.IgnoreCurrent()},
			Skip: []string{
				// these tests are skipped, because they need valid json of type map[string]string to work
				// whereas the code generates random string payload
				"TestSource_Open_ResumeAtPosition",
			},
		},
	})
}

func TestSource_Read_Success(t *testing.T) {
	sheetService, err := sheets.NewService(context.Background(), option.WithHTTPClient(&http.Client{}))
	if err != nil {
		fmt.Printf("error creating sheet client: %v", err)
		return
	}

	var (
		rowOffset int64
		s         sheets.DataFilter
	)
	dataFilters := make([]*sheets.DataFilter, 0)
	s.GridRange = &sheets.GridRange{
		SheetId:       0, // sheetID in int64
		StartRowIndex: rowOffset,
	}
	dataFilters = append(dataFilters, &s)
	rbt := &sheets.BatchGetValuesByDataFilterRequest{
		ValueRenderOption:    "",
		DataFilters:          dataFilters,
		DateTimeRenderOption: "FORMATTED_STRING",
	}
	res, err := sheetService.Spreadsheets.Values.BatchGetByDataFilter("spreadsheetID", rbt).Context(context.Background()).Do()
	if err != nil {
		fmt.Printf("error response from sheetsAPI: %v", err)
		return
	}
	valueRange := res.ValueRanges[0].ValueRange

	// Eliminating the empty records
	responseData := valueRange.Values
	for index, value := range responseData {
		if len(value) == 0 {
			responseData = responseData[:index]
			break
		}
	}

	// iterating over the records
	records := make([]sdk.Record, 0, len(responseData))
	for index, val := range responseData {
		rawData, err := json.Marshal(val)
		if err != nil {
			fmt.Printf("error marshaling the map: %v", err)
			return
		}
		rowOffset := rowOffset + int64(index) + 1
		lastRowPosition := position.SheetPosition{
			RowOffset: rowOffset,
		}

		// Creating every record
		records = append(records, sdk.Record{
			Position:  lastRowPosition.RecordPosition(),
			Metadata:  nil,
			CreatedAt: time.Now(),
			Key:       sdk.RawData(fmt.Sprintf("A%d", rowOffset)),
			Payload:   sdk.RawData(rawData),
		})
	}

	// All the records sent to the conduit server
	fmt.Printf("Records sent to conduit; %v", records)
}
func TestDestination_Write_Success(t *testing.T) {
	var dataFormat [][]interface{}
	record := []sdk.Record{}

	for _, dataValue := range record {
		rowArr := make([]interface{}, 0)
		err := json.Unmarshal(dataValue.Payload.Bytes(), &rowArr)
		if err != nil {
			fmt.Printf("unable to marshal the record %v", err)
			return
		}
		dataFormat = append(dataFormat, rowArr)
	}

	sheetService, err := sheets.NewService(context.Background(), option.WithHTTPClient(&http.Client{}))
	if err != nil {
		fmt.Printf("unable to create google-sheet service %v", err)
		return
	}

	sheetValueFormat := &sheets.ValueRange{
		MajorDimension: "ROWS",
		Range:          "SheetRange",
		Values:         dataFormat,
	}

	sheetResponse, err := sheetService.Spreadsheets.Values.Append(
		"GoogleSpreadsheetID", "SheetRange",
		sheetValueFormat).ValueInputOption(
		"ValueInputOption").InsertDataOption(
		"InsertDataOption").Context(context.Background()).Do()

	if err != nil {
		fmt.Printf("error pushing records to google-sheets %v", err)
		return
	}

	fmt.Printf("Response from sheets API: %#v", sheetResponse)
}
