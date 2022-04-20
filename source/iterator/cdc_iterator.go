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
	s        *sheets.ValueRange
	rowCount int64
}

type CDCIterator struct {
	service         *sheets.Service
	client          *http.Client
	spreadsheetId   string
	spreadsheetName string
	rp              position.SheetPosition
}

func NewCDCIterator(ctx context.Context, client *http.Client, cfg config.Config, pos position.SheetPosition) (*CDCIterator, error) {
	var err error
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	c := &CDCIterator{
		service:         srv,
		client:          client,
		spreadsheetId:   cfg.GoogleSpreadsheetId,
		spreadsheetName: cfg.GoogleSpreadsheetName,
		rp:              pos,
	}

	return c, nil
}

func (i *CDCIterator) HasNext(ctx context.Context) bool {
	return i.rp.RowOffset == 0 || fmt.Sprintf("%d", time.Now().Unix()) > fmt.Sprintf("%d", i.rp.NextRun) //i.rp.NextRun > time.Time{} //!i.iter
}

func (i *CDCIterator) Next(ctx context.Context) (sdk.Record, error) {
	// read object
	sheetData, err := fetchSheetData(ctx, i.service, i.spreadsheetId, i.rp.RowOffset)
	if err != nil {
		return sdk.Record{}, err
	}

	lastRow := position.SheetPosition{
		SheetName: i.spreadsheetName,
		RowOffset: sheetData.rowCount,
		NextRun:   time.Now().Unix(),
	}

	if sheetData.s.Values == nil {
		i.rp.NextRun = time.Now().Add(3 * time.Minute).Unix()
		// i.iter = false
		return sdk.Record{}, sdk.ErrBackoffRetry
	}

	rawData, err := json.Marshal(sheetData.s.Values)
	if err != nil {
		return sdk.Record{}, fmt.Errorf("could not read the object's body: %w", err)
	}

	// create the record
	output := sdk.Record{
		Metadata: map[string]string{
			"SpreadsheetId": i.spreadsheetId,
			"SheetId":       "0",
			"dimension":     sheetData.s.MajorDimension,
		},
		Position:  lastRow.RecordPosition(),
		Payload:   sdk.RawData(rawData),
		CreatedAt: time.Now(),
	}

	sdk.Logger(ctx).Info().Msg(fmt.Sprintf("Data: %s", output))
	i.rp.RowOffset = sheetData.rowCount
	// i.iter = true
	return output, nil
}

func (i *CDCIterator) Stop() {
	// if !i.iter {
	// 	return
	// }
}

func fetchSheetData(ctx context.Context, srv *sheets.Service, spreadsheetId string, offset int64) (*Object, error) {
	var s sheets.DataFilter
	dataFilters := []*sheets.DataFilter{}

	s.GridRange = &sheets.GridRange{
		// SheetId:       0,
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

	if res == nil {
		return &Object{
			s:        nil,
			rowCount: offset,
		}, nil
	}

	count := offset + int64(len(res.ValueRanges[0].ValueRange.Values))
	return &Object{
		s:        res.ValueRanges[0].ValueRange,
		rowCount: count,
	}, nil

}
