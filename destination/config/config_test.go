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
package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type destTestCase []struct {
	testCase string
	params   map[string]string
	expected Config
}

func TestParse(t *testing.T) {
	cases := destTestCase{
		{
			testCase: "Checking against default values",
			params: map[string]string{
				"SheetRange":       "",
				"ValueInputOption": "",
				"InsertDataOption": "",
			},
			expected: Config{},
		},
		{
			testCase: "Checking against if any required value is empty",
			params: map[string]string{
				"SheetRange":       "Sheet",
				"ValueInputOption": "",
				"InsertDataOption": "",
			},
			expected: Config{},
		},
		{
			testCase: "Checking against random values case",
			params: map[string]string{
				"SheetRange":       "",
				"ValueInputOption": "USER_ENTERED",
				"InsertDataOption": "",
			},
			expected: Config{},
		},
		{
			testCase: "Checking for IDEAL case - 1",
			params: map[string]string{
				"SheetRange":       "Sheet",
				"ValueInputOption": "USER_ENTERED",
				"InsertDataOption": "",
			},
			expected: Config{
				SheetRange:       "Sheet",
				ValueInputOption: "USER_ENTERED",
				InsertDataOption: "INSERT_ROW",
			},
		},
		{
			testCase: "Checking for IDEAL case - 2",
			params: map[string]string{
				"SheetRange":       "Sheet",
				"ValueInputOption": "USER_ENTERED",
			},
			expected: Config{
				SheetRange:       "Sheet",
				ValueInputOption: "USER_ENTERED",
				InsertDataOption: "INSERT_ROW",
			},
		},
		{
			testCase: "Checking for IDEAL case - 3",
			params: map[string]string{
				"SheetRange":       "Sheet",
				"ValueInputOption": "USER_ENTERED",
				"InsertDataOption": "INSERT_ROW",
			},
			expected: Config{
				SheetRange:       "Sheet",
				ValueInputOption: "USER_ENTERED",
				InsertDataOption: "INSERT_ROW",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testCase, func(t *testing.T) {
			cfg, err := Parse(tc.params)
			if assert.NotNil(t, err) {
				fmt.Println(fmt.Errorf("%w", err))
			} else {
				assert.Equal(t, tc.expected, cfg)
			}
		})
	}
}
