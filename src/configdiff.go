/**
 * Created with IntelliJ IDEA.
 * User: mahang
 * Date: 13-11-15
 * Time: 上午10:12
 * To change this template use File | Settings | File Templates.
 */
package main

import (
	"fmt"
	"flag"
	"os"
	"strings"
)

func readFile(filepath string) []byte{
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return nil
	}
	fileSize := fileInfo.Size()

	file, err := os.Open(filepath)
	if err != nil {
		return nil
	}
	defer file.Close()
	buffer := make([]byte, fileSize)
	file.Read(buffer)
	return buffer
}
func readMap(bytes []byte) map[string]string {
	if bytes == nil {
		return nil
	}
	var keyMap = make(map[string]string)
	content := string(bytes)
	lineTemp := strings.Split(content, "\n")
	for _, line := range lineTemp {
		keyAndValueTemp := strings.Split(line, "=")
		if len(keyAndValueTemp) == 2 {
			key := strings.TrimSpace(keyAndValueTemp[0])
			value := strings.TrimSpace(keyAndValueTemp[1])
			keyMap[key] = value
		}
	}
	return keyMap
}
func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 2 {
		content := `
useage :
	go run main.go file1 file2
		`
		fmt.Println(content)
		os.Exit(1)
	}
	file1 := args[0]
	file2 := args[1]

	map1 := readMap(readFile(file1))
	map2 := readMap(readFile(file2))
	if map1 == nil || map2 == nil {
		fmt.Println("input empty")
		os.Exit(1)
	}
	added1 := []string{}
	added2 := []string{}
	for key, value := range map1 {
		if _, ok := map2[key]; !ok {
			str := key + " = " + value
			added1 = append(added1, str)
		}
	}
	for key, value := range map2 {
		if _, ok := map1[key]; !ok {
			str := key + " = " + value
			added2 = append(added2, str)
		}
	}
	fmt.Println(file1 + " + : ")
	for _, value := range added1 {
		fmt.Println(value)
	}
	fmt.Println("\n")
	fmt.Println(file2 + " + : ")
	for _, value := range added2 {
		fmt.Println(value)
	}
}
