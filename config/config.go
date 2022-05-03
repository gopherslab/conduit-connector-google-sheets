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
	ConfigKeyGoogleAccessToken = "access_token"
	ConfigKeyRefreshToken      = "refresh_token"
	// ConfigKeySheetID             = "sheet_id"
	ConfigKeyGoogleSpreadsheetID = "spreadsheet_id"
	// ConfigKeyIterationInterval   = "iteration_interval"
	// DefualtTimeInterval          = "3m"
)

type Config struct {
	GoogleAccessToken   string
	AuthRefreshToken    string
	GoogleSpreadsheetID string
	// GoogleSheetID       int64
	// IterationInterval   time.Duration
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

	spreadsheetID, ok := config[ConfigKeyGoogleSpreadsheetID]
	if !ok || spreadsheetID == "" {
		return Config{}, requiredConfigErr(ConfigKeyGoogleSpreadsheetID)
	}

	// gSheetID, ok := config[ConfigKeySheetID]
	// if !ok || gSheetID == "" {
	// 	return Config{}, requiredConfigErr(ConfigKeySheetID)
	// }

	// sheetID, err := convertToInt64(gSheetID)
	// if err != nil || sheetID < 0 {
	// 	return Config{}, fmt.Errorf("%q cannot parse sheetID from string to int64", ConfigKeySheetID)
	// }

	// Time interval being an optional value
	// interval := config[ConfigKeyIterationInterval]
	// if interval == "" {
	// 	interval = DefualtTimeInterval
	// }

	// timeInterval, err := time.ParseDuration(interval)
	// if err != nil {
	// 	return Config{}, fmt.Errorf("%q cannot parse interval to time duration", interval)
	// }

	cfg := Config{
		GoogleAccessToken:   accessToken,
		AuthRefreshToken:    refreshToken,
		GoogleSpreadsheetID: spreadsheetID,
		// GoogleSheetID:       sheetID,
		// IterationInterval:   timeInterval,
	}
	return cfg, nil
}

// func convertToInt64(value string) (int64, error) {
// 	parsed, err := strconv.ParseInt(value, 10, 64)
// 	if err != nil {
// 		return -1, fmt.Errorf("%s cannot parse value from string to int64", value)
// 	}

// 	return parsed, err
// }

func requiredConfigErr(name string) error {
	return fmt.Errorf("%q config value must be set", name)
}
