/* Copyright © 2022 Meroxa, Inc. & Gophers Lab Technologies Pvt. Ltd.

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

package sheets

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func TestWriter_NoRecord(t *testing.T) {
	ctx := context.Background()
	writer, err := NewWriter(ctx, &oauth2.Config{}, &oauth2.Token{}, "dummy_spreadsheet_id", "Sheet", "", 3)
	assert.NoError(t, err)
	err = writer.Write(ctx, nil)
	assert.NoError(t, err)
}

func TestWriter_Succeeds(t *testing.T) {
	header := http.Header{}
	th := &testHandler{
		t:          t,
		url:        &url.URL{Path: "/v4/spreadsheets/dummy/values/sheet:append", RawQuery: "alt=json&insertDataOption=INSERT_ROWS&prettyPrint=false&valueInputOption=USER_ENTERED"},
		statusCode: 200,
		resp:       []byte(`{}`),
		header:     header,
	}
	testServer := httptest.NewServer(th)
	sheetSvc, err := sheets.NewService(
		context.Background(),
		option.WithEndpoint(testServer.URL),
		option.WithHTTPClient(&http.Client{}))
	assert.NoError(t, err)
	ctx := context.Background()
	writer := &Writer{
		sheetSvc:         sheetSvc,
		sheetName:        "sheet",
		spreadsheetID:    "dummy",
		valueInputOption: "USER_ENTERED",
	}
	err = writer.Write(ctx, []sdk.Record{{Payload: sdk.RawData(`["1","2","3","4"]`)}})
	assert.NoError(t, err)
}

func TestWriter_429(t *testing.T) {
	header := http.Header{}
	th := &testHandler{
		t:          t,
		url:        &url.URL{Path: "/v4/spreadsheets/dummy/values/sheet:append", RawQuery: "alt=json&insertDataOption=INSERT_ROWS&prettyPrint=false&valueInputOption=USER_ENTERED"},
		statusCode: 429,
		resp:       []byte(`{}`),
		header:     header,
	}
	testServer := httptest.NewServer(th)
	sheetSvc, err := sheets.NewService(
		context.Background(),
		option.WithEndpoint(testServer.URL),
		option.WithHTTPClient(&http.Client{}))
	assert.NoError(t, err)
	ctx := context.Background()
	writer := &Writer{
		sheetSvc:         sheetSvc,
		sheetName:        "sheet",
		spreadsheetID:    "dummy",
		valueInputOption: "USER_ENTERED",
		maxRetries:       2,
	}
	err = writer.Write(ctx, []sdk.Record{{Payload: sdk.RawData(`["1","2","3","4"]`)}})
	assert.EqualError(t, err, "rate limit exceeded, retries: 2, error: googleapi: got HTTP response code 429 with body: {}")
}
