/*
Copyright © 2022 Meroxa, Inc. & Gophers Lab Technologies Pvt. Ltd.

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

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"

	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	scopes = []string{
		"https://www.googleapis.com/auth/spreadsheets.readonly",
		"https://www.googleapis.com/auth/spreadsheets",
	}
	defaultCredentialFile = "./credentials.json"
	credFile              string
	config                *oauth2.Config
	out                   string
	port                  string
	host                  string
	workingDirectory      string
	log                   *zerolog.Logger
	authCode              string
)

func init() {
	log = sdk.Logger(context.Background())

	var err error
	workingDirectory, err = os.Getwd()
	if err != nil {
		log.Fatal().Err(err).Msg("error getting working directory")
	}
	// generate an output token filename, to be used as default
	filename := fmt.Sprintf("token_%d.json", time.Now().Unix())
	flag.StringVar(&credFile, "credentials",
		path.Join(workingDirectory, defaultCredentialFile),
		"path to the credentials.json, default: "+defaultCredentialFile)
	flag.StringVar(&out, "out", path.Join(workingDirectory, filename), "file to store the generated tokens, default: ./token_<ts>.json")
	flag.StringVar(&authCode, "code", "", "generate token from auth code, if already available, and don't start the redirect server")
	flag.StringVar(&port, "port", "3000", "url port to start redirect URI listener at, default: 3000")
	flag.StringVar(&host, "host", "127.0.0.1", "url host to start redirect URI listener at, default: 127.0.0.1")

	flag.Parse()
}

func main() {
	credBytes, err := ioutil.ReadFile(defaultCredentialFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to read client secret file")
	}

	// get config from JSON
	config, err = google.ConfigFromJSON(credBytes, scopes...)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to parse client secret file to config")
	}

	if strings.TrimSpace(authCode) != "" {
		// directly generate token
		if err = processAuthCode(context.Background(), authCode); err != nil {
			log.Fatal().Err(err).Msg("error processing auth code")
		}
		log.Info().
			Str("token.json", out).
			Str("credentials.json", credFile).
			Msg("token file generated")
		return
	}

	// generate auth URL
	url := getAuthURL(config)
	log.Printf("If the browser doesn't open in a few seconds. \n"+
		"Go to the following link in your browser\n%s\n", url)

	// open a new browser with the auth URL
	if err := open(url); err != nil {
		log.Error().Err(err).Msg("error opening the URL, try opening manually")
	}

	// start a server to intercept the redirect from auth url
	http.HandleFunc("/", redirectURI)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := http.ListenAndServe(host+":"+port, nil)
		if err != nil {
			log.Error().Err(err).Msg("http listen and server stopped")
		}
	}()
	wg.Wait()
}

// Returns an url used to authenticate the user
func getAuthURL(config *oauth2.Config) string {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	authURL += "&prompt=consent"
	return authURL
}

func redirectURI(w http.ResponseWriter, r *http.Request) {
	log.Info().Str("url", r.URL.String()).Msg("redirect received")

	if err := processRedirectCallURL(r.Context(), r.URL); err != nil {
		log.Error().Err(err).Msg("error processing redirect call")
		msg := []byte("unable to process redirect call. error: " + err.Error())
		_, _ = w.Write(msg)
		return
	}

	log.Info().
		Str("token.json", out).
		Str("credentials.json", credFile).
		Msg("token file generated successfully")

	msg := []byte(`Token file generated successfully.
credentials.json file path: ` + credFile + `
token.json file path: ` + out)
	_, _ = w.Write(msg)
}

func processRedirectCallURL(ctx context.Context, url *url.URL) error {
	scope := html.UnescapeString(url.Query().Get("scope"))
	// validate scope
	scopeMap := map[string]struct{}{}
	for _, s := range strings.Split(scope, " ") {
		scopeMap[s] = struct{}{}
	}
	for _, s := range scopes {
		if _, ok := scopeMap[s]; !ok {
			log.Error().Str("scope", s).Msg("scope missing")
			return fmt.Errorf("missing scope: %s", s)
		}
	}
	return processAuthCode(ctx, url.Query().Get("code"))
}

func processAuthCode(ctx context.Context, authCode string) error {
	token, err := exchangeAuthCode(ctx, authCode)
	if err != nil {
		return err
	}
	if err := saveToken(token); err != nil {
		log.Error().Err(err).Msg("error saving token to file")
		return fmt.Errorf("unable to write token to file: %w", err)
	}
	return nil
}

func exchangeAuthCode(ctx context.Context, authCode string) (*oauth2.Token, error) {
	if strings.TrimSpace(authCode) == "" {
		log.Error().Msg("empty auth code received")
		return nil, fmt.Errorf("empty auth code received")
	}
	token, err := config.Exchange(ctx, authCode)
	if err != nil {
		log.Error().Err(err).Msg("unable to retrieve token from web")
		return nil, fmt.Errorf("unable to retrieve token from web: %w", err)
	}
	return token, nil
}

// Saves a token to a file path.
func saveToken(token *oauth2.Token) error {
	f, err := os.OpenFile(out, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer f.Close()

	if err = json.NewEncoder(f).Encode(token); err != nil {
		return fmt.Errorf("error writing token to file: %w", err)
	}
	return nil
}

// open opens the specified URL in the default browser of the user.
func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
