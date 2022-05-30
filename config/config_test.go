// Copyright © 2022 Meroxa, Inc. & Gophers Lab Technologies Pvt. Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	filePath := getFilePath("conduit-connector-google-sheets") // Get project root path
	validCredFile := fmt.Sprintf("%s/testdata/dummy_cred.json", filePath)
	invalidCredFile := fmt.Sprintf("%s/testdata/dummy_invalid_cred.json", filePath)
	tests := []struct {
		name   string
		config map[string]string
		err    error
		want   Config
	}{{
		name:   "missing required params",
		config: map[string]string{},
		err:    fmt.Errorf(`"credentialsFile" config value must be set`),
		want:   Config{},
	}, {
		name: "config succeeds",
		config: map[string]string{
			KeyTokensFile:      validCredFile,
			KeyCredentialsFile: validCredFile,
			KeySheetURL:        "https://docs.google.com/spreadsheets/d/19VVe4M-j8MGw-a3B7fcJQnx5JnHjiHf9dwChUkqQ4/edit#gid=158080911",
		},
		err: nil,
		want: Config{
			GoogleSpreadsheetID: "19VVe4M-j8MGw-a3B7fcJQnx5JnHjiHf9dwChUkqQ4",
			GoogleSheetID:       158080911,
		},
	}, {
		name: "missing required token file params",
		config: map[string]string{
			KeyCredentialsFile: validCredFile,
			KeySheetURL:        "https://docs.google.com/spreadsheets/d/19VVe4M-j8MGw-a3B7fcJQnx5JnHjiHf9dwChUkqQ4/edit#gid=158080911",
		},
		err:  fmt.Errorf(`"tokensFile" config value must be set`),
		want: Config{},
	}, {
		name: "missing required sheets url params",
		config: map[string]string{
			KeyTokensFile:      validCredFile,
			KeyCredentialsFile: validCredFile,
		},
		err:  fmt.Errorf(`"sheetsURL" config value must be set`),
		want: Config{},
	}, {
		name: "missing gid in sheets url",
		config: map[string]string{
			KeyTokensFile:      validCredFile,
			KeyCredentialsFile: validCredFile,
			KeySheetURL:        "https://docs.google.com/spreadsheets/d/19VVe4M-j8MGw-a3B7fcJQnx5JnHjiHf9dwChUkqQ4/edit",
		},
		err:  fmt.Errorf("invalid url passed, should match regex: \\/spreadsheets\\/d\\/([a-zA-Z0-9-_]+)\\/(.*)#gid=([0-9]+)"),
		want: Config{},
	}, {
		name: "invalid file",
		config: map[string]string{
			KeyTokensFile:      "invalid_file.json",
			KeyCredentialsFile: validCredFile,
			KeySheetURL:        "https://docs.google.com/spreadsheets/d/19VVe4M-j8MGw-a3B7fcJQnx5JnHjiHf9dwChUkqQ4/edit",
		},
		err:  fmt.Errorf("unable to read tokens file: open invalid_file.json: no such file or directory"),
		want: Config{},
	}, {
		name: "invalid creds",
		config: map[string]string{
			KeyTokensFile:      invalidCredFile,
			KeyCredentialsFile: invalidCredFile,
			KeySheetURL:        "https://docs.google.com/spreadsheets/d/19VVe4M-j8MGw-a3B7fcJQnx5JnHjiHf9dwChUkqQ4/edit",
		},
		err:  fmt.Errorf("unable to parse client secret file to config: oauth2/google: no credentials found"),
		want: Config{},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Parse(tt.config)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
				tt.want.OAuthConfig = cfg.OAuthConfig
				tt.want.OAuthToken = cfg.OAuthToken
				assert.Equal(t, tt.want, cfg)
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
