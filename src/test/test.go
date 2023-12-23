package main

import "fmt"

func main() {
	data := map[string][]int{
		"key1": {1},
		"key2": {10, 20, 30, 40, 50},
	}

	for key, value := range data {
		var newDate []int
		for _, item := range value {
			if item == 1 {
				continue
			}
			newDate = append(newDate, item)
		}
		data[key] = newDate
	}
	fmt.Println(data)
}
