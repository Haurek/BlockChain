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
func CreateClient() (*Client, error) {
	// initialize the chain

	// initialize the network

	// initialize the consensus

	// initialize the world state
	return nil, nil
}

// Run the client
func (c *Client) Run() error {
	// run p2p net

	// run pBFT consensus

	// run transaction pool

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
