package main

import (
	"fmt"
	osuParseMpLinks "osu_api_usage"
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
	parsConf.Debug = true
	mplinkData, userData, err := client.ParseMplink("", parsConf)
	if err != nil {
		fmt.Println("ERROR:", err)
	} else {
		fmt.Println("data from ParseMpLink")
		fmt.Println(mplinkData)
		fmt.Println(userData)
	}
	mplinkData, userData, err = client.ParseScrim("https://osu.ppy.sh/community/matches/111534249", parsConf)
	if err != nil {
		fmt.Println("ERROR:", err)
	} else {
		fmt.Println("\ndata from ParseScrim")
		fmt.Println(mplinkData)
		fmt.Println(userData)
	}

}
