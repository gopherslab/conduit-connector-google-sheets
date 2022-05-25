# Conduit Connnector Google Sheets

### General

The Conduit Connector for [Google Sheets](https://github.com/gopherslab/conduit-connector-google-sheets) provides both source as-well-as destination connector for the Google sheets

### How to build it

Run `make`.

### Testing

Run `make test` to run all the tests.

To run the integration tests under acceptance_test.go, These env variables need to be set:
`CONDUIT_GOOGLE_CREDENTIAL_JSON`: this env variable should contain the google service account's credentials JSON
`CONDUIT_GOOGLE_TOKEN_JSON`: this env variable should contain the oauth2 token JSON containing at least `refresh_token`.
`CONDUIT_GOOGLE_SHEET_URL`: the Google sheet URL, used to get the spreadsheet id and sheet id
`CONDUIT_GOOGLE_SHEET_NAME`: the name of the target sheet, this is required to be able to write to the sheet

If any of these values are not set, the integration tests will fail.

Sample execution:
```shell
CONDUIT_GOOGLE_CREDENTIAL_JSON=$(cat testdata/dummy_cred.json) \
CONDUIT_GOOGLE_TOKEN_JSON=$(cat testdata/dummy_token.json) \
CONDUIT_GOOGLE_SHEET_URL=https://docs.google.com/spreadsheets/d/1SJ5x-gyC0iWMjq6o9WMz21kBk5P8m4GtP0loaRcpNJU/edit#gid=0 \
CONDUIT_GOOGLE_SHEET_NAME=Sheet1 \
make test
```

## Generating credentials

In order to generate credentials.json, please follow [this link](https://developers.google.com/workspace/guides/create-credentials). After the credentials.json is generated, download the json file and place it inside your root project. To generate token file(i.e token_UnixTimeStamp.json), run `./google-token-gen` from the root project. A browser window will open, to verify the gmail account followed by the conset page. 

Once successful, you will get the following message: 
```
Token file generated successfully.
credentials.json file path: file/path/to/credentials.json
token.json file path: file/path/to/token_1653466634.json

```

Copy both the .json file paths and provide them in `google.credentialsFile`, `google.tokensFile`. 


## Google Sheet Source

The Google Sheet Connector connects to google sheets via Google Sheets API(v4) with the provided configuration using the `google.credentialsFile`, `google.tokensFile`, `google.sheetsURL` and along with a configurable pollingPeriod(Optional). 

The `Configure` method is called to parse the configurations. After which, the `Open` method is called to start the connection from the provided position.


### Record Fetching

Upon successful connection, an api hit will fetch all the records present in the sheet. If no more data is available, the iterator will programatically pause till the configurable pollingPeriod duration(which has a default value of "6s") and on completion of this duration, the api is hit again to fetch the newly added records/rows.

If there are single/multiple empty rows in between the two records, it will fetch only the last record before the first empty row, and will hold that position until a new row/record has been added.


### Position Handling

The Google Sheets connector stores the last row of the fetched sheet data as position. If in case, there are empty row(s), the Sheets connector will fetch till the last non-empty row and that last row will be stored as in position. 


### Configuration

The config passed to `Configure` in source connector can contain the following fields.

| name                                    | description                                                                                                                    | required | example                                                                  |
|-----------------------------------------|--------------------------------------------------------------------------------------------------------------------------------|----------|--------------------------------------------------------------------------|
| `google.credentialsFile`                | Path to credentials file which can be downloaded from Google Cloud Platform(in .json format) to authorise the user.            | yes      | "path://to/credential/file"                                              |
| `google.tokensFile`                     | Path to file in .json format which includes the `access_token`, `token_type`, `refresh_token` and `expiry`.                    | yes      | "path://to/token/file"                                                   |
| `google.sheetsURL`                      | URL of the google spreadsheet(copy the entire url from the address bar).                                                       | yes      | "https://docs.google.com/spreadsheets/d/dummy_spreadsheet_id/edit#gid=0" |
| `google.DateTimeRenderOption`           | Format of the Date/time related values. Valid values: SERIAL_NUMBER, FORMATTED_STRING                                          | no       | "FORMATTED_STRING"                                                       |
| `google.ValueRenderOption`              | Format of the dynamic/reference data. Valid values: FORMATTED_VALUE, UNFORMATTED_VALUE, FORMULA                                | no       | "FORMATTED_VALUE"                                                        |
| `pollingPeriod`                         | time interval between two consecutive hits. Can be in format as s for seconds, m for minutes, h for hours (for eg: 2s; 2m; 2h) | no       | "6s"                                                                     |

### Known Limitations

* At a time, only one `gid` as s subset string in `google.sheetsURL` can be used to fetch the records from the google sheets.
* Empty Rows will be skipped while fetching.
* Any modification/update/delete made to a previous row(s) in google sheets, after the records are fetched will not be visible in the next api hit.



## Google Sheet Destination

The Google Sheet Destination connector connects to the provided Google SheetID with the provided configurations, using `google.credentialsFile`, `google.tokensFile`, `google.sheetsURL` and `sheetName`.  Then will call `Configure` to parse the configurations. If parsing was not successful, then an error will occur. After that, the `Open` method is called to start the connection. 


### Google Sheet Writer

The destination writer maintains a configurable buffer(default length is 100), or each time `WriteAsync` is called, a new record is added to the buffer. The `bufferSize` is configurable and the max value, the buffer can be is 100, minumum it could be 1. Once the buffer is full(as per the configured value), all the records from it will be written/appended to the last row of google sheets and an ack function will be called for each record after being written.


### Configuration

The config passed to `Configure` in destination connector can contain the following fields.


| name                      | description                                                                                                                                                                            | required  | example                                                                  |
|---------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------|--------------------------------------------------------------------------|
| `google.credentialsFile`  | Path to credentials file which can be downloaded from Google Cloud Platform(in .json format) to authorise the user.                                                                    | yes       | "path://to/credential/file"                                              |
| `google.tokensFile`       | Path to file in .json format which includes the `access_token`, `token_type`, `refresh_token` and `expiry`.                                                                            | yes       | "path://to/token/file"                                                   |
| `google.sheetsURL`        | URL of the google spreadsheet(copy the entire url from the address bar).                                                                                                               | yes       | "https://docs.google.com/spreadsheets/d/dummy_spreadsheet_id/edit#gid=0" |
| `google.sheetName`        | Sheet name on which the data is to be appended.                                                                                                                                        | yes       | "sheetName"                                                              |
| `google.valueInputOption` | Whether the data should be parsed, similar to adding data from browser, or as a raw string. Values: "RAW", "USER_ENTERED"(default)                                                     | no        | "USER_ENTERED"                                                           |
| `maxRetries`              | Number of API retries to be made, in case of rate-limit error, before returning an error. Default: 3                                                                                   | no       | "3"                                                                      |
| `bufferSize`              | Minumun number of records in buffer to hit the google sheet api. The `buffer_size` should be less than the `maxBufferSize` whose default value is `100`, otherwise an error is thrown. | no       | "bufferSize"                                                             |



### Known Limitations

* At current, while appending data to google sheets, we are only supporting ROWS parameter.
* The `insertDataOption` field value is kept to `INSERT_ROWS`, as `OVERWRITE` does not provide the expected action. For more on `insertDataOption`, kindly refer to [this](https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/append#InsertDataOption).



## Note of caution

As the Google Sheets API is a shared service, quotas and limitations are applied to make sure it's used fairly by all users. If a quota is exceeded, you'll generally receive a 429: Too many requests HTTP status code response. Though the case is handled by implementing our own retry strategy which is to multiplying the counter value to the configurable polling period which plays a significant role in avoiding such HTTP statuses.

Ref: https://developers.google.com/sheets/api/limits



## References 

1. Google Credentials creation
    https://developers.google.com/workspace/guides/create-credentials

2. Google Sheets API(v4)
    https://developers.google.com/sheets/api/reference/rest

3. To fetch all data of a Google Sheets
    https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/batchGetByDataFilter

4. Add/Append data to Google Sheets
    https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/append