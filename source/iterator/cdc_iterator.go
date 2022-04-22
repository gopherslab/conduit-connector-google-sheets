package iterator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/conduitio/conduit-connector-google-sheets/config"
	"github.com/conduitio/conduit-connector-google-sheets/source/position"
	sdk "github.com/conduitio/conduit-connector-sdk"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Object struct {
	sheetData *sheets.ValueRange
	rowCount  int64
}

type CDCIterator struct {
	service *sheets.Service
	client  *http.Client
	cfg     config.Config
	rp      position.SheetPosition
}

func NewCDCIterator(ctx context.Context, client *http.Client, cfg config.Config, pos position.SheetPosition) (*CDCIterator, error) {
	var err error
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	c := &CDCIterator{
		service: srv,
		client:  client,
		cfg:     cfg,
		rp:      pos,
	}

	return c, nil
}

func (i *CDCIterator) HasNext(ctx context.Context) bool {
	return time.Now().After(i.rp.NextRun)
	// return i.rp.RowOffset == 0 || fmt.Sprintf("%d", time.Now().Unix()) > fmt.Sprintf("%d", i.rp.NextRun)
}

func (i *CDCIterator) Next(ctx context.Context) (sdk.Record, error) {
	// read object
	sheetData, err := fetchSheetData(ctx, i.service, i.cfg, i.rp.RowOffset)
	if err != nil {
		return sdk.Record{}, err
	}

	lastRowPosition := position.SheetPosition{
		RowOffset: sheetData.rowCount,
		NextRun:   time.Now(),
	}

	if sheetData.sheetData.Values == nil {
		i.rp.NextRun = time.Now().Add(i.cfg.IterationInterval)
		return sdk.Record{
			Position: lastRowPosition.RecordPosition(),
		}, sdk.ErrBackoffRetry
	}

	rawData, err := json.Marshal(sheetData.sheetData.Values)
	if err != nil {
		return sdk.Record{
			Position: lastRowPosition.RecordPosition(),
		}, fmt.Errorf("could not read the object's body: %w", err)
	}

	// create the record
	output := sdk.Record{
		Metadata: map[string]string{
			"SpreadsheetId": i.cfg.GoogleSpreadsheetId,
			"SheetId":       fmt.Sprintf("%d", i.cfg.GoogleSheetID),
			"dimension":     sheetData.sheetData.MajorDimension,
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

func fetchSheetData(ctx context.Context, srv *sheets.Service, gsheet config.Config, offset int64) (*Object, error) {
	var s sheets.DataFilter
	dataFilters := []*sheets.DataFilter{}
	s.GridRange = &sheets.GridRange{
		SheetId:       gsheet.GoogleSheetID,
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

	res, err := srv.Spreadsheets.Values.BatchGetByDataFilter(gsheet.GoogleSpreadsheetId, rbt).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	if (res.HTTPStatusCode != http.StatusOK) || res == nil {
		return &Object{
			sheetData: nil,
			rowCount:  offset,
		}, nil
	}

	count := offset + int64(len(res.ValueRanges[0].ValueRange.Values))
	return &Object{
		sheetData: res.ValueRanges[0].ValueRange,
		rowCount:  count,
	}, nil

}
