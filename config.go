package main

type config struct {
	Continuous         int      `json:"continuous"`
	DreamAddressSubstr []string `json:"dreamAddressSubstr"`
	BarkUrl            string   `json:"barkUrl"`
	BarkKey            string   `json:"barkKey"`
}
