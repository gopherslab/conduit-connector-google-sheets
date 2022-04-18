package position

import (
	"fmt"
	"strconv"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
)

type Position struct {
	Key       int64
	Timestamp time.Time
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
