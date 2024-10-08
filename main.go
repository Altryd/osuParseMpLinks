package main

import (
	"fmt"
	osuParseMpLinks "osuParseMpLinks/osu_api_usage"
)

func main() {

	print("Hellow\n")
	client := osuParseMpLinks.NewHttpClient()
	data, err := client.GetUserDataByUsernameOrId("9109550")
	if err != nil {
		print(err)
	} else {
		fmt.Println("data from GetUserDataByUsernameOrId", data)
	}
	var parsConf osuParseMpLinks.ParsingConfig
	parsConf.Verbose = true
	mplinkData, userData, err := client.ParseMplink("https://osu.ppy.sh/community/matches/111534249/", parsConf)
	if err != nil {
		fmt.Println("ERROR:", err)
	} else {
		fmt.Println("data from ParseMpLink")
		fmt.Println(mplinkData)
		fmt.Println(userData)
	}
}
