package main

import (
	gs "github.com/conduitio/conduit-connector-google-sheets"
	"github.com/conduitio/conduit-connector-google-sheets/destination"
	"github.com/conduitio/conduit-connector-google-sheets/source"
	sdk "github.com/conduitio/conduit-connector-sdk"
)

func main() {
	// sdk.Serve(gs.Specification, source.NewSource, nil)
	sdk.Serve(gs.Specification, source.NewSource, destination.NewDestination)
}
