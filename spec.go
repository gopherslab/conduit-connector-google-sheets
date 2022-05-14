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
	dconfig "github.com/conduitio/conduit-connector-google-sheets/destination/config"
	sconfig "github.com/conduitio/conduit-connector-google-sheets/source/config"

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
				Description: "path to token.json file containing a json with atleast refresh_token.",
			},
			config.KeySheetURL: {
				Default:     "",
				Required:    true,
				Description: "Google sheet url to fetch the records from",
			},
			dconfig.KeySheetRange: {
				Default:     "",
				Required:    true,
				Description: "Google sheet id to fetch the records",
			},
			dconfig.KeyValueInputOption: {
				Default:     "",
				Required:    true,
				Description: "Google sheet id to fetch the records",
			},
			dconfig.KeyInsertDataOption: {
				Default:     dconfig.DefaultKeyInsertDataOption,
				Required:    false,
				Description: "Google sheet id to fetch the records",
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
			sconfig.KeyPollingPeriod: {
				Default:     "6s",
				Required:    false,
				Description: "Time interval for consecutive fetching data.",
			},
		},
	}
}
