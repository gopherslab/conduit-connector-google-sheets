package position

import (
	"encoding/json"
	"fmt"

	sdk "github.com/conduitio/conduit-connector-sdk"
)

type SheetPosition struct {
	RowOffset int64  `json:"row_offset"`
	NextRun   int64  `json:"next_run"`
}

func ParseRecordPosition(p sdk.Position) (SheetPosition, error) {
	var (
		err            error
		recordPosition SheetPosition
	)

	if p == nil {
		return SheetPosition{}, err
	}

	err = json.Unmarshal(p, &recordPosition)
	if err != nil {
		return SheetPosition{}, fmt.Errorf("could not parse the position timestamp: %w", err)
	}

	return recordPosition, err
}

func (s SheetPosition) RecordPosition() sdk.Position {
	pos, err := json.Marshal(s)
	if err != nil {
		return sdk.Position{}
	}
	return pos
}
