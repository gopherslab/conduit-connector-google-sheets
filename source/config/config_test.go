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
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/conduitio/conduit-connector-google-sheets/config"
	"github.com/stretchr/testify/assert"
)

type sourceTestCase []struct {
	testCase string
	params   map[string]string
	expected Config
	err      error
}

func TestParse(t *testing.T) {
	// validCredFile := "../testdata/dummy_cred.json" //#nosec // nolint: gosec // not valid creds
	filePath := getFilePath("conduit-connector-google-sheets")
	validCredFile := fmt.Sprintf("%s/testdata/dummy_cred.json", filePath)

	cases := sourceTestCase{
		{
			testCase: "Checking against default values",
			params:   map[string]string{},
			err:      fmt.Errorf("\"google.credentialsFile\" config value must be set"),
			expected: Config{},
		},
		{
			testCase: "Checking if credentialsFile parameter is empty",
			params: map[string]string{
				config.KeyTokensFile:      validCredFile,
				config.KeyCredentialsFile: "",
				config.KeySheetURL:        "",
				KeyPollingPeriod:          "",
			},
			err:      fmt.Errorf("\"google.credentialsFile\" config value must be set"),
			expected: Config{},
		},
		{
			testCase: "Checking if tokensFile parameter is empty",
			params: map[string]string{
				config.KeyTokensFile:      "",
				config.KeyCredentialsFile: validCredFile,
				config.KeySheetURL:        "",
				KeyPollingPeriod:          "",
			},
			err:      fmt.Errorf("\"google.tokensFile\" config value must be set"),
			expected: Config{},
		},
		{
			testCase: "Checking if sheetsURL parameter is empty",
			params: map[string]string{
				config.KeyTokensFile:      validCredFile,
				config.KeyCredentialsFile: validCredFile,
				config.KeySheetURL:        "",
			},
			err:      fmt.Errorf("\"google.sheetsURL\" config value must be set"),
			expected: Config{},
		},
		{
			testCase: "Checking if pollingPeriod parameter is in non-acceptable format",
			params: map[string]string{
				config.KeyTokensFile:      validCredFile,
				config.KeyCredentialsFile: validCredFile,
				config.KeySheetURL:        "https://docs.google.com/spreadsheets/d/19VVe4M-j8MGw-a3B7fcJQnx5JnHjiHf9dwChUkqQ4/edit#gid=158080911",
				KeyPollingPeriod:          "minute",
			},
			err:      fmt.Errorf("\"minute\" cannot parse interval to time duration"),
			expected: Config{},
		},
		{
			testCase: "Checking if pollingPeriod parameter is empty",
			params: map[string]string{
				config.KeyTokensFile:      validCredFile,
				config.KeyCredentialsFile: validCredFile,
				config.KeySheetURL:        "https://docs.google.com/spreadsheets/d/19VVe4M-j8MGw-a3B7fcJQnx5JnHjiHf9dwChUkqQ4/edit#gid=158080911",
			},
			err: nil,
			expected: Config{
				Config: config.Config{
					Client:              nil,
					GoogleSpreadsheetID: "19VVe4M-j8MGw-a3B7fcJQnx5JnHjiHf9dwChUkqQ4",
					GoogleSheetID:       158080911,
				},
				PollingPeriod: 6 * time.Second,
			},
		},
		{
			testCase: "Checking for ideal case",
			params: map[string]string{
				config.KeyTokensFile:      validCredFile,
				config.KeyCredentialsFile: validCredFile,
				config.KeySheetURL:        "https://docs.google.com/spreadsheets/d/19VVe4M-j8MGw-a3B7fcJQnx5JnHjiHf9dwChUkqQ4/edit#gid=158080911",
				KeyPollingPeriod:          "2m",
			},
			err: nil,
			expected: Config{
				Config: config.Config{
					Client:              nil,
					GoogleSpreadsheetID: "19VVe4M-j8MGw-a3B7fcJQnx5JnHjiHf9dwChUkqQ4",
					GoogleSheetID:       158080911,
				},
				PollingPeriod: 2 * time.Minute,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testCase, func(t *testing.T) {
			cfg, err := Parse(tc.params)
			if err != nil {
				assert.NotNil(t, err)
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
