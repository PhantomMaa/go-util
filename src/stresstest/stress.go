package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

var count int32 = 0

const (
	content = `
Useage :
	stress [-option] filePath
Options :
	-p maxCPUs : how many process to run
	-f urls.txt : file contains urls to be requested
		`
)

type Error struct {
	Err error
	Msg string
}
type Param struct {
	url    string
	params map[string]string
}

func (e *Error) Error() string {
	return e.Msg + ", " + e.Err.Error()
}

func readFile(filepath string) ([]byte, error) {
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return nil, err
	}
	fileSize := fileInfo.Size()
	if fileSize == 0 {
		return nil, &Error{nil, "file size is 0"}
	}
	file, err := os.Open(filepath)
	defer file.Close()

	if err != nil {
		return nil, err
	}
	buffer := make([]byte, fileSize)
	file.Read(buffer)
	return buffer, nil
}
func prepareParam(filepath string) ([]Param, error) {
	buffer, err := readFile(filepath)
	if err != nil {
		return nil, err
	}

	content := string(buffer)

	lineTemp := strings.Split(content, "\n")
	if lineTemp == nil || len(lineTemp) <= 0 {
		return nil, &Error{nil, "file content is empty"}
	}
	lineLength := len(lineTemp)
	params := make([]Param, 0, lineLength)
	for i := 0; i < lineLength; i++ {
		line := lineTemp[i]
		param := Param{}
		param.params = make(map[string]string)
		temp := strings.Split(line, "?")
		if len(temp) == 2 {
			param.url = temp[0]
			variables := temp[1]
			variablesTemp := strings.Split(variables, "&")
			for _, varibale := range variablesTemp {
				keyValueTemp := strings.Split(varibale, "=")
				param.params[keyValueTemp[0]] = keyValueTemp[1]
			}
			params = append(params, param)
		}
	}
	return params, nil
}
func request(client *http.Client, param Param) {
	time1 := time.Now().UnixNano()
	form := url.Values{}
	for key, value := range param.params {
		form.Add(key, value)
	}
	response, err := client.PostForm(param.url, form)
	if err != nil {
		fmt.Println(param.url)
		fmt.Println(form)
		fmt.Printf("%s\n", err)
		return
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		fmt.Printf("response.code = %d, param : %s\n", response.StatusCode, param)
	}
	atomic.AddInt32(&count, 1)
	time2 := time.Now().UnixNano()
	costTime := (time2 - time1) / (1000000)
	fmt.Printf("%d, costTime = %d\n", count, costTime)
}
func loop(params []Param) {
	var client = &http.Client{}
	for true {
		for _, param := range params {
			request(client, param)
		}
	}
}

func main() {
	var processerNum = runtime.NumCPU()
	runtime.GOMAXPROCS(processerNum)

	flag.IntVar(&processerNum, "p", processerNum, "how many processer to run")
	filePath := flag.String("f", "urls.txt", "file contains urls to be requested")

	flag.Parse()

	params, err := prepareParam(*filePath)
	if err != nil {
		fmt.Println(err)
		fmt.Println(content)
		os.Exit(1)
		return
	}
	endFlag := make(chan int)

	fmt.Printf("start %d process\n", processerNum)
	for i := 0; i < processerNum; i++ {
		go loop(params)
	}
	<-endFlag

}
