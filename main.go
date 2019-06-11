package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
)

const (
	twitterOAuth2TokenURL = "https://api.twitter.com/oauth2/token"
)

var (
	keyFile = flag.String("key_file", "./keys.json", "JSON containing consumer key and secret.")
)

type keys struct {
	Key    string `json:"consumer_key"`
	Secret string `json:"consumer_secret"`
}

// parseJSON reads a file containing a json formatted consumer key and secret data.
func (k *keys) parseJSON(keyFile string) error {
	jsonData, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return fmt.Errorf("parseJSON read file: %v", err)
	}
	if err := json.Unmarshal(jsonData, &k); err != nil {
		return fmt.Errorf("parseJSON could not unmarshall: %v", err)
	}
	return nil
}

func (k *keys) getB64BearerToken() string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", k.Key, k.Secret)))
}

func main() {
	flag.Parse()
	ctx := context.Background()

	var k keys
	if err := k.parseJSON(*keyFile); err != nil {
		log.Fatalf("Impossible to parse credentials from %s: %v", *keyFile, err)
	}

	bearerToken := k.getB64BearerToken()
	req, err := http.NewRequest("POST", twitterOAuth2TokenURL, strings.NewReader("grant_type=client_credentials"))
	if err != nil {
		log.Fatalf("Unable to generate http request: %v", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", bearerToken))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")

	var client http.Client
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Unable to authenticate: %v", err)
	}
	defer res.Body.Close()

	var token oauth2.Token
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&token); err != nil {
		log.Fatalf("Could not decode JSON token: %v", err)
	}

	var conf oauth2.Config
	twClient := conf.Client(ctx, &token)
	res, err = twClient.Get("https://api.twitter.com/1.1/statuses/retweets/1007246074317365248.json")
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	io.Copy(os.Stdout, res.Body)

}
