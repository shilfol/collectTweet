package main

import (
	"fmt"
	"github.com/mrjones/oauth"
	"io/ioutil"
	"log"
)

const (
	ConsumerKey       = "consumer key"
	ConsumerSecret    = "consumer secret"
	AccessToken       = "access token"
	AccessTokenSecret = "access token secret"
)

func main() {
	consumer := oauth.NewConsumer(ConsumerKey, ConsumerSecret, oauth.ServiceProvider{})

	// for debug
	consumer.Debug(true)

	accessToken := &oauth.AccessToken{Token: AccessToken, Secret: AccessTokenSecret}

	endPoint := "https://api.twitter.com/1.1/status/mentions_timelines.json"

	response, err := consumer.Get(endPoint, nil, accessToken)
	if err != nil {
		log.Fatal(err, response)
	}
	defer response.Body.Close()
	fmt.Println(response.StatusCode, response.Status)
	responseBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Fatal(err)

	}

	fmt.Println(string(responseBody))

}
