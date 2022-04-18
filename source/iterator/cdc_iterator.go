package iterator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type CDCIterator struct {
	service       *sheets.Service
	spreadsheetId string
	iter          bool
	endPage       int64
}
type Position struct {
	Key       int64
	Timestamp time.Time
}

func NewCDCIterator(ctx context.Context, client *http.Client, spreadsheetId string, p int64) (*CDCIterator, error) {
	var err error
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	c := &CDCIterator{
		service:       srv,
		spreadsheetId: spreadsheetId,
		endPage:       p,
	}

	return c, nil
}

func (i *CDCIterator) HasNext(ctx context.Context) bool {
	sdk.Logger(ctx).Info().Msg("This is HasNext")
	sdk.Logger(ctx).Info().Msg(fmt.Sprintf("%v", i.endPage))
	sdk.Logger(ctx).Info().Msg(fmt.Sprintf("%v", (i.endPage == 0)))

	sdk.Logger(ctx).Info().Msg(fmt.Sprintf("Bool value of iter: %v", i.iter))

	// return i.endPage == 0 || i.iter
	return i.endPage > 0 || !i.iter
}

func (i *CDCIterator) Next(ctx context.Context) (sdk.Record, error) {
	sdk.Logger(ctx).Info().Msg("This is next function")

	// read object
	sheetData, err := fetchSheetData(ctx, i.service, i.spreadsheetId, i.endPage)
	if err != nil {
		return sdk.Record{}, err
	}

	if sheetData.s.Values == nil {
		sdk.Logger(ctx).Info().Msg("Data is coming nil")
		i.iter = false
		return sdk.Record{
			Metadata: map[string]string{
				"SpreadsheetId": i.spreadsheetId,
				"SheetId":       "0",
				"dimension":     sheetData.s.MajorDimension,
			},
			Position:  []byte(fmt.Sprint(sheetData.rowCount)),
			CreatedAt: time.Now(),
		}, nil
	}

	rawData, err := json.Marshal(sheetData.s.Values)
	if err != nil {
		return sdk.Record{}, fmt.Errorf("could not read the object's body: %w", err)
	}

	sdk.Logger(ctx).Info().Msg("Data from rawData: " + string(rawData))

	sdk.Logger(ctx).Info().Msg("This is positionCount: " + string([]byte(fmt.Sprintf(" %d", sheetData.rowCount))))
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
	i.endPage = sheetData.rowCount
	i.iter = true
	return output, nil
}

func (i *CDCIterator) Stop() {
	sdk.Logger(context.TODO()).Info().Msg("Hey this is inside the stop function")
	if !i.iter {
		sdk.Logger(context.TODO()).Info().Msg("Getting inside if statement of the stop function")
		return
	}
}

func fetchSheetData(ctx context.Context, srv *sheets.Service, spreadsheetId string, offset int64) (*Object, error) {
	var s sheets.DataFilter
	dataFilters := []*sheets.DataFilter{}

	s.GridRange = &sheets.GridRange{
		SheetId:       0,
		StartRowIndex: offset,
		EndRowIndex:   offset + 10,
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
			rowCount: offset + 0,
		}, nil
	}

	count := offset + int64(len(res.ValueRanges[0].ValueRange.Values))
	return &Object{
		s:        res.ValueRanges[0].ValueRange,
		rowCount: count,
	}, nil

}

type Object struct {
	s        *sheets.ValueRange
	rowCount int64
}

func ParseRecordPosition(p sdk.Position) (Position, error) {
	s := string(p)
	var err error
	if s == "" {
		return Position{
			Key:       0,
			Timestamp: time.Unix(0, 0),
		}, err
	}

	page, err := strconv.Atoi(s)
	if err != nil {
		return Position{}, fmt.Errorf("could not parse the position timestamp: %w", err)
	}

	return Position{
		Key:       int64(page),
		Timestamp: time.Unix(0, 0),
	}, err
}
