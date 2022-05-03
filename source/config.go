package source

import (
	"fmt"
	"strconv"
	"time"

	"github.com/conduitio/conduit-connector-google-sheets/config"
)

const (
	ConfigKeySheetID           = "sheet_id"
	ConfigKeyIterationInterval = "iteration_interval"
	DefaultTimeInterval        = "3m"
)

type Config struct {
	config.Config
	GoogleSheetID     int64
	IterationInterval time.Duration
}

func Parse(cfg map[string]string) (Config, error) {
	commonConfig, err := config.Parse(cfg)
	if err != nil {
		return Config{}, err
	}

	gSheetID, ok := cfg[ConfigKeySheetID]
	if !ok || gSheetID == "" {
		return Config{}, requiredConfigErr(ConfigKeySheetID)
	}

	sheetID, err := convertToInt64(gSheetID)
	if err != nil || sheetID < 0 {
		return Config{}, fmt.Errorf("%q cannot parse sheetID from string to int64", ConfigKeySheetID)
	}

	// Time interval being an optional value
	interval := cfg[ConfigKeyIterationInterval]
	if interval == "" {
		interval = DefaultTimeInterval
	}

	timeInterval, err := time.ParseDuration(interval)
	if err != nil {
		return Config{}, fmt.Errorf("%q cannot parse interval to time duration", interval)
	}

	sourceConfig := Config{
		Config:            commonConfig,
		GoogleSheetID:     sheetID,
		IterationInterval: timeInterval,
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
