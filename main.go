package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

func retweeters(client *http.Client, tweetID string) ([]string, error) {
	type user struct {
		ScreenName string `json:"screen_name"`
	}
	// retweet contains a slice of retweets for a given tweet ID.
	type retweet struct {
		User user `json:"user"`
	}

	url := fmt.Sprintf("https://api.twitter.com/1.1/statuses/retweets/%s.json", tweetID)
	res, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("(retweeters) get \"%s\" failed: %v", url, err)
	}

	var retweets []retweet
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&retweets); err != nil {
		return nil, fmt.Errorf("(retweeters) decode JSON failed: %v", err)
	}

	// This creates the slice and allocate the memory in one go. It is
	// faster than growing the slice.
	usernames := make([]string, 0, len(retweets))
	for _, u := range retweets {
		usernames = append(usernames, u.User.ScreenName)
	}
	return usernames, nil
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
	usernames, err := retweeters(twClient, "1007246074317365248")
	if err != nil {
		log.Fatalf("Could not get retweets: %v", err)
	}
	fmt.Println(usernames)

}
