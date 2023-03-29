package main

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	filename = "./wallet.txt"
	mutex    sync.Mutex
	wg       sync.WaitGroup
	genNum   int
)

func main() {
	var f *os.File
	CPUNum := runtime.NumCPU()
	threadNumChan := make(chan struct{}, CPUNum)
	if err := checkFileIsExist(filename); err == nil { //如果文件存在
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

	con, err := readConfig()
	if err != nil {
		fmt.Println("读取配置文件失败")
		return
	}

	threadNum := int(math.Floor(float64(CPUNum-1) * con.Rate))
	fmt.Println("本次生成协程数:", threadNum)
	wg.Add(1)
	sendMessageBybark("开始生成", "开始生成", con.BarkUrl, con.BarkKey)
	for i := 0; i < threadNum; i++ {
		go createWallet(con, threadNumChan)
		genNum++
	}

	go DynamicSetThreadNum(con, threadNumChan, CPUNum)
	wg.Wait()
}

func DynamicSetThreadNum(con config, threadNumChan chan struct{}, CPUNum int) {
	tmpRate := con.Rate
	for {
		select {
		case <-time.After(time.Second * 30):
			fmt.Println("GetThreadNum 当前生成使用的协程数:", genNum)
			con, err := readConfig()
			if err != nil {
				fmt.Println("读取配置文件失败")
				return
			}

			if tmpRate == con.Rate {
				continue
			}
			tmpRate = con.Rate

			threadNum := int(math.Floor(float64(CPUNum-1) * con.Rate))
			needNum := threadNum - genNum
			fmt.Println("DynamicSetThreadNum 当前生成地址的协程数:", genNum)
			fmt.Printf("DynamicSetThreadNum 本次需要生成线程数:%d,还需生成%d\n", threadNum, needNum)

			if needNum > 0 {
				fmt.Println("线程数不足，开始创建新线程")
				for i := 0; i < needNum; i++ {
					genNum++
					go createWallet(con, threadNumChan)
				}
			} else if needNum < 0 {
				for i := 0; i > needNum; i-- {
					genNum--
					threadNumChan <- struct{}{}
				}
			}
		}
	}
}

func createWallet(con config, threadNumChan <-chan struct{}) {
	var f *os.File
	f, _ = os.OpenFile(filename, os.O_APPEND, 0666) //打开文件
	defer f.Close()
	strLength := con.Continuous
	for {
		select {
		case <-threadNumChan:
			runtime.Goexit()
		default:
		}
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
		endstr := address[42-strLength : 42]
		if strings.Count(endstr, string(endstr[0])) >= strLength {
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
			go sendMessageBybark("生成一个Address", address, con.BarkUrl, con.BarkKey)
			privateKeyBytes := crypto.FromECDSA(privateKey)
			//fmt.Println(hexutil.Encode(privateKeyBytes)[2:])
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
