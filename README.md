# Conduit Connnector Google-Sheets

###  General
The Conduit Connector for [Google-Sheets](https://github.com/gopherslab/conduit-connector-google-sheets/tree/dev) fetches all the records from a particular sheet.


## Google-Sheet Source

The Google-Sheet Connector connects to google sheets via google-sheets api(v4) witht the provided configuration using the OAuth2 credentials, SpreadsheetId and SheetId along with a configurable time interval for the next consecutive run. 

The `Configure` method is called to parse the configurations. // what will happen if we provide wrong spreadsheetid and sheet id

After which, the `Open` method is called to start the connection from the provided position.


### Record Fetching

Upon successful connection, an api hit will fetch all the records present in the sheet. If no more data is available, there will be a timer interval, which has a default value of 2 minutes(than can be configured in the config). On completion of this timer, the api will again hit to fetch the newly added records/rows.

If there are single/multiple empty rows in between the two records, it will fetch only the last record before the first empty row, and will hold that position until a new row/record has been added.


#### Position Handling

The Google-sheet connector stores the the last row of the fetched sheet data as position. If in case, there are empty rows, Google-Sheet connector will fetch till the next row is empty and that last row will be stored as in position. 


### Configuration

The config passed to `Configure` can contain the following fields.

| name                  | description                                                                            | required  | example             |
|-----------------------|----------------------------------------------------------------------------------------|-----------|---------------------|
| `access_key`     |  Google Oauth2 Access Token                                                                    | yes       | "ACCESS_TOKEN" |
| `refresh_token` | Google Oauth2 Refresh Token                                                                   | yes       | "REFRESH_TOKEN" |
| `spreadsheet_id`          | Spreadsheet ID                                                                | yes       | "SPREADSHEET_ID"         |
| `sheet_id`          | Unique ID(integer value) for every sheet (i.e gid in the url)                                                                  | yes       | 0       |
| `iteration_interval`       | time interval between two consecutive hits. Can be in format as s for seconds, m for minutes, h for hours (for eg: 2s; 2m; 2h)  | no        | "2m"            |


### Known Limitations

* At a time, only one `sheet_id` can be used to fetch the records from the google-sheet.
* Any modification/update/delete made to a previous row(s) in google-sheet, after the records are fetched will not be visible in the next api hit.