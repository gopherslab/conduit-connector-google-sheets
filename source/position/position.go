/*
Copyright © 2022 Meroxa, Inc. & Gophers Lab Technologies Pvt. Ltd.

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

package position

import (
	"encoding/json"
	"fmt"

	sdk "github.com/conduitio/conduit-connector-sdk"
)

type SheetPosition struct {
	RowOffset     int64  `json:"row_offset"`
	SpreadsheetID string `json:"spreadsheet_id"`
	SheetID       int64  `json:"sheet_id"`
}

// ParseRecordPosition is used to parse the sdk.Position to SheetPosition type
func ParseRecordPosition(p sdk.Position) (SheetPosition, error) {
	var recordPosition SheetPosition

	if p == nil {
		return SheetPosition{}, nil
	}

	if err := json.Unmarshal(p, &recordPosition); err != nil {
		return SheetPosition{}, fmt.Errorf("could not parse the position timestamp: %w", err)
	}
	return recordPosition, nil
}

// RecordPosition converts the SheetPosition to sdk.Position to be returned in sdk.Record
func (s SheetPosition) RecordPosition() sdk.Position {
	pos, err := json.Marshal(s)
	if err != nil {
		return sdk.Position{}
	}
	return pos
}
