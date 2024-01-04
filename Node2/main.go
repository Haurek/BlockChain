package main

import (
	"BlockChain/src/blockchain"
	"BlockChain/src/client"
	"sync"
)

func main() {
	// create client
	//config, err := client.LoadConfig("./config.json")
	config, err := client.LoadConfig("./Node2/debug.json")
	if err != nil {
		return
	}

	// create wallet
	wallet, err := blockchain.LoadWallet(config.WalletCfg.PubKeyPath, config.WalletCfg.PriKeyPath)
	if err != nil {
		// load wallet fail
		wallet = blockchain.CreateWallet()
		// create new wallet
		wallet.SaveWallet(config.WalletCfg.PubKeyPath, config.WalletCfg.PriKeyPath)
	}

	// initialize the chain
	//chain, err := blockchain.CreateChain(wallet.GetAddress(), config.ChainCfg.ChainDataBasePath, config.ChainCfg.LogPath)
	chain, err := blockchain.LoadChain(config.ChainCfg.ChainDataBasePath, config.ChainCfg.LogPath)
	if err != nil {
		return
	}

	// create client
	c, err := client.CreateClient(config, chain, wallet)
	if err != nil {
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
