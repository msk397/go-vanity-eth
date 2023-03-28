package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
)

var (
	filename = "./wallet.txt"
	mutex    sync.Mutex
	wg       sync.WaitGroup
)

func main() {
	var f *os.File
	CPUNum := runtime.NumCPU()
	if checkFileIsExist(filename) { //如果文件存在
		f, _ = os.OpenFile(filename, os.O_APPEND, 0666) //打开文件
		fmt.Println("文件存在")
	} else {
		f, _ = os.Create(filename) //创建文件
		fmt.Println("文件不存在")
	}

	fmt.Println("本程序会自动尝试生成符合config.json要求的eth钱包地址，可能需要几天到几周的时间")
	fmt.Println("cpu内核数量:", CPUNum)
	fmt.Println("你可以在多个电脑上运行本程序加快速度")
	fmt.Println("开始生成……")

	f.Close()
	// 打开json文件
	jsonFile, err := os.Open("config.json")

	// 最好要处理以下错误
	if err != nil {
		fmt.Println("config.json文件不存在，请查看该文件")
	} else {
		byteValue, _ := ioutil.ReadAll(jsonFile)
		var con config

		err := json.Unmarshal([]byte(byteValue), &con)
		if err != nil {
			fmt.Println("config.json文件错误，请查看该文件")
		}
		jsonFile.Close()
		threadNum := CPUNum - 1
		wg.Add(1)
		sendMessageBybark("开始生成", "开始生成", con.BarkUrl, con.BarkKey)
		for i := 0; i < threadNum; i++ {
			go createWallet(con)
		}
		wg.Wait()
	}
}

func createWallet(con config) {
	var f *os.File
	f, _ = os.OpenFile(filename, os.O_APPEND, 0666) //打开文件
	str_length := con.Continuous
	for {
		privateKey, err := crypto.GenerateKey()
		if err != nil {
			log.Fatal(err)
		}

		publicKey := privateKey.Public()
		publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
		if !ok {
			log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
		}

		address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
		isGood := false
		endstr := address[42-str_length : 42]
		if strings.Count(endstr, string(endstr[0])) >= str_length {
			isGood = true
		}
		for _, valueStr := range con.DreamAddressSubstr {
			//后缀是valueStr
			if strings.HasSuffix(address, valueStr) {
				isGood = true
				break
			}
		}
		if isGood {
			mutex.Lock()
			fmt.Println(address)
			sendMessageBybark("生成一个Address", address, con.BarkUrl, con.BarkKey)
			privateKeyBytes := crypto.FromECDSA(privateKey)
			fmt.Println(hexutil.Encode(privateKeyBytes)[2:])
			f.WriteString(address)
			f.WriteString("\n")
			f.WriteString(hexutil.Encode(privateKeyBytes)[2:])
			f.WriteString("\n")
			f.WriteString("\n")
			f.Sync()
			mutex.Unlock()
		}

	}
}

func checkFileIsExist(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}

func sendMessageBybark(title, mess, barkUrl, barkKey string) {
	/*            data=json.dumps ({
	              "body": mess,
	              "device_key": barkKey,
	              "title": title,
	          })*/
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
