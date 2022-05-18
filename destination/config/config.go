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

package config

import (
	"fmt"
	"strconv"

	"github.com/conduitio/conduit-connector-google-sheets/config"
)

const (
	// KeySheetRange is the name of the sheet needed to fetch data.
	KeySheetName = "sheetName"

	// KeyBufferSize is the config name for buffer size.
	KeyBufferSize = "bufferSize"

	// KeyValueInputOption is the config name for how the input data
	// should be interpreted. This could be RAW or USER_ENTERED
	KeyValueInputOption = "valueInputOption"

	// KeyValueInputOption is the config name for how the input data
	// should be inserted. This could be INSERT_ROWS or OVERWRITE
	KeyInsertDataOption = "insertDataOption"

	// DefaultKeyInsertDataOption is the value InsertDataOption assumes when the config omits
	// the InsertDataOption parameter
	DefaultKeyInsertDataOption = "INSERT_ROWS"

	// maxBufferSize determines maximum buffer size a config can accept.
	// When config with bigger buffer size is parsed, an error is returned.
	maxBufferSize uint64 = 100
)

// Config represents destination configuration with Google-Sheet configurations
type Config struct {
	config.Config
	SheetName        string
	ValueInputOption string
	InsertDataOption string
	BufferSize       uint64
}

// Parse attempts to parse the configurations into a Config struct that Destination could utilize
func Parse(cfg map[string]string) (Config, error) {
	sharedConfig, err := config.Parse(cfg)
	fmt.Println(err)
	if err != nil {
		return Config{}, err
	}

	sheetName := cfg[KeySheetName]
	if sheetName == "" {
		return Config{}, requiredConfigErr(KeySheetName)
	}

	sheetValueInput := cfg[KeyValueInputOption]
	if sheetValueInput == "" {
		return Config{}, requiredConfigErr(KeyValueInputOption)
	}

	sheetDataOption := cfg[KeyInsertDataOption]
	if sheetDataOption == "" {
		sheetDataOption = DefaultKeyInsertDataOption
	}

	bufferSizeString := cfg[KeyBufferSize]
	if bufferSizeString == "" {
		bufferSizeString = fmt.Sprintf("%d", maxBufferSize)
	}

	bufferSize, err := strconv.ParseUint(bufferSizeString, 10, 64)
	if err != nil {
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

	destinationConfig := Config{
		Config:           sharedConfig,
		SheetName:        sheetName,
		ValueInputOption: sheetValueInput,
		InsertDataOption: sheetDataOption,
		BufferSize:       bufferSize,
	}

	return destinationConfig, nil
}

func requiredConfigErr(name string) error {
	return fmt.Errorf("%q config value must be set", name)
}
