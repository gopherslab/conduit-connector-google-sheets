package iterator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type CDCIterator struct {
	service       *sheets.Service
	spreadsheetId string
}

type Position struct {
	Key       string
	Timestamp time.Time
	// Type      Type
}

var limit, offset int64

func NewCDCIterator(ctx context.Context, client *http.Client, spreadsheetId string) (*CDCIterator, error) {
	var err error
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	c := &CDCIterator{
		service:       srv,
		spreadsheetId: spreadsheetId,
	}

	return c, nil
}

func (i *CDCIterator) HasNext(ctx context.Context) bool {
	return offset > 0 //|| lastModified == 0
}

func (i *CDCIterator) Next(ctx context.Context) (sdk.Record, error) {
	// read object
	sheetData, err := fetchSheetData(ctx, i.service, i.spreadsheetId, offset)
	if err != nil {
		return sdk.Record{}, err
	}

	rawData, err := json.Marshal(sheetData.s.Values)
	if err != nil {
		return sdk.Record{}, fmt.Errorf("could not read the object's body: %w", err)
	}

	sdk.Logger(ctx).Info().Msg("Data from rawData: " + string(rawData))

	// create the record
	output := sdk.Record{
		Metadata: map[string]string{
			"SpreadsheetId": i.spreadsheetId,
			"SheetId":       "0",
			"dimension":     sheetData.s.MajorDimension,
		},
		Position:  []byte(fmt.Sprintf("%d", sheetData.rowCount)), //ToRecordPosition
		Payload:   sdk.RawData(rawData),
		CreatedAt: time.Now(),
	}

	sdk.Logger(ctx).Info().Msg(fmt.Sprintf("Data: %s", output))
	offset = sheetData.rowCount
	return output, nil
}

func (i *CDCIterator) Stop() {
	// under development
}

func fetchSheetData(ctx context.Context, srv *sheets.Service, spreadsheetId string, offset int64) (*Object, error) {
	var s sheets.DataFilter

	dataFilters := []*sheets.DataFilter{}
	limit = offset + 10
	s.GridRange = &sheets.GridRange{
		SheetId:       0,
		StartRowIndex: offset,
		EndRowIndex:   limit,
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

	obj := &Object{
		s:        res.ValueRanges[0].ValueRange,
		rowCount: limit,
	}

	return obj, nil
}

type Object struct {
	s        *sheets.ValueRange
	rowCount int64
}
