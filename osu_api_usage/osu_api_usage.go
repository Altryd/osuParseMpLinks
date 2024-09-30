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

func (client *HttpClient) UpdateToken() {
	secretData := NewSecretData()
	jsonData, err := json.Marshal(secretData)
	if err != nil {
		panic(err)
	}

	url := "https://osu.ppy.sh/oauth/token"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}
	q := req.URL.Query()
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err)
	}

	client.AccessToken = data["token_type"].(string) + " " + data["access_token"].(string)
}

func (client *HttpClient) reqUserData(usernameOrId string) (*http.Response, error) {
	url := fmt.Sprintf("https://osu.ppy.sh/api/v2/users/%s/osu", usernameOrId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	q := req.URL.Query()
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", client.AccessToken)
	req.URL.RawQuery = q.Encode()

	return client.Client.Do(req)
}

func (client *HttpClient) GetUserDataByUsernameOrId(usernameOrId string) map[string]interface{} {
	// получить данные юзера
	resp, err := client.reqUserData(usernameOrId)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.Status == "401" {
		client.UpdateToken()
		resp, err := client.reqUserData(usernameOrId)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err)
	}
	return data
}

type ParsingConfig struct {
	warmups  int
	skipLast int
	verbose  bool
	debug    bool
}

func (client *HttpClient) ParseMplink(matchArg string, parsingConfig ParsingConfig) map[string]interface{} {
	var matchUrl string
	var matchId int
	if parsingConfig.debug && len(matchArg) == 0 {
		fmt.Println("Вставьте ссылку на матч")
		reader := bufio.NewReader(os.Stdin)
		matchUrl, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
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
			panic(errors.New("invalid link: cannot find matches/"))
		}
		endOfUrl := allSubstr[0]
		splitUrl := strings.Split(endOfUrl, "/")
		if len(splitUrl) != 2 {
			panic(errors.New("invalid link: can't find match id"))
		}
		var matchIdStr string = splitUrl[1]
		var err error
		matchId, err = strconv.Atoi(matchIdStr)
		if err != nil {
			panic(errors.New("invalid link: can't convert matchIdStr to int"))
		}
		// fmt.Println(matchId)
		//matched, _ := regexp.(, matchUrl)
		//fmt.Println(matched) // false
	} else {
		var err error
		matchId, err = strconv.Atoi(matchArg)
		if err != nil {
			panic(errors.New("invalid link: can't convert matchArg to int"))
		}
	}
	// fmt.Println(matchId)
	// fmt.Println(matchUrl)
	// client.UpdateToken()
	url := fmt.Sprintf("https://osu.ppy.sh/api/v2/matches/%d", matchId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	q := req.URL.Query()
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", client.AccessToken)
	req.URL.RawQuery = q.Encode()
	resp, err := client.Client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err)
	}
	return data
}
