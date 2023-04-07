package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func checkFileIsExist(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return err
	}
	return nil
}

func sendMessageBybark(title, mess, barkUrl, barkKey string) {
	if barkUrl == "" || barkKey == "" {
		return
	}
	var data = []byte(`{"body":"` + mess + `","device_key":"` + barkKey + `","title":"` + title + `"}`)
	response, err := http.Post(barkUrl, "application/json; charset=utf-8",
		strings.NewReader(string(data)))
	if err != nil {
		fmt.Println("failed to post", err)
	} else {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Println("failed to close response body", err)
			}
		}(response.Body)
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Println("failed to read response body", err)
		} else {
			fmt.Println(string(body))
		}
	}
}

func contrastConfig(config Config, con Config) {
	// 比较配置文件，把不同的地方打印出来
	if config.RemoteConfig.Continuous != con.RemoteConfig.Continuous {
		go sendMessageBybark("continuous", fmt.Sprintf("%d -> %d", config.RemoteConfig.Continuous, con.RemoteConfig.Continuous),
			config.LocalConfig.BarkUrl, config.LocalConfig.BarkKey)
	}
	if len(config.RemoteConfig.DreamAddressSubstr) != len(con.RemoteConfig.DreamAddressSubstr) {
		go sendMessageBybark("dreamAddressSubstr", fmt.Sprintf("%v -> %v", config.RemoteConfig.DreamAddressSubstr, con.RemoteConfig.DreamAddressSubstr),
			config.LocalConfig.BarkUrl, config.LocalConfig.BarkKey)
	}
	if config.LocalConfig.Rate != con.LocalConfig.Rate {
		go sendMessageBybark("rate", fmt.Sprintf("%f - > %f", config.LocalConfig.Rate, con.LocalConfig.Rate),
			config.LocalConfig.BarkUrl, config.LocalConfig.BarkKey)
	}
}
