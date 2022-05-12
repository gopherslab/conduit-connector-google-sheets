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
package iterator

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/conduitio/conduit-connector-google-sheets/config"
	sourceConfig "github.com/conduitio/conduit-connector-google-sheets/source/config"
	"github.com/conduitio/conduit-connector-google-sheets/source/position"
	"golang.org/x/oauth2"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/stretchr/testify/assert"
	"gopkg.in/tomb.v2"
)

func TestNewSheetsIterator(t *testing.T) {
	tests := []struct {
		name    string
		config  sourceConfig.Config
		tp      position.SheetPosition
		isError bool
	}{
		{
			name: "NewSheetsIterator with RowOffset=0",
			config: sourceConfig.Config{
				Config: config.Config{
					GoogleAccessToken:   "ACCESS_TOKEN",
					GoogleSpreadsheetID: "SPREADSHEET_ID",
					TokenExpiry:         time.Time{},
					AuthRefreshToken:    "REFRESH_TOKEN",
				},
				GoogleSheetID: 0,
				PollingPeriod: time.Millisecond,
			},
			tp: position.SheetPosition{RowOffset: 0},
		}, {
			name: "NewSheetsIterator without SheetID",
			config: sourceConfig.Config{
				Config: config.Config{
					GoogleAccessToken:   "ACCESS_TOKEN",
					GoogleSpreadsheetID: "SPREADSHEET_ID",
					TokenExpiry:         time.Time{},
					AuthRefreshToken:    "REFRESH_TOKEN",
				},
				PollingPeriod: time.Millisecond,
			},
			tp: position.SheetPosition{
				RowOffset: 5,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := NewSheetsIterator(context.Background(), &http.Client{}, tt.tp, "SPREADSHEET_ID", 0, tt.config.PollingPeriod)
			if tt.isError {
				assert.NotNil(t, err)
			} else {
				assert.NotNil(t, res)
				assert.NotNil(t, res.caches)
				assert.NotNil(t, res.buffer)
				assert.NotNil(t, res.tomb)
				assert.NotNil(t, res.ticker)
			}
		})
	}
}

func TestFlush(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	tmbWithCtx, _ := tomb.WithContext(ctx)
	cdc := &SheetsIterator{
		buffer: make(chan sdk.Record, 1),
		caches: make(chan []sdk.Record, 1),
		tomb:   tmbWithCtx,
	}
	randomErr := errors.New("random error")
	cdc.tomb.Go(cdc.flush)

	in := sdk.Record{Position: []byte("some_position")}
	cdc.caches <- []sdk.Record{in}
	for {
		select {
		case <-cdc.tomb.Dying():
			assert.EqualError(t, cdc.tomb.Err(), randomErr.Error())
			cancel()
			return
		case out := <-cdc.buffer:
			assert.Equal(t, in, out)
			cdc.tomb.Kill(randomErr)
		}
	}
}

func TestGetSheetRecords(t *testing.T) {
	ctx := context.Background()
	token := &oauth2.Token{
		AccessToken:  "ACCESS_TOKEN",
		TokenType:    "Bearer",
		RefreshToken: "REFRESH_TOKEN",
		Expiry:       time.Time{},
	}

	var authCfg *oauth2.Config
	gclient := authCfg.Client(ctx, token)

	cdc := &SheetsIterator{
		sheetID:       0,
		spreadsheetID: "1ANUB)NS*765%L-NDB",
		rowOffset:     0,
		client:        gclient,
	}

	test := testHandler{}

	t.Run("", func(t *testing.T) {
		recs, _ := cdc.getSheetRecords(ctx)
		if test.isError {
			assert.Len(t, recs, 1)
			assert.Equal(t, test.expected, test.record)
		}
	})
}

type testHandler struct {
	record   []sdk.Record
	isError  bool
	expected []sdk.Record
}

func TestGetSheetRecords_RateLimit(t *testing.T) {
	// in case of nextRun being set later than now, no processing should occur
	cdc := &SheetsIterator{
		nextRun: time.Now().Add(time.Minute),
	}
	recs, err := cdc.getSheetRecords(context.Background())
	assert.Nil(t, err)
	assert.Nil(t, recs)
}

// func TestFetchRecords_429(t *testing.T) {
// 	header := http.Header{}
// 	header.Set("Retry_After", "93")
// 	th := &testHandler{
// 		t:             t,
// 		resp:          []byte(``),
// 		sheetID:       0,
// 		spreadsheetID: "",
// 	}
// 	testServer := httptest.NewServer(th)
// 	cdc := &SheetsIterator{
// 		sheetID:       0,
// 		spreadsheetID: "",
// 		rowOffset:     0,
// 		client:        &http.Client{},
// 	}
// 	ctx := context.Background()
// 	recs, err := cdc.getSheetRecords(ctx)
// 	assert.NoError(t, err)
// 	assert.Len(t, recs, 0)
// 	assert.GreaterOrEqual(t, cdc.nextRun.Unix(), time.Now().Add(90*time.Second).Unix())
// }

// type testHandler struct {
// 	t             *testing.T
// 	client        *http.Client
// 	resp          []byte
// 	sheetID       int64
// 	spreadsheetID string
// }

func TestNext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	tmbWithCtx, _ := tomb.WithContext(ctx)
	cdc := &SheetsIterator{
		buffer: make(chan sdk.Record, 1),
		caches: make(chan []sdk.Record, 1),
		tomb:   tmbWithCtx,
	}
	cdc.tomb.Go(cdc.flush)

	in := sdk.Record{Position: []byte("some_position")}
	cdc.caches <- []sdk.Record{in}
	out, err := cdc.Next(ctx)
	assert.NoError(t, err)
	assert.Equal(t, in, out)
	cancel()
	out, err = cdc.Next(ctx)
	assert.EqualError(t, err, ctx.Err().Error())
	assert.Empty(t, out)
}

func TestHasNext(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(c *SheetsIterator)
		response bool
	}{{
		name: "Has next",
		fn: func(c *SheetsIterator) {
			c.buffer <- sdk.Record{}
		},
		response: true,
	}, {
		name:     "no record in buffer",
		fn:       func(c *SheetsIterator) {},
		response: false,
	}, {
		name: "record in buffer, tomb dead",
		fn: func(c *SheetsIterator) {
			c.tomb.Kill(errors.New("random error"))
			c.buffer <- sdk.Record{}
		},
		response: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cdc := &SheetsIterator{buffer: make(chan sdk.Record, 1), tomb: &tomb.Tomb{}}
			tt.fn(cdc)
			res := cdc.HasNext(context.Background())
			assert.Equal(t, res, tt.response)
		})
	}
}

func TestStreamIterator_Stop(t *testing.T) {
	cdc := &SheetsIterator{
		tomb:   &tomb.Tomb{},
		ticker: time.NewTicker(time.Second),
	}
	cdc.Stop()
	assert.False(t, cdc.tomb.Alive())
}
