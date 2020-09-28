package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)



var wg sync.WaitGroup
var url string


type Fingers struct {
	RECORDS []struct {
		Cmsname   string `json:"cmsname"`
		Staticurl string `json:"staticurl"`
		Checksum  string `json:"checksum"`
		Homeurl   string `json:"homeurl"`
		Keyword   string `json:"keyword"`
		Cookie    string `json:"Cookie"`
		Type      string `json:"type"`
		Remark    string `json:"remark"`
	} `json:"RECORDS"`
}

type Finger struct {
	Cmsname   string `json:"cmsname"`
	Staticurl string `json:"staticurl"`
	Checksum  string `json:"checksum"`
	Homeurl   string `json:"homeurl"`
	Keyword   string `json:"keyword"`
	Cookie    string `json:"Cookie"`
	Type      string `json:"type"`
	Remark    string `json:"remark"`
}

type JsonStruct struct {
}

func NewJsonStruct() *JsonStruct {
	return &JsonStruct{}
}

// 解析json
func (jst *JsonStruct) Load(filename string, v interface{}) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, v)
	if err != nil {
		return
	}
}

func main() {
	args := os.Args
	url = args[1]
	fmt.Println("Load finger dict....")
	JsonParse := NewJsonStruct()
	v := Fingers{}
	JsonParse.Load("./finger.json", &v)

	check := make(chan Finger, len(v.RECORDS))
	for _,data := range v.RECORDS{
		check <- data
	}

	// 开始扫描
	fmt.Println("Scanning....")
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go work(check)
	}
	wg.Wait()
	fmt.Println("Scan End!")
}

func work(check chan Finger){
	for len(check) > 0 {
		data := <- check
		md5Scan(data)
		containsScan(data)
		cookieScan(data)
	}
	wg.Done()
}


// 通过静态文件MD5进行检查
func md5Scan(data Finger){
	if data.Staticurl != "" {
		client := http.Client{
			Timeout: time.Duration(3) * time.Second,
		}
		request, err := http.NewRequest("GET", url + data.Staticurl, nil)
		if err != nil {
			return
		}
		resp, err := client.Do(request)
		if err != nil {
			return
		}
		body, err:= ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}

		h := md5.New()
		h.Write(body)
		if data.Checksum == hex.EncodeToString(h.Sum(nil)){
			fmt.Println("Md5 ==> " + data.Cmsname)
			os.Exit(1)
		}
	}
}

// 通过文件包含字符串进行检查
func containsScan(data Finger){
	if data.Homeurl != "" {
		client := http.Client{
			Timeout: time.Duration(3) * time.Second,
		}
		request, err := http.NewRequest("GET", url + data.Homeurl, nil)
		if err != nil {
			return
		}
		resp, err := client.Do(request)
		if err != nil {
			return
		}
		body, err:= ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}

		if  resp.StatusCode == 200 &&  strings.Contains(string(body), data.Keyword) {
			fmt.Println("Contains ==> " + data.Cmsname)
			os.Exit(1)
			return
		}
	}
}

// 通过cookie进行检查
func cookieScan(data Finger){
	if data.Cookie != "" {
		client := http.Client{
			Timeout: time.Duration(3) * time.Second,
		}
		request, err := http.NewRequest("HEAD", url, nil)
		if err != nil {
			return
		}
		resp, err := client.Do(request)
		if err != nil {
			return
		}
		for _,value := range(resp.Cookies()){
			if  value.Name == data.Cookie {
				fmt.Println("Cookies ==> " + data.Cmsname)
				os.Exit(1)
				return
			}
		}
	}
}
