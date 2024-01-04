package main

import "fmt"

type test struct {
	s string
}

func main() {
	views := make([]map[string]*test, 4)
	for i, _ := range views {
		views[i] = make(map[string]*test)
	}
	to := 1
	id := "id"
	t := &test{
		s: "123",
	}
	views[to][id] = t
	fmt.Printf("%v", views[to][id])
}
