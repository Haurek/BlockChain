package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Person struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Address string `json:"address"`
}

func main() {
	// 创建一个 Person 对象
	person := Person{
		Name:    "John Doe",
		Age:     30,
		Address: "123 Main St",
	}

	// 保存为 JSON 文件
	saveJSON("person.json", person)

	// 从 JSON 文件还原对象
	restoredPerson := restoreJSON("person.json")
	fmt.Println(restoredPerson)
}

func saveJSON(filename string, data interface{}) {
	// 将数据转换为 JSON 格式
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	// 将 JSON 数据写入文件
	err = ioutil.WriteFile(filename, jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing JSON file:", err)
		return
	}

	fmt.Println("JSON data saved to", filename)
}

func restoreJSON(filename string) Person {
	// 从文件中读取 JSON 数据
	jsonData, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return Person{}
	}

	// 将 JSON 数据解析为对象
	var restoredPerson Person
	err = json.Unmarshal(jsonData, &restoredPerson)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return Person{}
	}

	fmt.Println("JSON data restored from", filename)
	return restoredPerson
}
