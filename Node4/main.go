package main

import (
	"BlockChain/src/blockchain"
	"BlockChain/src/client"
	"fmt"
	"sync"
)

func main() {
	// create client
	//config, err := client.LoadConfig("./config.json")
	config, err := client.LoadConfig("./Node4/debug.json")
	if err != nil {
		fmt.Println("load config fail")
		return
	}

	// create wallet
	wallet, err := blockchain.LoadWallet(config.WalletCfg.PubKeyPath, config.WalletCfg.PriKeyPath)
	if err != nil {
		fmt.Println("Create wallet...")
		// load wallet fail
		wallet = blockchain.CreateWallet()
		// create new wallet
		err = wallet.SaveWallet(config.WalletCfg.PubKeyPath, config.WalletCfg.PriKeyPath)
		if err != nil {
			fmt.Println("Create wallet fail")
			return
		}
	}

	// initialize the chain
	//chain, err := blockchain.CreateChain(wallet.GetAddress(), config.ChainCfg.ChainDataBasePath, config.ChainCfg.LogPath)
	chain, err := blockchain.LoadChain(config.ChainCfg.ChainDataBasePath, config.ChainCfg.LogPath)
	if err != nil {
		fmt.Println("Create chain fail")
		return
	}

	// create client
	c, err := client.CreateClient(config, chain, wallet)
	if err != nil {
		fmt.Println("Create client fail")
		return
	}

	// run client
	var wg sync.WaitGroup
	var exitChan = make(chan struct{})
	wg.Add(1)
	go c.Run(&wg, exitChan)

	<-exitChan
	wg.Wait()
}
