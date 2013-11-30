package main

import (
	"net/http"
	"fmt"
	"os"
	"encoding/json"
	"net/url"
	"time"
	"runtime"
	"sync/atomic"
	"flag"
	"strconv"
)

const paramFile = "param.json"
const variableFile = "variable.json"

var count int32 = 0

type Error struct {
	Err error
	Msg string
}
type Param struct {
	url string
	params map[string]interface{}
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
func prepareParam(filepath string) (*Param, error) {
	buffer, err := readFile(filepath)
	if err != nil {
		return nil, err
	}
	var jsonData interface{}
	err = json.Unmarshal(buffer, &jsonData)
	if err != nil {
		return nil, err
	}
	jsonMap, ok := jsonData.(map[string]interface{})
	if !ok {
		return nil, &Error{nil, "转化成json格式出错"}
	}
	param := Param{}
	param.url, _ = jsonMap["url"].(string)
	param.params, _ = jsonMap["params"].(map[string]interface{})
	return &param, nil
}
func prepareVariable(filepath string) ([]interface{}, error) {
	buffer, err := readFile(filepath)
	if err != nil {
		return nil, err
	}
	var jsonData interface{}
	err = json.Unmarshal(buffer, &jsonData)
	if err != nil {
		return nil, err
	}
	jsonMap, ok := jsonData.(map[string]interface{})
	if !ok {
		return nil, &Error{nil, "转化成json格式出错"}
	}
	variables, ok := jsonMap["variables"].([]interface{})
	if !ok {
		return nil, &Error{nil, "转化variables格式出错"}
	}
	return variables, nil
}
func request(client *http.Client, param Param) {
	time1 := time.Now().UnixNano()
	form := url.Values{}
	for key, value := range param.params {
		form.Add(key, value.(string))
	}
	response, _ := client.PostForm(param.url, form)
	defer response.Body.Close()
	if response.StatusCode != 200 {
		fmt.Printf("response.code = %d, param : %s\n", response.StatusCode, param)
	}
	atomic.AddInt32(&count, 1)
	time2 := time.Now().UnixNano()
	costTime := (time2 - time1) / (1000000)
	fmt.Printf("%d, costTime = %d\n", count, costTime)
}

func loop(param Param, variables []interface{}) {
	var client = &http.Client{}
	for true {
		if variables != nil {
			for _, value := range variables {
				valueMap, ok := value.(map[string]interface {})
				if ok {
					param.params = valueMap
					request(client, param)
				}
			}
		} else {
			request(client, param)
		}
	}
}
func main() {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)

	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		content := `
useage :
	go run main.go parallelNum
		`
		fmt.Println(content)
		os.Exit(1)
	}
	parallelNum, _ := strconv.Atoi(args[0])
	endFlag := make(chan int)
	param, err := prepareParam(paramFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		return
	}
	variables, err := prepareVariable(variableFile)
	if err != nil || variables == nil {
		fmt.Println("没有配置请求参数变量，将以固定的参数发送请求")
	}
	for i := 0; i < parallelNum; i++ {
		go loop(*param, variables)
	}
	<- endFlag

}
