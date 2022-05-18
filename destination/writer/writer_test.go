/* Copyright © 2022 Meroxa, Inc.

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
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/conduitio/conduit-connector-google-sheets/config"
	destConfig "github.com/conduitio/conduit-connector-google-sheets/destination/config"
	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/stretchr/testify/assert"
)

type writerTest []struct {
	testName string
	r        []sdk.Record
	cfg      destConfig.Config
	err      error
	expected [][]interface{}
}

func TestWriter(t *testing.T) {
	cases := writerTest{
		{
			testName: "Empty Record Payload",
			r:        []sdk.Record{},
			cfg: destConfig.Config{
				Config: config.Config{
					GoogleSpreadsheetID: "123abcd",
				},
				SheetName:        "Sheet",
				InsertDataOption: "INSERT_ROW",
			},
			err:      fmt.Errorf("error pushing records to google-sheets"),
			expected: [][]interface{}{},
		},
		{
			testName: "Non-Empty Record Payload",
			r: []sdk.Record{
				{
					Position:  []byte(``),
					Metadata:  nil,
					CreatedAt: time.Time{},
					Key:       sdk.RawData{0x41, 0x31, 0x33, 0x37},
					Payload: sdk.RawData{0x5b, 0x22, 0x4f, 0x76, 0x65,
						0x72, 0x77, 0x72, 0x69, 0x74, 0x65, 0x20, 0x64,
						0x61, 0x74, 0x61, 0x31, 0x22, 0x2c, 0x22, 0x4f,
						0x76, 0x65, 0x72, 0x77, 0x72, 0x69, 0x74, 0x65,
						0x20, 0x64, 0x61, 0x74, 0x61, 0x32, 0x22, 0x2c,
						0x22, 0x4f, 0x76, 0x65, 0x72, 0x77, 0x72, 0x69,
						0x74, 0x65, 0x20, 0x64, 0x61, 0x74, 0x61, 0x33,
						0x22, 0x2c, 0x22, 0x4f, 0x76, 0x65, 0x72, 0x77,
						0x72, 0x69, 0x74, 0x65, 0x20, 0x64, 0x61, 0x74,
						0x61, 0x34, 0x22, 0x2c, 0x22, 0x4f, 0x76, 0x65,
						0x72, 0x77, 0x72, 0x69, 0x74, 0x65, 0x20, 0x64,
						0x61, 0x74, 0x61, 0x35, 0x22, 0x2c, 0x22, 0x4f,
						0x76, 0x65, 0x72, 0x77, 0x72, 0x69, 0x74, 0x65,
						0x20, 0x64, 0x61, 0x74, 0x61, 0x36, 0x22, 0x5d},
				},
			},
			cfg: destConfig.Config{
				Config: config.Config{
					Client:              nil,
					GoogleSpreadsheetID: "19VVe4M-j8MGw-a3B7fcJQnx5JnHjiHf9dwChUkqQ4",
					GoogleSheetID:       158080911,
				},
				SheetName:        "Sheet",
				InsertDataOption: "INSERT_ROW",
			},
			err:      fmt.Errorf("error pushing records to google-sheets"),
			expected: [][]interface{}{},
		},
	}

	ctx := context.Background()
	for _, tc := range cases {
		t.Run(tc.testName, func(t *testing.T) {
			err := Writer(ctx, tc.r, tc.cfg, &http.Client{})
			if err != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tc.err.Error())
			}
		})
	}
}
