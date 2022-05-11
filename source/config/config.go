/*Copyright © 2022 Meroxa, Inc.

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
	"strconv"
	"time"

	"github.com/conduitio/conduit-connector-google-sheets/config"
)

const (
	KeySheetID           = "sheet_id"
	KeyPollingPeriod     = "polling_period"
	defaultPollingPeriod = "6s"
)

type Config struct {
	config.Config
	GoogleSheetID int64
	PollingPeriod time.Duration
}

func Parse(cfg map[string]string) (Config, error) {
	commonConfig, err := config.Parse(cfg)
	if err != nil {
		return Config{}, err
	}

	gSheetID, ok := cfg[KeySheetID]
	if !ok || gSheetID == "" {
		return Config{}, requiredConfigErr(KeySheetID)
	}

	sheetID, err := convertToInt64(gSheetID)
	if err != nil || sheetID < 0 {
		return Config{}, fmt.Errorf("%q cannot parse sheetID from string to int64", KeySheetID)
	}

	// Time interval being an optional value
	interval := cfg[KeyPollingPeriod]
	if interval == "" {
		interval = defaultPollingPeriod
	}

	timeInterval, err := time.ParseDuration(interval)
	if err != nil {
		return Config{}, fmt.Errorf("%q cannot parse interval to time duration", interval)
	}

	sourceConfig := Config{
		Config:        commonConfig,
		GoogleSheetID: sheetID,
		PollingPeriod: timeInterval,
	}

	return sourceConfig, nil
}

func convertToInt64(value string) (int64, error) {
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return -1, fmt.Errorf("%s cannot parse value from string to int64", value)
	}

	return parsed, err
}

func requiredConfigErr(name string) error {
	return fmt.Errorf("%q config value must be set", name)
}
