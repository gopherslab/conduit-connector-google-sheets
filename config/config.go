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
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	// KeyCredentialsFile is the config name for Google access key
	KeyCredentialsFile = "google.credentialsFile"

	// KeyTokensFile is the config name for google generated token file
	KeyTokensFile = "google.tokensFile"

	// KeySheetURL is the config name for google-sheets url
	KeySheetURL = "google.sheetsURL"
)

var (
	// scopes for spreadsheets are required in order to access the google SheetAPI.
	scopes = []string{
		"https://www.googleapis.com/auth/spreadsheets.readonly",
		"https://www.googleapis.com/auth/spreadsheets",
	}
	sheetsRegexp = regexp.MustCompile(`\/spreadsheets\/d\/([a-zA-Z0-9-_]+)\/(.*)#gid=([0-9]+)`)
)

// Config represent configuration needed for google-sheets
type Config struct {
	// haris: at least to me, it feels like the client should be part of the iterator
	// and should be "assembled" there, not in the config. Config is a passive entity,
	// it only contains the config params and functions to manipulate them.
	Client              *http.Client
	GoogleSpreadsheetID string
	GoogleSheetID       int64
}

// Parse attempts to parse plugins.Config into a Config struct
func Parse(config map[string]string) (Config, error) {
	// check if configs exist
	credFile := config[KeyCredentialsFile]
	if credFile == "" {
		return Config{}, requiredConfigErr(KeyCredentialsFile)
	}

	tokenFile := config[KeyTokensFile]
	if tokenFile == "" {
		return Config{}, requiredConfigErr(KeyTokensFile)
	}

	sheetURL := config[KeySheetURL]
	if sheetURL == "" {
		return Config{}, requiredConfigErr(KeySheetURL)
	}

	// parse files and URL
	authClient, err := NewOauthClient(credFile, tokenFile)
	if err != nil {
		return Config{}, err
	}

	// parse sheets url
	spreadSheetID, sheetID, err := parseSheetURL(sheetURL)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		// for some reason using cancellable context causes refresh functionality to stop working
		// using context.Background to avoid that issue
		// haris: what's the above comment about? Maybe outdated, since I don't see any context usage here.
		Client:              authClient,
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

func NewOauthClient(credFile, tokenFile string) (*http.Client, error) {
	// parse credentials.json
	credBytes, err := ioutil.ReadFile(credFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %w", err)
	}

	oauthConfig, err := google.ConfigFromJSON(credBytes, scopes...)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %w", err)
	}

	// parse tokens file
	var token *oauth2.Token
	tokenBytes, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read tokens file: %w", err)
	}
	if err := json.Unmarshal(tokenBytes, &token); err != nil {
		return nil, fmt.Errorf("unable to unmarshal tokens file: %w", err)
	}
	return oauthConfig.Client(context.Background(), token), nil
}
