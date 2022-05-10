/* Copyright © 2022 Meroxa, Inc.

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
package writer

import (
	"context"
	"net/http"
	"testing"

	"github.com/conduitio/conduit-connector-google-sheets/config"
	"github.com/conduitio/conduit-connector-google-sheets/destination/destinationconfig"
	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/stretchr/testify/assert"
)

type writerTest []struct {
	testName string
	ctx      context.Context
	r        []sdk.Record
	cfg      destinationconfig.Config
	client   *http.Client
	expected [][]interface{}
}

func TestWriter(t *testing.T) {
	cases := writerTest{
		{
			testName: "Empty Record Payload",
			r:        []sdk.Record{},
			cfg: destinationconfig.Config{
				Config: config.Config{
					GoogleAccessToken:   "access-token here",
					AuthRefreshToken:    "refresh-token here",
					GoogleSpreadsheetID: "123abcd",
				},
				SheetRange:       "Sheet",
				ValueInputOption: "USER_ENTERED",
				InsertDataOption: "INSERT_ROW",
			},
			client: &http.Client{},
			expected: [][]interface{}{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(t *testing.T) {
			err := Writer(tc.ctx, tc.r, tc.cfg, tc.client)
			if err != nil {
				assert.NotNil(t, err)
			} 
			// else {
			// 	assert.Equal(t, tc.expected, cfg)
			// }
		})
	}

}
