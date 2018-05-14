package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/mrjones/oauth"
	"io/ioutil"
	"log"
	"os"
	//"net/url"
)

var (
	ConsumerKey       = "Consumer Key"
	ConsumerSecret    = "Consumer Secret"
	AccessToken       = "Access Token"
	AccessTokenSecret = "Access Token Secret"
)

func readKeys() {

	if len(os.Args) < 2 {
		// error
		return
	}

	fp, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	sc := bufio.NewScanner(fp)

	sc.Scan()
	ConsumerKey = sc.Text()
	sc.Scan()
	ConsumerSecret = sc.Text()
	sc.Scan()
	AccessToken = sc.Text()
	sc.Scan()
	AccessTokenSecret = sc.Text()
}

func main() {
	readKeys()

	consumer := oauth.NewConsumer(ConsumerKey, ConsumerSecret, oauth.ServiceProvider{})

	// for debug
	consumer.Debug(true)

	accessToken := &oauth.AccessToken{Token: AccessToken, Secret: AccessTokenSecret}

	/*
		client, err := consumer.MakeHttpClient(accessToken)
		if err != nil {
			log.Fatal(err, client)
		}
	*/

	endPointRaw := "https://api.twitter.com/1.1/search/universal.json"
	//endPoint := (&url.URL{Path: endPointRaw}).String()
	queryRaw := "【アイドルマスター シンデレラガールズ】 -RT max_id:990305126484033536"
	params := map[string]string{"q": queryRaw, "modules": "status", "count": "100"}

	response, err := consumer.Get(endPointRaw, params, accessToken)
	if err != nil {
		log.Fatal(err, response)
	}
	defer response.Body.Close()
	fmt.Println(response.StatusCode, response.Status)

	responseBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Fatal(err)

	}

	var testjson interface{}
	if err := json.Unmarshal(responseBody, &testjson); err != nil {
		log.Fatal(err)
	}

	datas := testjson.(map[string]interface{})["modules"].([]interface{})

	count := 0
	for _, v := range datas {
		out := v.(map[string]interface{})["status"]
		data := out.(map[string]interface{})["data"]
		text := data.(map[string]interface{})
		user := text["user"].(map[string]interface{})
		fmt.Println(text["text"], text["id_str"], user["screen_name"])
		fmt.Println()
		count++
	}

	fmt.Println(count)

}
