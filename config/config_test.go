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
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase []struct {
	testCase string
	params   map[string]string
	expected Config
}

func TestParse(t *testing.T) {
	cases := testCase{
		{
			testCase: "Checking against if any/all required value are empty",
			params: map[string]string{
				"access_token":       "",
				"refresh_token":      "",
				"spreadsheet_id":     "",
			},
			expected: Config{},
		},
		{
			testCase: "Checking against random values case",
			params: map[string]string{
				"access_token":       "asdfghjkl",
				"refresh_token":      "qweafdfv",
				"spreadsheet_id":     "",
			},
			expected: Config{},
		},
		{
			testCase: "Checking for IDEAL case",
			params: map[string]string{
				"access_token":       "access-token here",
				"refresh_token":      "refresh-token here",
				"spreadsheet_id":     "123abcd",
			},
			expected: Config{
				GoogleAccessToken:   "access-token here",
				AuthRefreshToken:    "refresh-token here",
				GoogleSpreadsheetID: "123abcd",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testCase, func(t *testing.T) {
			cfg, err := Parse(tc.params)
			if err != nil {
				assert.NotNil(t, err)
			} else {
				assert.Equal(t, tc.expected, cfg)
			}
		})
	}
}
