package client

import (
	"BlockChain/src/blockchain"
	"BlockChain/src/consensus"
	"BlockChain/src/network"
	"BlockChain/src/pool"
	"BlockChain/src/state"
	"BlockChain/src/utils"
	"encoding/json"
)

type Client struct {
	chain      *blockchain.Chain
	network    *p2pnet.P2PNet
	consensus  *consensus.PBFT
	worldState *state.WorldState
	wallet     *blockchain.Wallet
	txPool     *pool.TxPool
}

// CreateClient create a new client
func CreateClient(config *Config) (*Client, error) {
	// create wallet
	var wallet *blockchain.Wallet
	if wallet, err := blockchain.LoadWallet(config.WalletCfg.PubKeyPath, config.WalletCfg.PriKeyPath); err != nil {
		// load wallet fail
		wallet = blockchain.CreateWallet()
		// create new wallet
		wallet.SaveWallet(config.WalletCfg.PubKeyPath, config.WalletCfg.PriKeyPath)
	}
	// initialize the world state
	ws := state.NewWorldState()

	// initialize the chain
	chain, err := blockchain.LoadChain(config.ChainCfg.ChainDataBasePath)
	if err != nil {
		return nil, err
	}

	// initialize the network
	net, err := p2pnet.CreateNode(config.P2PNetCfg.PriKeyPath, config.P2PNetCfg.ListenAddr, config.P2PNetCfg.Bootstrap, config.P2PNetCfg.BootstrapPeers)
	if err != nil {
		return nil, err
	}

	// initialize TxPool
	txPool, err := pool.NewTxPool(net)
	if err != nil {
		return nil, err
	}

	// initialize the consensus
	pbft, err := consensus.NewPBFT(ws, txPool, net)
	if err != nil {
		return nil, err
	}
	client := &Client{
		chain:     chain,
		network:   net,
		consensus: pbft,
		wallet:    wallet,
		txPool:    txPool,
	}
	return client, nil
}

// Run the client
func (c *Client) Run() error {
	// run p2p net
	go c.network.StartNode()
	// run pBFT consensus
	go c.consensus.Run()
	// run transaction pool
	go c.txPool.Run()
	return nil
}

// CreateTransaction create a transaction and broadcast to peers
func (c *Client) CreateTransaction(amount int, to []byte) {
	tx, err := blockchain.NewTransaction(c.wallet, c.chain, to, amount)
	if err != nil {
		return
	}
	// add Tx to local TxPool
	c.txPool.AddTransaction(tx)

	// broadcast transaction to peers
	txByte, err := utils.Serialize(tx)
	if err != nil {
		return
	}

	txMessage := pool.TxMessage{
		Type:    pool.SendTxMsg,
		TxBytes: txByte,
	}
	payload, err := json.Marshal(txMessage)
	if err != nil {
		return
	}

	msg := &p2pnet.Message{
		Type: p2pnet.TransactionMsg,
		Data: payload,
	}
	c.network.Broadcast(msg)
}

// GetBalance get balance of the client
func (c *Client) GetBalance() {
	// TODO
}

// ProposalNewBlock proposal new block when TxPool is full
func (c *Client) ProposalNewBlock() {
	// check TxPool

	// get Tx from TxPool and verify Tx

	// pack TxsBytes into Request message

	// send request message to primary node

}
