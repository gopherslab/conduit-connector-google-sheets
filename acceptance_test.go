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
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/conduitio/conduit-connector-google-sheets/destination"
	"github.com/conduitio/conduit-connector-google-sheets/source"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"go.uber.org/goleak"
)

var (
	records   []sdk.Record
	pos       sdk.Position
	rowOffset int64
	ctx       context.Context
)

func init() {
	var inputRecords []sdk.Record
	inputBytes := []byte(`["Name","EmployeID","Salary","Age","State","Position"]`)
	inputRecord := sdk.Record{
		Payload:  sdk.RawData(inputBytes),
		Key:      sdk.RawData(`A1`),
		Position: []byte(`{"row_offset":1}`),
	}
	inputRecords = append(inputRecords, inputRecord)

	records = inputRecords
	pos = []byte(`{"row_offset":1}`)
	rowOffset = 1
	ctx = context.Background()
}

func TestAcceptance(t *testing.T) {
	filePath := getFilePath("conduit-connector-google-sheets")
	validCredFile := fmt.Sprintf("%s/testdata/dummy_cred.json", filePath)

	sourceConfig := map[string]string{
		"google.credentialsFile": validCredFile,
		"google.tokensFile":      validCredFile,
		"google.sheetsURL":       "https://docs.google.com/spreadsheets/d/1gQjm4hnSdrMFyPjhlwSGLBbj0ACOxFQJpVST1LmW6Hg/edit#gid=0",
		"pollingPeriod":          "6s", // Configurable polling period
	}

	destConfig := map[string]string{
		"google.credentialsFile": validCredFile,
		"google.tokensFile":      validCredFile,
		"google.sheetsURL":       "https://docs.google.com/spreadsheets/d/1gQjm4hnSdrMFyPjhlwSGLBbj0ACOxFQJpVST1LmW6Hg/edit#gid=0",
		"sheetName":              "Sheet2",
		"insertDataOption":       "INSERT_ROWS",
		"bufferSize":             "10",
	}

	sdk.AcceptanceTest(t, sdk.ConfigurableAcceptanceTestDriver{
		Config: sdk.ConfigurableAcceptanceTestDriverConfig{
			Connector: sdk.Connector{ // Note that this variable should rather be created globally in `connector.go`
				NewSpecification: Specification,
				NewSource:        source.NewSource,
				NewDestination:   destination.NewDestination,
			},
			SourceConfig:      sourceConfig,
			DestinationConfig: destConfig,
			GoleakOptions:     []goleak.Option{goleak.IgnoreCurrent()},
			Skip: []string{
				// the following tests are skipped, because they need a valid credential file and token file
				// required for oauth2 authorisation in order to create a google-sheet client to work properly.
				"TestSource_Open_ResumeAtPosition",
				"TestDestination_WriteAsync_Success",
				"TestDestination_WriteOrWriteAsync",
				"TestDestination_Write_Success",
				"TestSource_Read_Success",
				"TestSource_Read_Timeout",
			},
		},
	})
}

func getFilePath(path string) string {
	wd, _ := os.Getwd()
	for !strings.HasSuffix(wd, path) {
		wd = filepath.Dir(wd)
	}
	return wd
}
