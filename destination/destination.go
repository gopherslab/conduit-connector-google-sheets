/*
Copyright © 2022 Meroxa, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package destination

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	// "github.com/conduitio/conduit-connector-google-sheets/destination/writer"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"golang.org/x/oauth2"
)

type Destination struct {
	sdk.UnimplementedDestination

	Buffer            []sdk.Record
	AckCache          []sdk.AckFunc
	Error             error
	Token             *oauth2.Token
	Client            *http.Client
	Mutex             *sync.Mutex
	DestinationConfig Config
}

func NewDestination() sdk.Destination {
	return &Destination{}
}

func (d *Destination) Configure(ctx context.Context, cfg map[string]string) error {
	fmt.Println("**********entering in Configure******************* ")

	sheetsConfig, err := Parse(cfg)
	if err != nil {
		return err
	}

	d.DestinationConfig = Config{
		Config:           sheetsConfig.Config,
		ValueInputOption: sheetsConfig.ValueInputOption,
	}

	d.Token = &oauth2.Token{
		AccessToken:  sheetsConfig.GoogleAccessToken,
		TokenType:    "Bearer",
		RefreshToken: sheetsConfig.AuthRefreshToken,
	}

	return nil
}

func (d *Destination) Open(context.Context) error {
	fmt.Println("**********entering in Open method*****************")

	var authCfg *oauth2.Config

	// initializing the buffer
	d.Buffer = make([]sdk.Record, 0, 1)
	d.AckCache = make([]sdk.AckFunc, 0, 1)

	d.Client = authCfg.Client(context.Background(), d.Token)

	return nil
}

// Data can be in following formats:
// Object {}
// Object of Array {[], []}
// Array of Array [[], []]
// Array Object [{}, {}]

func (d *Destination) WriteAsync(ctx context.Context, r sdk.Record, ack sdk.AckFunc) error {
	fmt.Println("**********entering in WriteAsyn*********************")

	if d.Error != nil {
		return d.Error
	}

	d.Mutex.Lock()
	defer d.Mutex.Unlock()

	fmt.Println("+++++++++This is the running till here+++++++++")
	fmt.Printf("\n\n+++++++++\n%#v\n+++++++++\n\n", ack)

	d.Buffer = append(d.Buffer, r)
	d.AckCache = append(d.AckCache, ack)

	fmt.Println("**********entering in WriteAsyn loop")
	if len(d.Buffer) == 1 {
		fmt.Println("===========Entered inside loop===========")
		err := d.Flush(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
func (d *Destination) Flush(ctx context.Context) error {
	bufferedRecords := d.Buffer
	d.Buffer = d.Buffer[:0]

	err := writer.Writer(bufferedRecords)
	if err != nil {
		d.Error = err
	}

	// call all the written records' ackFunctions
	for _, ack := range d.AckCache {
		err := ack(d.Error)
		if err != nil {
			return err
		}
	}
	d.AckCache = d.AckCache[:0]

	return nil
}
*/
func (d *Destination) Teardown(ctx context.Context) error {
	fmt.Println("**********entering in Teardown")

	d.Client = nil
	return nil
}
