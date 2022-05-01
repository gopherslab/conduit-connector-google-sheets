package iterator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/conduitio/conduit-connector-google-sheets/source/position"
	sdk "github.com/conduitio/conduit-connector-sdk"
	"google.golang.org/api/option"
	sheets "google.golang.org/api/sheets/v4"
)

type Object struct {
	sheetDimension string
	sheetRecords   [][]interface{}
	rowCount       int64
}

type CDCIterator struct {
	service *sheets.Service
	client  *http.Client
	// cfg          config.Config
	spreadsheetID string
	sheetID       int64
	timeInterval  time.Duration
	rp            position.SheetPosition
}

func NewCDCIterator(ctx context.Context, client *http.Client, spreadsheetId string, sheetId int64, interval time.Duration, pos position.SheetPosition) (*CDCIterator, error) {
	var err error
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	c := &CDCIterator{
		service: srv,
		client:  client,
		// cfg:           cfg,
		spreadsheetID: spreadsheetId,
		sheetID:       sheetId,
		timeInterval:  interval,
		rp:            pos,
	}
	return c, nil
}

func (i *CDCIterator) HasNext(ctx context.Context) bool {
	return time.Now().After(i.rp.NextRun)
}

func (i *CDCIterator) Next(ctx context.Context) (sdk.Record, error) {
	// read object
	sheetData, err := fetchSheetData(ctx, i.service, i.spreadsheetID, i.sheetID, i.rp.RowOffset)
	if err != nil {
		return sdk.Record{}, err
	}

	lastRowPosition := position.SheetPosition{
		RowOffset: sheetData.rowCount,
		NextRun:   time.Now(),
	}

	if len(sheetData.sheetRecords) == 0 {
		i.rp.NextRun = time.Now().Add(i.timeInterval) //i.cfg.IterationInterval)
		sdk.Logger(ctx).Info().Msg(fmt.Sprintf("The next API will hit after: %v", i.rp.NextRun))
		return sdk.Record{
			Position: lastRowPosition.RecordPosition(),
		}, sdk.ErrBackoffRetry
	}

	rawData, err := json.Marshal(sheetData.sheetRecords)
	if err != nil {
		return sdk.Record{
			Position: lastRowPosition.RecordPosition(),
		}, fmt.Errorf("could not read the object's body: %w", err)
	}

	// create the record
	output := sdk.Record{
		Metadata: map[string]string{
			"SpreadsheetId": i.spreadsheetID,              //.cfg.GoogleSpreadsheetID,
			"SheetId":       fmt.Sprintf("%d", i.sheetID), //.cfg.GoogleSheetID),
			"dimension":     sheetData.sheetDimension,
		},
		Position:  lastRowPosition.RecordPosition(),
		Payload:   sdk.RawData(rawData),
		CreatedAt: time.Now(),
	}
	i.rp.RowOffset = sheetData.rowCount
	return output, nil
}

func (i *CDCIterator) Stop() {
	// nothing to do here
}

func fetchSheetData(ctx context.Context, srv *sheets.Service, spreadsheetId string, sheetId int64, offset int64) (*Object, error) {
	var s sheets.DataFilter
	dataFilters := []*sheets.DataFilter{}
	s.GridRange = &sheets.GridRange{
		SheetId:       sheetId,
		StartRowIndex: offset,
	}
	dataFilters = append(dataFilters, &s)
	valueRenderOption := ""
	dateTimeRenderOption := "FORMATTED_STRING"
	rbt := &sheets.BatchGetValuesByDataFilterRequest{
		ValueRenderOption:    valueRenderOption,
		DataFilters:          dataFilters,
		DateTimeRenderOption: dateTimeRenderOption,
	}
	res, err := srv.Spreadsheets.Values.BatchGetByDataFilter(spreadsheetId, rbt).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	valueRange := res.ValueRanges[0].ValueRange

	if (res.HTTPStatusCode != http.StatusOK) || res == nil {
		return &Object{
			sheetDimension: valueRange.MajorDimension,
			sheetRecords:   nil,
			rowCount:       offset,
		}, nil
	}

	responseData := valueRange.Values
	for index, value := range responseData {
		if len(value) == 0 {
			responseData = responseData[:index]
			break
		}
	}
	count := offset + int64(len(responseData))
	return &Object{
		sheetDimension: valueRange.MajorDimension,
		sheetRecords:   responseData,
		rowCount:       count,
	}, nil
}
