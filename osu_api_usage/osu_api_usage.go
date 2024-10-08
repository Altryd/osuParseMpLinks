package osuParseMpLinks

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type HttpClient struct {
	AccessToken string
	Client      *http.Client
}

func NewHttpClient() HttpClient {
	client := HttpClient{
		"DefaultToken",
		&http.Client{},
	}
	client.UpdateToken()

	return client
}

func (client *HttpClient) UpdateToken() error {
	secretData := NewSecretData()
	jsonData, err := json.Marshal(secretData)
	if err != nil {
		return err
	}

	url := "https://osu.ppy.sh/oauth/token"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	q := req.URL.Query()
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return err
	}

	client.AccessToken = data["token_type"].(string) + " " + data["access_token"].(string)
	return nil
}

func (client *HttpClient) reqUserData(usernameOrId string) (*http.Response, error) {
	url := fmt.Sprintf("https://osu.ppy.sh/api/v2/users/%s/osu", usernameOrId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", client.AccessToken)
	req.URL.RawQuery = q.Encode()

	return client.Client.Do(req)
}

func (client *HttpClient) GetUserDataByUsernameOrId(usernameOrId string) (map[string]interface{}, error) {
	// получить данные юзера
	resp, err := client.reqUserData(usernameOrId)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.Status == "401" {
		client.UpdateToken()
		resp, err := client.reqUserData(usernameOrId)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type ParsingConfig struct {
	warmups  int
	skipLast int
	verbose  bool
	debug    bool
	matchcostStandard int
}

func (client *HttpClient) reqMatchData(method string, url string, body io.Reader) (map[string]interface{}, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", client.AccessToken)
	req.URL.RawQuery = q.Encode()
	resp, err := client.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return map[string]interface{}{}, err
	}
	var data map[string]interface{}
	err = json.Unmarshal(respBody, &data)
	return data, err
}

func (client *HttpClient) ParseMplink(matchArg string, parsingConfig ParsingConfig) ([]map[string]interface{}, map[int]interface{}, error) {
	var matchUrl string
	var matchId int
	if parsingConfig.debug && len(matchArg) == 0 {
		fmt.Println("Вставьте ссылку на матч")
		reader := bufio.NewReader(os.Stdin)
		matchUrl, err := reader.ReadString('\n')
		if err != nil {
			return nil, nil, err
		}
		fmt.Println(matchUrl)
	} else {
		if len(matchArg) == 0 {
			matchUrl = "111555364"
		} else {
			matchUrl = matchArg
		}
	}
	if strings.Contains(matchUrl, "/") || strings.Contains(matchUrl, "\\") {
		matchesRegexp := regexp.MustCompile(`matches/\d+`)
		allSubstr := matchesRegexp.FindAllString(matchArg, -1)
		if allSubstr == nil {
			 return nil, nil, errors.New("invalid link: cannot find matches/")
		}
		endOfUrl := allSubstr[0]
		splitUrl := strings.Split(endOfUrl, "/")
		if len(splitUrl) != 2 {
			 return nil, nil, errors.New("invalid link: can't find match id")
		}
		var matchIdStr = splitUrl[1]
		var err error
		matchId, err = strconv.Atoi(matchIdStr)
		if err != nil {
			 return nil, nil, errors.New("invalid link: can't convert matchIdStr to int")
		}
	} else {
		var err error
		matchId, err = strconv.Atoi(matchArg)
		if err != nil {
			return nil, nil, errors.New("invalid link: can't convert matchArg to int")
		}
	}
	url := fmt.Sprintf("https://osu.ppy.sh/api/v2/matches/%d", matchId)
	data, err := client.reqMatchData(http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("can't request match data: %s", err))
	}
	// START WORKING WITH DATA
	var userDict map[int]interface{} // for user data
	userDict = make(map[int]interface{})
	var allScoresList []map[string]interface{}  // for scores in multiplayer

	firstEventId, ok := data["first_event_id"].(float64)
	if ok != true {
		return nil, nil, errors.New("can't convert first_event_id to float64")
	}
	lastEventId, ok := data["latest_event_id"].(float64)
	if ok != true {
		return nil, nil, errors.New("can't convert latest_event_id to float64")
	}
	eventId := firstEventId
	for eventId < lastEventId {
		jsonDataForQueryStr := map[string]any{
			"after": eventId,
		}
		jsonDataForQuery, _ := json.Marshal(jsonDataForQueryStr)
		url := fmt.Sprintf("https://osu.ppy.sh/api/v2/matches/%d", matchId)
		matchData, err := client.reqMatchData(http.MethodGet, url, bytes.NewBuffer(jsonDataForQuery))
		if err != nil {
			return nil, nil, errors.New(fmt.Sprintf("can't request match data by url: %s; error: %s", url, err))
		}
		dataUsers, ok := matchData["users"].([]interface{})
		if ok != true {
			return nil, nil, errors.New("can't convert data['users'] to list of maps")
		}
		for _, user := range dataUsers {
			dataUserDict, ok := user.(map[string]interface{})
			if ok != true {
				 return nil, nil, errors.New("can't convert userDict to map")
			}

			userIdInterface := dataUserDict["id"]
			userIdFloat, ok := userIdInterface.(float64)
			if ok != true {
				 return nil, nil, errors.New("can't convert userId to float")
			}
			userId := int(userIdFloat)
			_, ok = userDict[userId]
			// If the key exists
			if !ok { // the key is not in dict => user has not been added to the dict
				userDict[userId] = map[string]interface{}{"username": dataUserDict["username"], "score_sum": 0, "played_maps": map[int]interface{}{}}

			}

		}
		eventsData := matchData["events"].([]interface{})
		for _, event := range eventsData {
			dataEventDict, ok := event.(map[string]interface{})
			if ok != true {
				return nil, nil, errors.New("can't convert event to map")
			}
			dataEventDetailDict, ok := dataEventDict["detail"].(map[string]interface{})
			if ok != true {
				continue
			}
			dataEventGameDict, ok := dataEventDict["game"].(map[string]interface{})
			if ok != true {
				continue
			}
			dataEventGameScoresList, ok := dataEventGameDict["scores"].([]interface{})
			if ok != true {
				continue
			}
			if dataEventDetailDict["type"] == "other" && len(dataEventGameScoresList) > 0 {
				var scoresMap map[string]interface{} // for user data
				scoresMap = make(map[string]interface{})
				scoresMap["scores"] = dataEventGameScoresList
				scoresMap["beatmap_id"] = dataEventGameDict["beatmap_id"]
				allScoresList = append(allScoresList, scoresMap)
			}
		}
		lastEventInBatchDict, ok := eventsData[len(eventsData)-1].(map[string]interface{})
		if ok != true {
			return nil, nil, errors.New("can't convert lastEventInBatch to map[string]interface{}")
		}
		eventId, ok = lastEventInBatchDict["id"].(float64)
		if ok != true {
			return nil, nil, errors.New("can't convert lastEventInBatch['id'] to float64")
		}
	}

	if parsingConfig.warmups > 0 {
		allScoresList = allScoresList[parsingConfig.warmups:]
	}
	if parsingConfig.skipLast > 0 {
		allScoresList = allScoresList[0:parsingConfig.skipLast]
	}
	for _, scoreStruct := range allScoresList {
		scoreStructDict := scoreStruct
		if ok != true {
			return nil, nil, errors.New("can't convert scoreStruct to map")
		}
		beatmapIdFloat, ok := scoreStructDict["beatmap_id"].(float64)
		if ok != true {
			return nil, nil, errors.New("can't convert beatmapId to float64")
		}
		beatmapId := int(beatmapIdFloat)
		scores, ok := scoreStructDict["scores"].([]interface{})
		if ok != true {
			return nil, nil, errors.New("can't convert scores to list of maps")
		}
		for _, score := range scores {
			scoreDict, ok := score.(map[string]interface{})
			if ok != true {
				return nil, nil, errors.New("can't convert score to map")
			}
			userIdFloat := scoreDict["user_id"].(float64)
			userId := int(userIdFloat)
			userInfo := userDict[userId].(map[string]interface{})
			playedMaps := userInfo["played_maps"].(map[int]interface{})
			_, hasPlayedTwice := playedMaps[beatmapId]
			if hasPlayedTwice {
				oldBeatmapScoreDict, ok := playedMaps[beatmapId].(map[string]interface{})
				if ok != true {
					return nil, nil, errors.New("can't convert playedMap[beatmapId] to map")
				}
				oldScore := oldBeatmapScoreDict["score"].(float64)
				if oldScore < scoreDict["score"].(float64) {
					playedMaps[beatmapId] = scoreDict["score"].(float64)
				}
			} else {
				playedMaps[beatmapId] = scoreDict["score"].(float64)
			}
		}
	}
	if parsingConfig.matchcostStandard == 0 {
		parsingConfig.matchcostStandard = 500000
	}
	for _, userDetails := range userDict {
		mapsPlayed := 0
		userDetailsDict, ok := userDetails.(map[string]interface{})
		if ok != true {
			return nil, nil, errors.New("can't convert userDict to map")
		}
		userPlayedMaps, ok := userDetailsDict["played_maps"].(map[int]interface{})
		if ok != true {
			return nil, nil, errors.New("can't convert userDict to map")
		}
		userDetailsDict["score_sum"] = 0
		userDictScoreSum := float64(userDetailsDict["score_sum"].(int))
		for _, scoreStruct := range userPlayedMaps {
			mapsPlayed += 1
			score, ok := scoreStruct.(float64)
			if ok != true {
				return nil, nil, errors.New("can't convert scoreStruct to float64")
			}
			userDictScoreSum += score
		}
		userDetailsDict["score_sum"] = userDictScoreSum
		userDetailsDict["average_score"] = 0
		if len(userPlayedMaps) > 0 {
			userDetailsDict["average_score"] = userDictScoreSum / float64(mapsPlayed)
		}
	}
	if parsingConfig.verbose {
		fmt.Println("All scores count:", len(allScoresList))
		var averageScoresList [][]float64
		for userId, userDetails := range userDict {
			userDetailsDict := userDetails.(map[string]interface{})
			userDetailsAvgScore := userDetailsDict["average_score"].(float64)
			newEntry := []float64{float64(userId), userDetailsAvgScore}
			averageScoresList= append(averageScoresList, newEntry)
		}
		sort.Slice(averageScoresList, func(i, j int) bool {
			return averageScoresList[i][1] > averageScoresList[j][1]  // DESC
		})
		for _, sortedEntry := range averageScoresList {
			userId := int(sortedEntry[0])
			userDetailsDict := userDict[userId].(map[string]interface{})
			avgScore := userDetailsDict["average_score"].(float64)
			username := userDetailsDict["username"].(string)
			playedMaps := userDetailsDict["played_maps"].(map[int]interface{})
			scoreSum := int(userDetailsDict["score_sum"].(float64))
			fmt.Printf("username: %s (id: %d); avg. score: %f ; match cost: %f ; played maps: %d ; score sum: %d \n",
				username, userId, avgScore, avgScore / float64(parsingConfig.matchcostStandard), len(playedMaps), scoreSum)
		}
	}
	// TODO: refactor
	return allScoresList, userDict, nil
}
