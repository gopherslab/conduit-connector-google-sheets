package destinationconfig

import (
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
			if err != nil {
				assert.NotNil(t, err)
			} else {
				assert.Equal(t, tc.expected, cfg)
			}
		})
	}
}
