package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
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

func main() {
	var k keys
	if err := k.parseJSON(*keyFile); err != nil {
		log.Fatalf("Impossible to parse credentials from %s: %v", *keyFile, err)
	}
	fmt.Println(k)

}
