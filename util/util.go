package util

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

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

func ProcessData(jsonPointer interface{}) (int, string, [][]string) {
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

func ParseBetweenTime(raw string) (since, until string) {
	tweetTime, _ := time.Parse("2006-01-02_15:04:05_MST", raw)
	since = tweetTime.Add(-1*time.Hour).Format("2006-01-02_15:04:05") + "_JST"
	until = tweetTime.Format("2006-01-02_15:04:05") + "_JST"

	return since, until
}
