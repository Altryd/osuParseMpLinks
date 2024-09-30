package osuParseMpLinks

import (
	"encoding/json"
	"io"
	"os"
)

type SecretData struct {
	ClientId     int    `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
	Scope        string `json:"scope"`
}

func (data *SecretData) getClientId() int {
	return data.ClientId
}
func (data *SecretData) getClientSecret() string {
	return data.ClientSecret
}
func NewSecretData() SecretData {

	jsonFile, err := os.Open("osu_api_usage/secrets.json")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)
	var data *SecretData
	json.Unmarshal(byteValue, &data)

	return SecretData{
		data.getClientId(),
		data.getClientSecret(),
		data.GrantType,
		data.Scope,
	}
}
