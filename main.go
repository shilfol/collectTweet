package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/mrjones/oauth"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
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

func textRegExp(rawstr string) (name, plat string) {
	regPlat := regexp.MustCompile(`[\s\S]*【(アイドルマスター シンデレラガールズ([\s　]*スターライトステージ)*)】[\s\S]+`)
	regName := regexp.MustCompile(`[\s\S]+[\r\n|\n|\r\s](\S+)に投票したよ!![\s\S]+`)

	platSlice := regPlat.FindStringSubmatch(rawstr)
	nameSlice := regName.FindStringSubmatch(rawstr)

	if len(platSlice) <= 0 || len(nameSlice) <= 0 {
		return "", ""
	}

	return nameSlice[1], platSlice[1]
}

func processData(jsonPointer interface{}) (int, string, [][]string) {
	outdatas := [][]string{}
	datas := jsonPointer.(map[string]interface{})["modules"].([]interface{})

	lastid := -1
	lastTime := ""
	for _, v := range datas {
		out := v.(map[string]interface{})["status"]
		data := out.(map[string]interface{})["data"]
		text := data.(map[string]interface{})
		user := text["user"].(map[string]interface{})

		createTimeRaw, _ := time.Parse("Mon Jan 2 15:04:05 -0700 2006", text["created_at"].(string))
		createTime := createTimeRaw.Add(9*time.Hour).Format("2006-01-02_15:04:05") + "_JST"
		plat, name := textRegExp(text["text"].(string))
		lastid, _ = strconv.Atoi(text["id_str"].(string))
		lastTime = createTime
		if plat != "" && name != "" {
			fmt.Println(createTime, plat, name, user["screen_name"], text["id_str"])
			outdata := []string{createTime, plat, name, user["screen_name"].(string), text["id_str"].(string)}
			outdatas = append(outdatas, outdata)
		}
	}
	return lastid, lastTime, outdatas
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
	tweetTime, _ := time.Parse("2006-01-02_15:04:05_MST", outslice[0])
	sinceTime := tweetTime.Add(-1*time.Hour).Format("2006-01-02_15:04:05") + "_JST"
	untilTime := tweetTime.Format("2006-01-02_15:04:05") + "_JST"

	endPointRaw := "https://api.twitter.com/1.1/search/universal.json"

	outdatas := [][]string{}

	writefile, err := os.OpenFile(os.Args[2], os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer writefile.Close()

	writer := csv.NewWriter(writefile)

	for i := 0; i < 800; i++ {
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

		if retid, retTime, retdatas := processData(jsonPointer); retid < 0 {
			fmt.Println("no id")
			break
		} else {
			queryID = retid - 1
			tweetTime, _ = time.Parse("2006-01-02_15:04:05_MST", retTime)
			sinceTime = tweetTime.Add(-1*time.Hour).Format("2006-01-02_15:04:05") + "_JST"
			untilTime = tweetTime.Format("2006-01-02_15:04:05") + "_JST"
			outdatas = append(outdatas, retdatas...)
		}

		response.Body.Close()

		if remCount, _ := strconv.Atoi(response.Header["X-Rate-Limit-Remaining"][0]); remCount <= 0 {
			fmt.Println("no remain count")
			limittimenum, _ := strconv.ParseInt(response.Header["X-Rate-Limit-Reset"][0], 10, 64)
			fmt.Println("next reset time: ", time.Unix(limittimenum, 0))
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
