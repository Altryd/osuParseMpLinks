package main

import (
	"fmt"
	osuParseMpLinks "osuParseMpLinks/osu_api_usage"
)

func main() {

	print("Hellow\n")
	client := osuParseMpLinks.NewHttpClient()
	data := client.GetUserDataByUsernameOrId("9109550")
	fmt.Println(data)
	mplinkData, userData := client.ParseMplink("https://osu.ppy.sh/community/matches/111534249/", osuParseMpLinks.ParsingConfig{})
	fmt.Println(mplinkData)
	fmt.Println(userData)
}
