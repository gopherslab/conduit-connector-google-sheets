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

package sheets

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func TestNewBatchReader(t *testing.T) {
	got, err := NewBatchReader(context.Background(), BatchReaderArgs{
		OAuthToken:           &oauth2.Token{},
		OAuthConfig:          &oauth2.Config{},
		SpreadsheetID:        "dummy_spreadsheet",
		SheetID:              1234,
		DateTimeRenderOption: "SOME_VALUE",
		ValueRenderOption:    "SOME_OTHER_VALUE",
		PollingPeriod:        3 * time.Second,
	})
	assert.NoError(t, err)
	want := &BatchReader{
		spreadsheetID:        "dummy_spreadsheet",
		sheetID:              1234,
		dateTimeRenderOption: "SOME_VALUE",
		valueRenderOption:    "SOME_OTHER_VALUE",
		pollingPeriod:        3 * time.Second,
	}
	want.sheetSvc = got.sheetSvc
	assert.Equal(t, want, got)
}

func TestBatchReader_getDataFilter(t *testing.T) {
	br := &BatchReader{
		sheetID:              1234,
		dateTimeRenderOption: "DATE_TIME_OPTION",
		valueRenderOption:    "VALUE_OPTION",
	}
	want := &sheets.BatchGetValuesByDataFilterRequest{
		DataFilters: []*sheets.DataFilter{
			{GridRange: &sheets.GridRange{
				SheetId: 1234, StartRowIndex: 10},
			},
		},
		DateTimeRenderOption: "DATE_TIME_OPTION",
		MajorDimension:       "ROWS",
		ValueRenderOption:    "VALUE_OPTION",
	}

	assert.Equal(t, want, br.getDataFilter(10))
}

func TestBatchReader_valueRangesToRecords(t *testing.T) {
	in := []*sheets.MatchedValueRange{{ValueRange: &sheets.ValueRange{
		MajorDimension: "ROWS",
		Range:          "Sheet1!A11:Z2870",
		Values: [][]interface{}{
			{"iqmQgVHVFVpPvpDE0byR5p1T5PUp2cI1", "UB3Io7g5OotmBHfcm77CHeGQ5PoZeYp1", "mZTcV547WwIwHROkNAT9x8yEGiV4ne8z", "FB6VpQxEUwEm8mGYePvJhnO8gtbVEmsC"},
			{"bE7DlmbAEvHpxSmKJrVNL56lH2RkD6Cj", "TpDm60cyptSfI2vRX1NgoHFxxAKBjFRB", "B3iRkGlbFCu2A8Hy3d0Ln6TqU0HO8rTT", "lTDJbi7FZPu9OrpFsz14X6msCdONz9a2"},
		}}}}

	br := &BatchReader{
		sheetID:              1234,
		spreadsheetID:        "dummy_spreadsheet",
		dateTimeRenderOption: "DATE_TIME_OPTION",
		valueRenderOption:    "VALUE_OPTION",
	}
	out, err := br.valueRangesToRecords(in, 10)
	assert.NoError(t, err)
	want := []sdk.Record{
		{
			Position: sdk.Position(`{"row_offset":11,"spreadsheet_id":"dummy_spreadsheet","sheet_id":1234}`),
			Key:      sdk.RawData(`11`),
			Payload:  sdk.RawData(`["iqmQgVHVFVpPvpDE0byR5p1T5PUp2cI1","UB3Io7g5OotmBHfcm77CHeGQ5PoZeYp1","mZTcV547WwIwHROkNAT9x8yEGiV4ne8z","FB6VpQxEUwEm8mGYePvJhnO8gtbVEmsC"]`),
		}, {
			Position: sdk.Position(`{"row_offset":12,"spreadsheet_id":"dummy_spreadsheet","sheet_id":1234}`),
			Key:      sdk.RawData(`12`),
			Payload:  sdk.RawData(`["bE7DlmbAEvHpxSmKJrVNL56lH2RkD6Cj","TpDm60cyptSfI2vRX1NgoHFxxAKBjFRB","B3iRkGlbFCu2A8Hy3d0Ln6TqU0HO8rTT","lTDJbi7FZPu9OrpFsz14X6msCdONz9a2"]`),
		},
	}
	for i := range out {
		out[i].CreatedAt = time.Time{}
		out[i].Metadata = nil
	}
	assert.Equal(t, want, out)
}

type testHandler struct {
	t          *testing.T
	url        *url.URL
	statusCode int
	header     http.Header
	resp       []byte
}

func (t *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	assert.Equal(t.t, t.url.Path, r.URL.Path)
	assert.Equal(t.t, t.url.RawQuery, r.URL.RawQuery)

	for key, val := range t.header {
		w.Header().Set(key, val[0])
	}
	w.WriteHeader(t.statusCode)
	_, _ = w.Write(t.resp)
}

func TestBatchReader_GetSheetRecords_429(t *testing.T) {
	header := http.Header{}
	header.Set("Retry-After", "93")
	th := &testHandler{
		t:          t,
		url:        &url.URL{Path: "/v4/spreadsheets/dummy_spreadsheet/values:batchGetByDataFilter", RawQuery: "alt=json&prettyPrint=false"},
		statusCode: 429,
		resp:       []byte(``),
		header:     header,
	}
	testServer := httptest.NewServer(th)
	sheetSvc, err := sheets.NewService(
		context.Background(),
		option.WithEndpoint(testServer.URL),
		option.WithHTTPClient(&http.Client{}))
	assert.NoError(t, err)
	cursor := &BatchReader{
		nextRun:       time.Time{},
		spreadsheetID: "dummy_spreadsheet",
		sheetID:       1234,
		sheetSvc:      sheetSvc,
		pollingPeriod: 10 * time.Second,
	}
	ctx := context.Background()
	recs, err := cursor.GetSheetRecords(ctx, 10)
	assert.NoError(t, err)
	assert.Len(t, recs, 0)
	assert.GreaterOrEqual(t, cursor.nextRun.Unix(), time.Now().Add(9*time.Second).Unix())
}

func TestBatchReader_GetSheetRecords_500(t *testing.T) {
	header := http.Header{}
	header.Set("Retry-After", "93")
	th := &testHandler{
		t:          t,
		url:        &url.URL{Path: "/v4/spreadsheets/dummy_spreadsheet/values:batchGetByDataFilter", RawQuery: "alt=json&prettyPrint=false"},
		statusCode: 500,
		resp:       []byte(``),
		header:     header,
	}
	testServer := httptest.NewServer(th)
	sheetSvc, err := sheets.NewService(
		context.Background(),
		option.WithEndpoint(testServer.URL),
		option.WithHTTPClient(&http.Client{}))
	assert.NoError(t, err)
	cursor := &BatchReader{
		nextRun:       time.Time{},
		spreadsheetID: "dummy_spreadsheet",
		sheetID:       1234,
		sheetSvc:      sheetSvc,
		pollingPeriod: 10 * time.Second,
	}
	ctx := context.Background()
	_, err = cursor.GetSheetRecords(ctx, 10)
	assert.EqualError(t, err, "error getting sheet(gid:1234) values, googleapi: got HTTP response code 500 with body: ")
}

func TestBatchReader_GetSheetRecords_304(t *testing.T) {
	header := http.Header{}
	header.Set("Retry-After", "93")
	th := &testHandler{
		t:          t,
		url:        &url.URL{Path: "/v4/spreadsheets/dummy_spreadsheet/values:batchGetByDataFilter", RawQuery: "alt=json&prettyPrint=false"},
		statusCode: 304,
		resp:       []byte(``),
		header:     header,
	}
	testServer := httptest.NewServer(th)
	sheetSvc, err := sheets.NewService(
		context.Background(),
		option.WithEndpoint(testServer.URL),
		option.WithHTTPClient(&http.Client{}))
	assert.NoError(t, err)
	cursor := &BatchReader{
		nextRun:       time.Time{},
		spreadsheetID: "dummy_spreadsheet",
		sheetID:       1234,
		sheetSvc:      sheetSvc,
		pollingPeriod: 10 * time.Second,
	}
	ctx := context.Background()
	recs, err := cursor.GetSheetRecords(ctx, 10)
	assert.NoError(t, err)
	assert.Nil(t, recs)
}
