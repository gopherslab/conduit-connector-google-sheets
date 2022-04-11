// Copyright © 2022 Meroxa, Inc.
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
)

const (
	ConfigKeyGoogleAccessToken      = "access_token"
	ConfigKeyRefreshToken           = "refresh_token"
	ConfigKeyGoogleSpreadsheetId    = "spreadsheet_id"
	ConfigKeyGoogleSpreadsheetRange = "range"
)

// var knownFieldTypes = []string{"int", "string", "time", "bool"}

type Config struct {
	GoogleAccessToken      string
	RefreshToken           string
	GoogleSpreadsheetId    string
	GoogleSpreadsheetRange string
}

func Parse(config map[string]string) (Config, error) {
	accessToken, ok := config[ConfigKeyGoogleAccessToken]
	if !ok || accessToken == "" {
		return Config{}, requiredConfigErr(ConfigKeyGoogleAccessToken)
	}

	refreshToken, ok := config[ConfigKeyRefreshToken]
	if !ok || refreshToken == "" {
		return Config{}, requiredConfigErr(ConfigKeyRefreshToken)
	}

	spreadsheetId, ok := config[ConfigKeyGoogleSpreadsheetId]
	if !ok || spreadsheetId == "" {
		return Config{}, requiredConfigErr(ConfigKeyGoogleSpreadsheetId)
	}

	sheetRange, ok := config[ConfigKeyGoogleSpreadsheetRange]
	if !ok || sheetRange == "" {
		return Config{}, requiredConfigErr(ConfigKeyGoogleSpreadsheetRange)
	}

	cfg := Config{
		GoogleAccessToken:      accessToken,
		RefreshToken:           refreshToken,
		GoogleSpreadsheetId:    spreadsheetId,
		GoogleSpreadsheetRange: sheetRange,
	}

	return cfg, nil
}

func requiredConfigErr(name string) error {
	return fmt.Errorf("%q config value must be set", name)
}
