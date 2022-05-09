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
package destinationconfig

import (
	"fmt"

	"github.com/conduitio/conduit-connector-google-sheets/config"
)

const (
	ConfigKeySheetRange = "sheet_range"

	// This could be RAW or USER_ENTERED
	ConfigKeyValueInputOption = "value_input_option"

	// Optional
	ConfigKeyInsertDataOption = "insert_data_option"

	// This could be INSERT_ROWS or OVERWRITE
	DefaultKeyInsertDataOption = "INSERT_ROWS"
)

type Config struct {
	config.Config
	SheetRange       string
	ValueInputOption string
	InsertDataOption string
}

func Parse(cfg map[string]string) (Config, error) {
	sharedConfig, err := config.Parse(cfg)
	if err != nil {
		return Config{}, err
	}

	sheetRange, exists := cfg[ConfigKeySheetRange]
	if !exists || sheetRange == "" {
		return Config{}, requiredConfigErr(ConfigKeySheetRange)
	}

	sheetValueInput, exists := cfg[ConfigKeyValueInputOption]
	if !exists || sheetValueInput == "" {
		return Config{}, requiredConfigErr(ConfigKeyValueInputOption)
	}

	sheetDataOption, exists := cfg[ConfigKeyInsertDataOption]
	if !exists || sheetDataOption == "" {
		sheetDataOption = DefaultKeyInsertDataOption
	}

	destinationConfig := Config{
		Config:           sharedConfig,
		SheetRange:       sheetRange,
		ValueInputOption: sheetValueInput,
		InsertDataOption: sheetDataOption,
	}

	return destinationConfig, nil
}

func requiredConfigErr(name string) error {
	return fmt.Errorf("%q config value must be set", name)
}
