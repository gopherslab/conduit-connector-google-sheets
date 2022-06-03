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
package googlesheets

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/conduitio/conduit-connector-google-sheets/config"
	"github.com/conduitio/conduit-connector-google-sheets/destination"
	"github.com/conduitio/conduit-connector-google-sheets/source"
	sdk "github.com/conduitio/conduit-connector-sdk"
	"go.uber.org/goleak"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var (
	offset        int
	sheetName     string
	credFilePath  string
	tokenFilePath string
	sheetURL      string
	spreadsheetID string
	sheetID       int64
)

func TestAcceptance(t *testing.T) {
	credJSON := strings.TrimSpace(os.Getenv("CONDUIT_GOOGLE_CREDENTIAL_JSON"))
	if credJSON != "" {
		credFile, err := os.CreateTemp("", "cred*.json")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(credFilePath)
		if _, err = credFile.WriteString(credJSON); err != nil {
			t.Error("error writing cred file", err)
		}
		credFilePath = credFile.Name()
	} else {
		t.Error("credentials not set in env CONDUIT_GOOGLE_CREDENTIAL_JSON")
		t.FailNow()
	}

	tokenJSON := strings.TrimSpace(os.Getenv("CONDUIT_GOOGLE_TOKEN_JSON"))
	if tokenJSON != "" {
		tokenFile, err := os.CreateTemp("", "token*.json")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tokenFilePath)

		if _, err = tokenFile.WriteString(tokenJSON); err != nil {
			t.Error("error writing token file", err)
		}
		tokenFilePath = tokenFile.Name()
	} else {
		t.Error("token not set in env CONDUIT_GOOGLE_TOKEN_JSON")
		t.FailNow()
	}

	sheetURL = strings.TrimSpace(os.Getenv("CONDUIT_GOOGLE_SHEET_URL"))
	if sheetURL == "" {
		t.Error("sheetURL not set in env CONDUIT_GOOGLE_SHEET_URL")
		t.Skip()
	}

	sheetName = strings.TrimSpace(os.Getenv("CONDUIT_GOOGLE_SHEET_NAME"))
	if sheetName == "" {
		t.Error("sheetName not set in env CONDUIT_GOOGLE_SHEET_NAME")
		t.FailNow()
	}

	sourceConfig := map[string]string{
		"credentialsFile": credFilePath,
		"tokensFile":      tokenFilePath,
		"sheetsURL":       sheetURL,
		"pollingPeriod":   "1s", // Configurable polling period
	}

	destConfig := map[string]string{
		"credentialsFile":  credFilePath,
		"tokensFile":       tokenFilePath,
		"sheetsURL":        sheetURL,
		"sheetName":        sheetName,
		"valueInputOption": "USER_ENTERED",
		"bufferSize":       "10",
	}

	ctx := context.Background()
	conf, err := config.Parse(sourceConfig)
	if err != nil {
		t.Fatal(err)
	}

	spreadsheetID = conf.GoogleSpreadsheetID
	sheetID = conf.GoogleSheetID

	client := conf.OAuthConfig.Client(ctx, conf.OAuthToken)
	sheetService, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		t.Fatal(err)
	}

	clearSheet := func(t *testing.T) {
		_, err := sheetService.Spreadsheets.Values.Clear(conf.GoogleSpreadsheetID, "1:1000", &sheets.ClearValuesRequest{}).Do()
		if err != nil {
			t.Errorf("error cleaning the sheet: %v", err.Error())
		}
	}

	// clear sheet before starting the tests
	clearSheet(t)
	sdk.AcceptanceTest(t, AcceptanceTestDriver{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())), //nolint: gosec // only used for testing
		ConfigurableAcceptanceTestDriver: sdk.ConfigurableAcceptanceTestDriver{
			Config: sdk.ConfigurableAcceptanceTestDriverConfig{
				Connector: sdk.Connector{
					NewSpecification: Specification,
					NewSource:        source.NewSource,
					NewDestination:   destination.NewDestination,
				},
				SourceConfig:      sourceConfig,
				DestinationConfig: destConfig,
				BeforeTest: func(t *testing.T) {
				},
				GoleakOptions: []goleak.Option{goleak.IgnoreCurrent()},
				AfterTest: func(t *testing.T) {
					// clear sheet after every test to ensure clean sheet for next test
					offset = 0
					clearSheet(t)
				},
			},
		},
	})
}

type AcceptanceTestDriver struct {
	rand *rand.Rand
	sdk.ConfigurableAcceptanceTestDriver
}

// GenerateRecord overrides the pre-defined generate record function to generate the records in required google sheets compatible format
// It generates payload with 4 column row as payload
// Sample Record:
// {
//      "metadata": null,
//		"position": "{\"row_offset\":1, \"spreadsheet_id\":\"some_id\", \"sheet_id\":123}"
//      "created_at": "0001-01-01 00:00:00 +0000 UTC",
//      "key": 123,
//      "payload": "[\"a\",\"b\",\"c\",\"d\"]"
//}
func (d AcceptanceTestDriver) GenerateRecord(*testing.T) sdk.Record {
	payload := fmt.Sprintf(`["%s","%s","%s","%s"]`, d.randString(32), d.randString(32), d.randString(32), d.randString(32))
	offset++
	return sdk.Record{
		Position:  sdk.Position(fmt.Sprintf(`{"row_offset":%v, "spreadsheet_id":%v, "sheet_id":%v}`, offset, spreadsheetID, sheetID)),
		Metadata:  nil,
		CreatedAt: time.Time{},
		Key:       sdk.RawData(fmt.Sprintf("%v", offset)),
		Payload:   sdk.RawData(payload),
	}
}

// randString generates a random string of length n.
// (source: https://stackoverflow.com/a/31832326)
func (d AcceptanceTestDriver) randString(n int) string {
	const letterBytes = `0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz`
	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)
	sb := strings.Builder{}
	sb.Grow(n)
	// src.Int63() generates 63 random bits, enough for letterIdxMax characters
	for i, cache, remain := n-1, d.rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = d.rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return sb.String()
}
