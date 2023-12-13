package main

import "BlockChain/src/client"

func main() {
	// create client
	config, err := client.LoadConfig("./client/config.json")
	if err != nil {
		return
	}
	c, err := client.CreateClient(config)
	if err != nil {
		return
	}
	c.Run()

	// run cmd
	// TODO
}
