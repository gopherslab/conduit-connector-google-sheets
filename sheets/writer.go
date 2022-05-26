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

package sheets

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const insertDataOption = "INSERT_ROWS"

type Writer struct {
	sheets           *sheets.Service
	sheetName        string
	spreadsheetID    string
	valueInputOption string
	maxRetries       uint64
	retryCount       uint64
}

func NewWriter(ctx context.Context, client *http.Client, spreadsheetID, sheetName, valueInputOption string, retries uint64) (*Writer, error) {
	sheetService, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	return &Writer{
		spreadsheetID:    spreadsheetID,
		sheets:           sheetService,
		sheetName:        sheetName,
		valueInputOption: valueInputOption,
		maxRetries:       retries,
	}, nil
}

// Write function writes the records to google sheets
func (w *Writer) Write(ctx context.Context, records []sdk.Record) error {
	var rows [][]interface{}

	// Looping on every record and unmarhsalling to google-sheets format.
	for index, rowRecord := range records {
		rowArr := make([]interface{}, 0)
		err := json.Unmarshal(rowRecord.Payload.Bytes(), &rowArr)
		if err != nil {
			return fmt.Errorf("at index %d unable to marshal the record %w", index, err)
		}
		rows = append(rows, rowArr)
	}
	if len(rows) == 0 {
		return nil
	}
	// KeyValueInputOption is the config name for how the input data
	// should be interpreted.
	// Creating a google-sheet format to append to google-sheet
	sheetValueFormat := &sheets.ValueRange{
		MajorDimension: majorDimension,
		Range:          w.sheetName,
		Values:         rows,
	}

	_, err := w.sheets.Spreadsheets.Values.Append(
		w.spreadsheetID, w.sheetName,
		sheetValueFormat).ValueInputOption(
		w.valueInputOption).InsertDataOption(
		insertDataOption).Context(ctx).Do()

	if err != nil {
		// retry mechanism, in case of rate limit exceeded error (429)
		if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == http.StatusTooManyRequests {
			if w.retryCount >= w.maxRetries {
				return fmt.Errorf("rate limit exceeded, retries: %d, error: %w", w.retryCount, err)
			}
			w.retryCount++
			// if retry count doesn't exceed maxRetries, retry with exponential back off
			// block till write either succeeds or all retries are exhausted
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(w.retryCount) * time.Second): // exponential back off
				return w.Write(ctx, records)
			}
		}
		return err
	}
	return nil
}
