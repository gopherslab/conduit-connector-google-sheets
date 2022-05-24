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

package googlesheets

import (
	"github.com/conduitio/conduit-connector-google-sheets/config"
	"github.com/conduitio/conduit-connector-google-sheets/destination"
	"github.com/conduitio/conduit-connector-google-sheets/source"
	sdk "github.com/conduitio/conduit-connector-sdk"
)

// Specification returns the Plugin's Specification.
func Specification() sdk.Specification {
	return sdk.Specification{
		Name:        "google-sheets",
		Summary:     "Google Sheets plugin",
		Description: "A plugin capable of fetching records (in JSON format) from google spreadsheet.",
		Version:     "v0.1.0",
		Author:      "Gophers Lab Technologies Pvt Ltd",
		DestinationParams: map[string]sdk.Parameter{
			config.KeyCredentialsFile: {
				Default:     "",
				Required:    true,
				Description: "path to credentials.json file used",
			},
			config.KeyTokensFile: {
				Default:     "",
				Required:    true,
				Description: "path to token.json file containing a json with at least refresh_token.",
			},
			config.KeySheetURL: {
				Default:     "",
				Required:    true,
				Description: "Google sheet url to fetch the records from",
			},
			destination.KeySheetName: {
				Default:     "",
				Required:    true,
				Description: "Google sheet name to fetch the records",
			},
			destination.KeyValueInputOption: {
				Default:     "USER_ENTERED",
				Required:    false,
				Description: "Whether the data be inserted in USER_ENTERED mode or RAW mode",
			},
			destination.KeyMaxRetries: {
				Default:     "3",
				Required:    false,
				Description: "Max API retries to be attempted, in case of 429 error, before returning error",
			},
		},
		SourceParams: map[string]sdk.Parameter{
			config.KeyCredentialsFile: {
				Default:     "",
				Required:    true,
				Description: "path to credentials.json file used",
			},
			config.KeyTokensFile: {
				Default:     "",
				Required:    true,
				Description: "path to token.json file containing a json with atleast refresh_token.",
			},
			config.KeySheetURL: {
				Default:     "",
				Required:    true,
				Description: "Google sheet url to fetch the records from",
			},
			source.KeyPollingPeriod: {
				Default:     "6s",
				Required:    false,
				Description: "Time interval for consecutive fetching data.",
			},
			source.KeyDateTimeRenderOption: {
				Default:     "FORMATTED_STRING",
				Required:    false,
				Description: "Format of the Date/time related values. Valid values: SERIAL_NUMBER, FORMATTED_STRING",
			},
			source.KeyValueRenderOption: {
				Default:     "FORMATTED_VALUE",
				Required:    false,
				Description: "Format of the dynamic/reference data. Valid values: FORMATTED_VALUE, UNFORMATTED_VALUE, FORMULA",
			},
		},
	}
}
