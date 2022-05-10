package sourceconfig

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type sourceTestCase []struct {
	testCase string
	params   map[string]string
	expected Config
}

func TestParse(t *testing.T) {
	cases := sourceTestCase{
		{
			testCase: "Checking against default values",
			params: map[string]string{
				"sheet_id":           "",
				"iteration_interval": "2m",
			},
			expected: Config{},
		},
		{
			testCase: "Checking against default values",
			params: map[string]string{
				"sheet_id": "",
			},
			expected: Config{},
		},
		{
			testCase: "Checking against default values",
			params: map[string]string{
				"sheet_id":           "22",
				"iteration_interval": "",
			},
			expected: Config{
				GoogleSheetID:     22,
				IterationInterval: time.Duration(3 * time.Minute),
			},
		},
		{
			testCase: "Checking against default values",
			params: map[string]string{
				"sheet_id": "32",
			},
			expected: Config{
				GoogleSheetID:     32,
				IterationInterval: time.Duration(3 * time.Minute),
			},
		},
		{
			testCase: "Checking against default values",
			params: map[string]string{
				"sheet_id":           "",
				"iteration_interval": "",
			},
			expected: Config{},
		},
		{
			testCase: "Checking against if any required value is empty",
			params: map[string]string{
				"sheet_id":           "-1",
				"iteration_interval": "2m",
			},
			expected: Config{},
		},
		{
			testCase: "Checking against random values case",
			params: map[string]string{
				"sheet_id":           "365",
				"iteration_interval": "2s",
			},
			expected: Config{
				GoogleSheetID:     365,
				IterationInterval: time.Duration(2 * time.Second),
			},
		},
		{
			testCase: "Checking for IDEAL case - 1",
			params: map[string]string{
				"sheet_id":           "12",
				"iteration_interval": "2m",
			},
			expected: Config{
				GoogleSheetID:     12,
				IterationInterval: time.Duration(2 * time.Minute),
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