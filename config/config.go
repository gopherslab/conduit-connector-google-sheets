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

package config

import (
	"fmt"
	"time"
)

const (
	KeyGoogleAccessToken   = "gsheets.accessToken"
	KeyRefreshToken        = "gsheets.refreshToken"
	KeyExpiry              = "gsheets.expiry"
	KeyGoogleSpreadsheetID = "gsheets.spreadsheetId"
)

type Config struct {
	GoogleAccessToken   string
	AuthRefreshToken    string
	TokenExpiry         time.Time
	GoogleSpreadsheetID string
}

func Parse(config map[string]string) (Config, error) {
	accessToken, ok := config[KeyGoogleAccessToken]
	if !ok || accessToken == "" {
		return Config{}, requiredConfigErr(KeyGoogleAccessToken)
	}

	refreshToken, ok := config[KeyRefreshToken]
	if !ok || refreshToken == "" {
		return Config{}, requiredConfigErr(KeyRefreshToken)
	}

	spreadsheetID, ok := config[KeyGoogleSpreadsheetID]
	if !ok || spreadsheetID == "" {
		return Config{}, requiredConfigErr(KeyGoogleSpreadsheetID)
	}

	expiryString, ok := config[KeyExpiry]
	if !ok || spreadsheetID == "" {
		return Config{}, requiredConfigErr(KeyExpiry)
	}
	expiryTime, err := time.Parse(time.RFC3339, expiryString)
	if err != nil {
		return Config{}, fmt.Errorf("invalid token_expiry time passed, expected format: %s, err:%w", time.RFC3339, err)
	}

	cfg := Config{
		GoogleAccessToken:   accessToken,
		AuthRefreshToken:    refreshToken,
		GoogleSpreadsheetID: spreadsheetID,
		TokenExpiry:         expiryTime,
	}
	return cfg, nil
}

func requiredConfigErr(name string) error {
	return fmt.Errorf("%q config value must be set", name)
}
