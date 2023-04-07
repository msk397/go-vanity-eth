package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type RemoteConfig struct {
	Continuous         int      `json:"continuous"`
	DreamAddressSubstr []string `json:"dreamAddressSubstr"`
}
type LocalConfig struct {
	BarkUrl      string  `json:"barkUrl"`
	BarkKey      string  `json:"barkKey"`
	Rate         float64 `json:"rate"`
	RemoteConfig string  `json:"remoteConfig"`
}

type Config struct {
	RemoteConfig RemoteConfig `json:"remoteConfig"`
	LocalConfig  LocalConfig  `json:"localConfig"`
}

func readConfigAll() (Config, error) {
	con2, err := readLocalConfig("localConfig.json")
	if err != nil {
		return Config{}, err
	}
	localConfig := con2.LocalConfig
	remoteConfig, err := readRemoteConfig(localConfig.RemoteConfig)
	if err != nil {
		return Config{}, err
	}
	return Config{RemoteConfig: remoteConfig, LocalConfig: localConfig}, nil
}

// 读取配置文件
func readLocalConfig(path string) (Config, error) {
	// 打开json文件
	jsonFile, err := os.Open(path)
	// 最好要处理以下错误
	if err != nil {
		return Config{}, err
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var localConfig LocalConfig
	err = json.Unmarshal([]byte(byteValue), &localConfig)

	if err != nil {
		return Config{}, err
	}

	if localConfig.Rate > 1 {
		localConfig.Rate = 1
	}
	if localConfig.Rate < 0 {
		localConfig.Rate = 0
	}
	return Config{LocalConfig: localConfig}, nil
}

func readRemoteConfig(path string) (RemoteConfig, error) {
	var remoteConfig RemoteConfig
	//去指定网址下载配置文件
	// 判断RemoteConfig是网址还是本地文件
	if strings.HasPrefix(path, "http") {
		resp, err := http.Get(path)
		if err != nil {
			return RemoteConfig{}, err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return RemoteConfig{}, err
		}
		err = json.Unmarshal([]byte(body), &remoteConfig)
	} else {
		jsonFile, err := os.Open(path)
		if err != nil {
			return RemoteConfig{}, err
		}
		defer jsonFile.Close()
		byteValue, _ := ioutil.ReadAll(jsonFile)
		err = json.Unmarshal([]byte(byteValue), &remoteConfig)
		if err != nil {
			return RemoteConfig{}, err
		}
	}
	return remoteConfig, nil
}
