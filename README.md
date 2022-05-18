# Conduit Connnector Google-Sheets

###  General
The Conduit Connector for [Google Sheets](https://github.com/gopherslab/conduit-connector-google-sheets) fetches all the records from a particular sheet.


## Google-Sheet Source

The Google-Sheet Connector connects to google sheets via google-sheets api(v4) witht the provided configuration using the `google.credentialsFile`, `google.tokensFile`, `google.sheetsURL` and along with a configurable pollingPeriod(Optional). 

The `Configure` method is called to parse the configurations. After which, the `Open` method is called to start the connection from the provided position.


### Record Fetching

Upon successful connection, an api hit will fetch all the records present in the sheet. If no more data is available, there will be a timer interval, which has a default value of 2 minutes(than can be configured in the config). On completion of this timer, the api will again hit to fetch the newly added records/rows.

If there are single/multiple empty rows in between the two records, it will fetch only the last record before the first empty row, and will hold that position until a new row/record has been added.


### Position Handling

The Google-sheet connector stores the the last row of the fetched sheet data as position. If in case, there are empty rows, Google-Sheet connector will fetch till the next row is empty and that last row will be stored as in position. 


### Configuration

The config passed to `Configure` can contain the following fields.

| name                  | description                                                                            | required  | example             |
|-----------------------|----------------------------------------------------------------------------------------|-----------|---------------------|
| `google.credentialsFile`     |  Path to credentials file which can be downloaded from Google Cloud Platform(in .json format) to authorise the user.                                                                     | yes       | "path://to/credential/file" |
| `google.tokensFile`          | Path to file in .json format which includes the `access_token`, `token_type`, `refresh_token` and `expiry`.                                                                   | yes       | "path://to/token/file"       |
| `google.sheetsURL`          | URL of the google spreadsheet(copy the entire url from the address bar).                                                                  | yes       | "https://docs.google.com/spreadsheets/d/dummy_spreadsheet_id/edit#gid=0"       |
| `pollingPeriod`       | time interval between two consecutive hits. Can be in format as s for seconds, m for minutes, h for hours (for eg: 2s; 2m; 2h)  | no        | "6s"            |


### Known Limitations

* At a time, only one `gid` inside `google.sheetsURL` can be used to fetch the records from the google-sheet.
* Any modification/update/delete made to a previous row(s) in google-sheet, after the records are fetched will not be visible in the next api hit.


## Google-Sheet Destination

The Google-Sheet Destination connector connects to the provided Google SheetID with the provided configurations, using `google.credentialsFile`, `google.tokensFile`, `google.sheetsURL` and `sheetName`.  Then will call `Configure` to parse the configurations, If parsing was not successful, then an error will occur. After that, the Open method is called to start the connection. 


### Google-Sheet Writer

The destination writer maintains a configurable buffer(default length is 10), or each time `WriteAsync` is called, a new record is added to the buffer. When the buffer is full, all the records from it will be written/appended to the last row of google-sheets and an ack function will be called for each record after being written.


### Configuration

The config passed to `Configure` can contain the following fields.


| name                  | description                                                                            | required  | example             |
|-----------------------|----------------------------------------------------------------------------------------|-----------|---------------------|
| `google.credentialsFile`     |  Path to credentials file which can be downloaded from Google Cloud Platform(in .json format) to authorise the user.                                                                     | yes       | "path://to/credential/file" |
| `google.tokensFile`          | Path to file in .json format which includes the `access_token`, `token_type`, `refresh_token` and `expiry`.                                                                   | yes       | "path://to/token/file"       |
| `google.sheetsURL`          | URL of the google spreadsheet(copy the entire url from the address bar).                                                                  | yes       | "https://docs.google.com/spreadsheets/d/dummy_spreadsheet_id/edit#gid=0"       |
| `sheetName`          | Sheet name on which the data is to be appended.                                                                  | yes       | "SHEET_NAME"       |
| `insertDataOption`       | How the data be inserted in google-sheets(i.e either `OVERWRITE` or `INSERT_ROWS`). The default value for `insertDataOption` is `INSERT_ROWS`   | no        | "insertDataOption"            |
| `bufferSize`          | Minumun number of records in buffer to hit the google-sheet api. The `buffer_size` should be less than the `maxBufferSize` whose default value is `100`, otherwise an error is thrown.                                                                 | no       | "bufferSize"            |


