/*
Copyright © 2022 Meroxa, Inc. & Gophers Lab Technologies Pvt. Ltd.

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

package destination

import (
	"fmt"
	"strconv"

	"github.com/conduitio/conduit-connector-google-sheets/config"
)

const (
	// KeySheetName is the name of the sheet needed to fetch data.
	KeySheetName = "sheetName"

	// KeyBufferSize is the config name for buffer size.
	KeyBufferSize = "bufferSize"

	// KeyValueInputOption is the config name for how the input data
	// should be inserted.
	KeyValueInputOption = "valueInputOption"

	// KeyMaxRetries is the config key for max retry
	KeyMaxRetries = "maxRetries"

	// defaultValueInputOption is the value ValueInputOption assumes when the config omits
	// the ValueInputOption parameter
	defaultValueInputOption = "USER_ENTERED"

	// maxBufferSize determines maximum buffer size a config can accept.
	// When config with bigger buffer size is parsed, an error is returned.
	maxBufferSize = 100

	defaultMaxRetries = "3"
)

// Config represents destination configuration with Google-Sheet configurations
type Config struct {
	config.Config
	SheetName string
	// How the data is to be interpreted by the Google sheets
	// In case of USER_ENTERED, the data is inserted similar to data insertion from browser
	// In RAW, the data is inserted without any parsing
	ValueInputOption string
	BufferSize       uint64
	MaxRetries       uint64
}

// Parse attempts to parse the configurations into a Config struct that Destination could utilize
func Parse(cfg map[string]string) (Config, error) {
	sharedConfig, err := config.Parse(cfg)
	if err != nil {
		return Config{}, fmt.Errorf("error parsing shared config, %w", err)
	}

	sheetName := cfg[KeySheetName]
	if sheetName == "" {
		return Config{}, requiredConfigErr(KeySheetName)
	}

	sheetValueOption := cfg[KeyValueInputOption]
	if sheetValueOption == "" {
		sheetValueOption = defaultValueInputOption
	}
	if sheetValueOption != "RAW" && sheetValueOption != "USER_ENTERED" {
		return Config{}, fmt.Errorf(
			"invalid value (%s) for `%s` config received, valid values: `RAW`, `USER_ENTERED`",
			KeyValueInputOption, sheetValueOption,
		)
	}

	bufferSizeString := cfg[KeyBufferSize]
	if bufferSizeString == "" {
		bufferSizeString = fmt.Sprintf("%v", maxBufferSize)
	}

	bufferSize, err := strconv.ParseUint(bufferSizeString, 10, 64)
	if err != nil || bufferSize < 1 {
		return Config{}, fmt.Errorf(
			"%q config value should be a positive integer",
			KeyBufferSize,
		)
	}

	if bufferSize > maxBufferSize {
		return Config{}, fmt.Errorf(
			"%q config value should not be bigger than %d, got %d",
			KeyBufferSize,
			maxBufferSize,
			bufferSize,
		)
	}

	retriesString := cfg[KeyMaxRetries]
	if retriesString == "" {
		retriesString = defaultMaxRetries
	}

	retries, err := strconv.ParseUint(retriesString, 10, 64)
	if err != nil {
		return Config{}, fmt.Errorf(
			"%q config value should be a positive integer",
			KeyMaxRetries,
		)
	}

	destinationConfig := Config{
		Config:           sharedConfig,
		SheetName:        sheetName,
		ValueInputOption: sheetValueOption,
		BufferSize:       bufferSize,
		MaxRetries:       retries,
	}

	return destinationConfig, nil
}

func requiredConfigErr(name string) error {
	return fmt.Errorf("%q config value must be set", name)
}
