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
package destination

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/conduitio/conduit-connector-google-sheets/config"
	"github.com/stretchr/testify/assert"
)

type destTestCase []struct {
	testCase string
	params   map[string]string
	err      error
	expected Config
}

func TestParse(t *testing.T) {
	filePath := getFilePath("conduit-connector-google-sheets")
	validCredFile := fmt.Sprintf("%s/testdata/dummy_cred.json", filePath)

	cases := destTestCase{
		{
			testCase: "Checking against default values",
			params:   map[string]string{},
			err:      fmt.Errorf("\"google.credentialsFile\" config value must be set"),
			expected: Config{},
		},
		{
			testCase: "Checking against if any required value is empty",
			params: map[string]string{
				config.KeyTokensFile:      validCredFile,
				config.KeyCredentialsFile: validCredFile,
				config.KeySheetURL:        "",
				KeySheetName:              "Sheet",
				KeyValueInputOption:       "",
				KeyBufferSize:             "",
			},
			err:      fmt.Errorf("\"google.sheetsURL\" config value must be set"),
			expected: Config{},
		},
		{
			testCase: "Checking against random values case",
			params: map[string]string{
				config.KeyTokensFile:      validCredFile,
				config.KeyCredentialsFile: validCredFile,
				config.KeySheetURL:        "https://docs.google.com/spreadsheets/d/19VVe4M-j8MGw-a3B7fcJQnx5JnHjiHf9dwChUkqQ4/edit#gid=158080911",
				KeySheetName:              "",
				KeyValueInputOption:       "",
				KeyBufferSize:             "10",
			},
			err:      fmt.Errorf("\"google.sheetName\" config value must be set"),
			expected: Config{},
		},
		{
			testCase: "Checking for IDEAL case - 1",
			params: map[string]string{
				config.KeyTokensFile:      validCredFile,
				config.KeyCredentialsFile: validCredFile,
				config.KeySheetURL:        "https://docs.google.com/spreadsheets/d/19VVe4M-j8MGw-a3B7fcJQnx5JnHjiHf9dwChUkqQ4/edit#gid=158080911",
				KeySheetName:              "Sheet",
				KeyValueInputOption:       "",
				KeyBufferSize:             "",
			},
			err: nil,
			expected: Config{
				Config: config.Config{
					Client:              nil,
					GoogleSpreadsheetID: "19VVe4M-j8MGw-a3B7fcJQnx5JnHjiHf9dwChUkqQ4",
					GoogleSheetID:       158080911,
				},
				SheetName:        "Sheet",
				ValueInputOption: defaultValueInputOption,
				BufferSize:       100,
				MaxRetries:       3,
			},
		},
		{
			testCase: "Checking for IDEAL case - 2",
			params: map[string]string{
				config.KeyTokensFile:      validCredFile,
				config.KeyCredentialsFile: validCredFile,
				config.KeySheetURL:        "https://docs.google.com/spreadsheets/d/19VVe4M-j8MGw-a3B7fcJQnx5JnHjiHf9dwChUkqQ4/edit#gid=158080911",
				KeySheetName:              "Sheet",
				KeyValueInputOption:       "RAW",
				KeyMaxRetries:             "5",
			},
			err: nil,
			expected: Config{
				Config: config.Config{
					Client:              nil,
					GoogleSpreadsheetID: "19VVe4M-j8MGw-a3B7fcJQnx5JnHjiHf9dwChUkqQ4",
					GoogleSheetID:       158080911,
				},
				SheetName:        "Sheet",
				ValueInputOption: "RAW",
				BufferSize:       100,
				MaxRetries:       5,
			},
		},
		{
			testCase: "Checking for IDEAL case - 3",
			params: map[string]string{
				config.KeyTokensFile:      validCredFile,
				config.KeyCredentialsFile: validCredFile,
				config.KeySheetURL:        "https://docs.google.com/spreadsheets/d/19VVe4M-j8MGw-a3B7fcJQnx5JnHjiHf9dwChUkqQ4/edit#gid=158080911",
				KeySheetName:              "Sheet",
				KeyValueInputOption:       "USER_ENTERED",
				KeyBufferSize:             "10",
			},
			err: nil,
			expected: Config{
				Config: config.Config{
					Client:              nil,
					GoogleSpreadsheetID: "19VVe4M-j8MGw-a3B7fcJQnx5JnHjiHf9dwChUkqQ4",
					GoogleSheetID:       158080911,
				},
				SheetName:        "Sheet",
				ValueInputOption: defaultValueInputOption,
				BufferSize:       10,
				MaxRetries:       3,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testCase, func(t *testing.T) {
			cfg, err := Parse(tc.params)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
				tc.expected.Client = cfg.Client
				assert.EqualValues(t, tc.expected, cfg)
			}
		})
	}
}

func getFilePath(path string) string {
	wd, _ := os.Getwd()
	for !strings.HasSuffix(wd, path) {
		wd = filepath.Dir(wd)
	}
	return wd
}
