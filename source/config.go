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

package source

import (
	"fmt"
	"strings"
	"time"

	"github.com/conduitio/conduit-connector-google-sheets/config"
)

const (
	// KeyPollingPeriod is the config name for the google-sheets polling period
	KeyPollingPeriod        = "pollingPeriod"
	KeyDateTimeRenderOption = "dateTimeRenderOption"
	KeyValueRenderOption    = "valueRenderOption"

	// defaultPollingPeriod is the value assumed for the pooling period when the
	// config omits the polling period parameter
	defaultPollingPeriod        = "6s"
	defaultDateTimeRenderOption = "FORMATTED_STRING"
	defaultValueRenderOption    = "FORMATTED_VALUE"
)

// Config represents source configuration with Google-Sheets configurations
type Config struct {
	config.Config
	PollingPeriod time.Duration

	// google sheets data fetch options.
	// Refer: https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/batchGet#query-parameters
	DateTimeRenderOption string // values: SERIAL_NUMBER, FORMATTED_STRING // default: SERIAL_NUMBER
	ValueRenderOption    string // values: FORMATTED_VALUE, UNFORMATTED_VALUE, FORMULA// default: FORMATTED_VALUE
}

// Parse attempts to parse the configurations into a Config struct that Source could utilize
func Parse(cfg map[string]string) (Config, error) {
	commonConfig, err := config.Parse(cfg)
	if err != nil {
		return Config{}, err
	}
	// Time interval being an optional value
	interval := strings.TrimSpace(cfg[KeyPollingPeriod])
	if interval == "" {
		interval = defaultPollingPeriod
	}

	timeInterval, err := time.ParseDuration(interval)
	if err != nil {
		return Config{}, fmt.Errorf("%q cannot parse interval to time duration", interval)
	}

	dateTimeOption := strings.TrimSpace(cfg[KeyDateTimeRenderOption])
	if dateTimeOption == "" {
		dateTimeOption = defaultDateTimeRenderOption
	}
	if dateTimeOption != "SERIAL_NUMBER" && dateTimeOption != "FORMATTED_STRING" {
		return Config{}, fmt.Errorf(
			"invalid value received for config(`%s`):`%s`, should be oneof [`SERIAL_NUMBER`, `FORMATTED_STRING`]",
			KeyDateTimeRenderOption, dateTimeOption,
		)
	}

	valueOption := strings.TrimSpace(cfg[KeyValueRenderOption])
	if valueOption == "" {
		valueOption = defaultValueRenderOption
	}
	if valueOption != "FORMATTED_VALUE" && valueOption != "UNFORMATTED_VALUE" && valueOption != "FORMULA" {
		return Config{}, fmt.Errorf(
			"invalid value received for config(`%s`):`%s`, should be oneof [`FORMATTED_VALUE`, `UNFORMATTED_VALUE`, `FORMULA`]",
			KeyValueRenderOption, valueOption,
		)
	}

	sourceConfig := Config{
		Config:               commonConfig,
		PollingPeriod:        timeInterval,
		DateTimeRenderOption: dateTimeOption,
		ValueRenderOption:    valueOption,
	}

	return sourceConfig, nil
}
