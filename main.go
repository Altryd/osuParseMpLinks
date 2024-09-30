package main

import (
	osuParseMpLinks "osuParseMpLinks/osu_api_usage"
)

func main() {

	print("Hellow\n")
	client := osuParseMpLinks.NewHttpClient()
	data := client.GetUserDataByUsernameOrId("9109550")
	print(data)

}
