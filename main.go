package main

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"math"
	"os"
	"reflect"
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

	con, err := readConfigAll()
	if err != nil {
		fmt.Println("读取配置文件失败")
		return
	}

	threadNum := int(math.Floor(float64(CPUNum-1) * con.LocalConfig.Rate))
	fmt.Println("本次生成协程数:", threadNum)
	wg.Add(1)
	sendMessageBybark("开始生成", "开始生成", con.LocalConfig.BarkUrl, con.LocalConfig.BarkKey)
	for i := 0; i < threadNum; i++ {
		go createWallet(con, threadNumChan)
		genNum++
	}

	go DynamicSetThreadNum(con, threadNumChan, CPUNum)
	wg.Wait()
}

func DynamicSetThreadNum(con Config, threadNumChan chan struct{}, CPUNum int) {
	tmpConfig := Config{}
	for {
		select {
		case <-time.After(time.Second * 30):
			//获取时间
			time := time.Now().Format("2006-01-02 15:04:05")
			fmt.Println(time, " 当前生成使用的协程数:", genNum)
			con, err := readConfigAll()
			if err != nil {
				fmt.Println("读取配置文件失败")
				return
			}

			//比较两个结构体是否相等
			//如果tmpConfig为空，说明是第一次读取
			if reflect.DeepEqual(tmpConfig, con) || reflect.DeepEqual(tmpConfig, Config{}) {
				if reflect.DeepEqual(tmpConfig, Config{}) {
					tmpConfig = con
				}
				continue
			}

			fmt.Println("配置文件发生变化")
			threadNum := int(math.Floor(float64(CPUNum-1) * con.LocalConfig.Rate))
			if !reflect.DeepEqual(tmpConfig.RemoteConfig, con.RemoteConfig) {
				//停止所有协程
				for genNum > 0 {
					genNum--
					threadNumChan <- struct{}{}
				}

				for i := 0; i < threadNum; i++ {
					genNum++
					go createWallet(con, threadNumChan)
				}
			}
			if tmpConfig.LocalConfig != con.LocalConfig {
				needNum := threadNum - genNum
				fmt.Printf("DynamicSetThreadNum 本次需要生成协程数: %d,还需生成: %d\n", threadNum, needNum)

				if needNum > 0 {
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
			go contrastConfig(tmpConfig, con)
			tmpConfig = con
		}
	}
}

func createWallet(con Config, threadNumChan <-chan struct{}) {
	var f *os.File
	f, _ = os.OpenFile(filename, os.O_APPEND, 0666) //打开文件
	defer f.Close()
	strLength := con.RemoteConfig.Continuous
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
		for _, valueStr := range con.RemoteConfig.DreamAddressSubstr {
			//后缀是valueStr
			if strings.HasSuffix(address, valueStr) {
				isGood = true
				break
			}
		}
		if isGood {
			mutex.Lock()
			fmt.Println(address)
			go sendMessageBybark("生成一个Address", address, con.LocalConfig.BarkUrl, con.LocalConfig.BarkKey)
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
