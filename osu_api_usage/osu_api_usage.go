package osuParseMpLinks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
