package main

import (
	. "./util"
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/mrjones/oauth"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
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
	consumer.Debug(false)

	accessToken := &oauth.AccessToken{Token: AccessToken, Secret: AccessTokenSecret}

	out, err := exec.Command("tail", "-1", "data.csv").Output()
	outslice := strings.Split(string(out), ",")
	lastid, _ := strconv.Atoi(strings.TrimRight(outslice[len(outslice)-1], "\n"))
	queryID := lastid - 1

	sinceTime, untilTime := ParseBetweenTime(outslice[0])

	endPointRaw := "https://api.twitter.com/1.1/search/universal.json"

	outdatas := [][]string{}

	writefile, err := os.OpenFile(os.Args[2], os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer writefile.Close()

	writer := csv.NewWriter(writefile)

	for i := 1; i <= 800; i++ {
		queryRaw := "【アイドルマスター シンデレラガールズ】 -RT"
		if queryID > 0 {
			queryRaw += " max_id:" + fmt.Sprint(queryID)
		}
		queryRaw += " since:" + sinceTime + " until:" + untilTime
		params := map[string]string{"q": queryRaw, "modules": "status", "count": "100", "result_type": "recent"}

		response, err := consumer.Get(endPointRaw, params, accessToken)
		if err != nil {
			log.Fatal(err, response)
		}

		if response.StatusCode != 200 {
			log.Fatal(response.Status)
			break
		}

		responseBody, err := ioutil.ReadAll(response.Body)

		if err != nil {
			log.Fatal(err)

		}

		var jsonPointer interface{}
		if err := json.Unmarshal(responseBody, &jsonPointer); err != nil {
			log.Fatal(err)
		}

		if retid, retTime, retdatas := ProcessData(jsonPointer); retid < 0 {
			fmt.Println("no id")
			break
		} else {
			queryID = retid - 1
			sinceTime, untilTime = ParseBetweenTime(retTime)
			outdatas = append(outdatas, retdatas...)
		}

		response.Body.Close()

		if remCount, _ := strconv.Atoi(response.Header["X-Rate-Limit-Remaining"][0]); remCount <= 0 {
			fmt.Println("no remain count")
			break
		} else {
			limittimenum, _ := strconv.ParseInt(response.Header["X-Rate-Limit-Reset"][0], 10, 64)
			fmt.Print("next reset time: ", time.Unix(limittimenum, 0))

			fmt.Println("remain: ", remCount)
		}

		if i%100 == 0 {
			if err := writer.WriteAll(outdatas); err != nil {
				log.Fatal(err)
			}
			fmt.Println("write!")
			outdatas = [][]string{}
		}
	}
	if err := writer.WriteAll(outdatas); err != nil {
		log.Fatal(err)
	}
}
