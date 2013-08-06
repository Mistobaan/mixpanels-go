package main

import mixpanel "github.com/mixpanel/mixpanel-go" 

const token string = "e919dea023855e3c8e2ea46a38e4032c"

func main() {
	mp := mixpanel.NewMixpanel(token)
	mp.Track("user_id", "clicked button", nil)
	mp.Track("user_id", "Enter Login", &mixpanel.P{
		"test":  "other",
		"test2": "maybe",
	})

	mp.Alias("user_id", "my_used_id")
}
