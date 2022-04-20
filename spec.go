// Copyright © 2022 Meroxa, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package googlesheets

import (
	"github.com/conduitio/conduit-connector-google-sheets/config"

	sdk "github.com/conduitio/conduit-connector-sdk"
)

// Specification returns the Plugin's Specification.
func Specification() sdk.Specification {
	return sdk.Specification{
		Name:              "google-sheets",
		Summary:           "Google Sheets plugin",
		Description:       "A plugin capable of fetching records (in JSON format) from google spreadsheet.",
		Version:           "v0.1.0",
		Author:            "Meroxa, Inc.",
		DestinationParams: map[string]sdk.Parameter{},
		SourceParams: map[string]sdk.Parameter{
			config.ConfigKeyGoogleAccessToken: {
				Default:     "",
				Required:    true,
				Description: "Google sign-in access token",
			},
			config.ConfigKeyGoogleSpreadsheetId: {
				Default:     "",
				Required:    true,
				Description: "Google sheet id to fetch the records",
			},
			config.ConfigKeySpreadsheetName: {
				Default:     "",
				Required:    true,
				Description: "Google Sheet Name to fetch the records from a single SpreadsheetId",
			},
		},
	}
}
