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
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
)

const (
	KeyCredentialsFile = "google.credentialsFile"
	KeyTokensFile      = "google.tokensFile"
	KeySheetURL        = "google.sheetsURL"
)

var (
	scopes = []string{
		"https://www.googleapis.com/auth/spreadsheets.readonly",
		"https://www.googleapis.com/auth/spreadsheets",
	}
	sheetsRegexp = regexp.MustCompile(`\/spreadsheets\/d\/([a-zA-Z0-9-_]+)\/(.*)#gid=([0-9]+)`)
)

type Config struct {
	Client              *http.Client
	GoogleSpreadsheetID string
	GoogleSheetID       int64
}

func Parse(config map[string]string) (Config, error) {
	// check if configs exist
	credFile, ok := config[KeyCredentialsFile]
	if !ok || credFile == "" {
		return Config{}, requiredConfigErr(KeyCredentialsFile)
	}

	tokenFile, ok := config[KeyTokensFile]
	if !ok || tokenFile == "" {
		return Config{}, requiredConfigErr(KeyTokensFile)
	}

	sheetURL, ok := config[KeySheetURL]
	if !ok || sheetURL == "" {
		return Config{}, requiredConfigErr(KeySheetURL)
	}

	// parse files and URL

	// parse credentials.json
	credBytes, err := ioutil.ReadFile(credFile)
	if err != nil {
		return Config{}, fmt.Errorf("unable to read client secret file: %w", err)
	}

	oauthConfig, err := google.ConfigFromJSON(credBytes)
	if err != nil {
		return Config{}, fmt.Errorf("unable to parse client secret file to config: %w", err)
	}

	// parse tokens file
	var token *oauth2.Token
	tokenBytes, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		return Config{}, fmt.Errorf("unable to read tokens file: %w", err)
	}
	if err := json.Unmarshal(tokenBytes, &token); err != nil {
		return Config{}, fmt.Errorf("unable to unmarshal tokens file: %w", err)
	}
	// parse sheets url
	spreadSheetID, sheetID, err := parseSheetURL(sheetURL)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		// for some reason using cancellable context causes refresh functionality to stop working
		// using context.Background to avoid that issue
		Client:              oauthConfig.Client(context.Background(), token),
		GoogleSheetID:       sheetID,
		GoogleSpreadsheetID: spreadSheetID,
	}
	return cfg, nil
}

func requiredConfigErr(name string) error {
	return fmt.Errorf("%q config value must be set", name)
}

func parseSheetURL(url string) (string, int64, error) {
	if !sheetsRegexp.MatchString(url) {
		return "", 0, fmt.Errorf("invalid url passed, should match regex: %s", sheetsRegexp.String())
	}
	stringMatches := sheetsRegexp.FindStringSubmatch(url)
	if len(stringMatches) != 4 {
		return "", 0, fmt.Errorf("invalid url, required 4 parts, got %d", len(stringMatches))
	}
	sheetID, err := strconv.ParseInt(stringMatches[3], 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("error converting sheet id to int: %w", err)
	}
	return stringMatches[1], sheetID, nil // spreadsheetID, sheetID, error
}
