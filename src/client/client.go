package client

import (
	"BlockChain/src/blockchain"
	"BlockChain/src/consensus"
	"BlockChain/src/network"
	"BlockChain/src/pool"
	"BlockChain/src/utils"
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Client struct {
	chain     *blockchain.Chain
	network   *p2pnet.P2PNet
	consensus *consensus.PBFT
	wallet    *blockchain.Wallet
	txPool    *pool.TxPool
	blockPool *pool.BlockPool
	config    *Config
	log       *log.Logger
}

// CreateClient create a new client
func CreateClient(config *Config, c *blockchain.Chain, w *blockchain.Wallet) (*Client, error) {
	// initialize log
	l := utils.NewLogger("[client] ", config.ClientCfg.LogPath)

	// initialize the network
	net := p2pnet.CreateNode(config.P2PNetCfg.PriKeyPath, config.P2PNetCfg.ListenAddr, config.P2PNetCfg.LogPath)

	// initialize TxPool
	txPool := pool.NewTxPool(net, config.TxPoolCfg.LogPath)

	// initialize BlockPool
	blockPool := pool.NewBlockPool(net, c, config.BlockPoolCfg.LogPath)

	// initialize the consensus
	pbft, err := consensus.NewPBFT(config.PBFTCfg.NodeNum, config.PBFTCfg.Index, config.PBFTCfg.MaxFaultNode, config.PBFTCfg.View, txPool, blockPool, net, c, w, config.PBFTCfg.LogPath)
	if err != nil {
		l.Panic("Initialize pBFT consensus fail")
		return nil, err
	}
	client := &Client{
		chain:     c,
		network:   net,
		consensus: pbft,
		wallet:    w,
		txPool:    txPool,
		blockPool: blockPool,
		config:    config,
		log:       l,
	}
	return client, nil
}

// Run the client
func (c *Client) Run(wg *sync.WaitGroup, exitChan chan struct{}) error {
	// run p2p net
	c.log.Println("Run p2p net")
	go c.network.StartNode()

	// run pBFT consensus
	c.log.Println("Run pBFT consensus")
	go c.consensus.Run()

	// run transaction pool
	c.log.Println("Run Transaction Pool")
	go c.txPool.Run()

	// start block sync
	c.log.Println("Run Block Pool")
	go c.blockPool.Run()

	var cmd string
	defer wg.Done()
	for {
		fmt.Println("\nPlease input your command:")
		var input string
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			input = scanner.Text()
		}

		// 将输入字符串按空格拆分为命令和参数
		parts := strings.Fields(input)
		if len(parts) == 0 {
			c.Usages()
		} else {
			cmd = parts[0]
			switch cmd {
			case "exit":
				fallthrough
			case "q":
				close(exitChan)
				return nil
			case "help":
				fallthrough
			case "h":
				c.Usages()

			case "transaction":
				fallthrough
			case "tx":
				if len(parts) == 3 {
					amount, err := strconv.Atoi(parts[1])
					if err != nil {
						fmt.Println("wrong amount")
					} else if toAddress := []byte(parts[2]); err != nil {
						fmt.Println("wrong address")
					} else {
						c.CreateTransaction(amount, toAddress)
					}
				} else {
					fmt.Println("Please input address and amount")
				}

			case "status":
				fallthrough
			case "s":
				c.ShowStatus()
			default:
				fmt.Println("Unknown command, use \"help\" or \"h\" for usage")
			}
		}
	}
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
	txByte, err := json.Marshal(tx)
	if err != nil {
		c.log.Println("Marshal transaction failed")
		return
	}

	txMessage := pool.TxMessage{
		Type:    pool.SendTxMsg,
		TxBytes: txByte,
	}
	payload, err := json.Marshal(txMessage)
	if err != nil {
		c.log.Println("Marshal TxMessage failed")
		return
	}

	msg := &p2pnet.Message{
		Type: p2pnet.TransactionMsg,
		Data: payload,
	}
	c.network.Broadcast(msg)
}

// GetBalance get balance of the client
func (c *Client) GetBalance() int {
	return blockchain.GetBalanceFromSet(c.chain.DataBase, c.wallet.GetAddress())
}

// ShowStatus show status of the blockchain
func (c *Client) ShowStatus() {
	fmt.Println("Address: ", string(c.wallet.GetAddress()))
	fmt.Println("Balance: ", c.GetBalance())

	fmt.Println("\nBlockChain Status:")
	fmt.Println("Height: ", c.chain.GetHeight())
	fmt.Println("Tip: ", c.chain.GetTip())

	fmt.Println("\nNetwork Status: ")
	fmt.Println("Host ID: ", c.network.Host.ID().String())

	fmt.Println("\nPoolStatus: ")
	fmt.Println("TxPool count: ", c.txPool.Count())
	fmt.Println("BlocKPool count: ", c.blockPool.Count())

	fmt.Println("\npBFT consensus status: ")
	if c.consensus.IsPrimary() {
		fmt.Println("Is Primary Node: true")
	} else {
		fmt.Println("Is Primary Node: false")
	}
	fmt.Println("View: ", c.consensus.GetView())
}

func (c *Client) Usages() {
	fmt.Println("Usages:")
	fmt.Println("exit, q: Exit the block chain")
	fmt.Println("balance, b: Check balance")
}
