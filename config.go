package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type config struct {
	Continuous         int      `json:"continuous"`
	DreamAddressSubstr []string `json:"dreamAddressSubstr"`
	BarkUrl            string   `json:"barkUrl"`
	BarkKey            string   `json:"barkKey"`
	Rate               float64  `json:"rate"`
}

// 读取配置文件
func readConfig() (config, error) {
	// 打开json文件
	jsonFile, err := os.Open("config.json")

	// 最好要处理以下错误
	if err != nil {
		return config{}, err
	} else {
		defer jsonFile.Close()
		byteValue, _ := ioutil.ReadAll(jsonFile)
		var con config

		err := json.Unmarshal([]byte(byteValue), &con)

		if err != nil {
			return config{}, err
		}

		if con.Rate > 1 {
			con.Rate = 1
		}
		if con.Rate < 0 {
			con.Rate = 0
		}
		return con, nil
	}
}
