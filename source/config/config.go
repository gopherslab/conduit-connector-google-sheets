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
	"time"

	"github.com/conduitio/conduit-connector-google-sheets/config"
)

const (
	// KeyPollingPeriod is the config name for the google-sheets polling period
	KeyPollingPeriod = "pollingPeriod"

	// defaultPollingPeriod is the value assumed for the pooling period when the
	// config omits the polling period parameter
	defaultPollingPeriod = "6s"
)

// Config represents source configuration with Google-Sheets configurations
type Config struct {
	config.Config
	PollingPeriod time.Duration
}

// Parse attempts to parse the configurations into a Config struct that Source could utilize
func Parse(cfg map[string]string) (Config, error) {
	commonConfig, err := config.Parse(cfg)
	if err != nil {
		return Config{}, err
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
		PollingPeriod: timeInterval,
	}

	return sourceConfig, nil
}
