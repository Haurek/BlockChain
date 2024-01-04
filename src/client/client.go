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

// Client represents a blockchain client
type Client struct {
	isConsensus bool
	chain       *blockchain.Chain
	network     *p2pnet.P2PNet
	consensus   *consensus.PBFT
	wallet      *blockchain.Wallet
	txPool      *pool.TxPool
	blockPool   *pool.BlockPool
	config      *Config
	log         *log.Logger
}

// CreateClient creates a new client
func CreateClient(config *Config, c *blockchain.Chain, w *blockchain.Wallet) (*Client, error) {
	// initialize log
	l := utils.NewLogger("[client] ", config.ClientCfg.LogPath)

	// initialize the network
	net := p2pnet.CreateNode(config.P2PNetCfg.PriKeyPath, config.P2PNetCfg.ListenAddr, config.P2PNetCfg.LogPath)

	// initialize TxPool
	txPool := pool.NewTxPool(config.TxPoolFull, net, config.TxPoolCfg.LogPath)

	// initialize BlockPool
	blockPool := pool.NewBlockPool(config.BlockPoolFull, net, c, config.BlockPoolCfg.LogPath)

	// initialize the consensus
	pbft, err := consensus.NewPBFT(config.PBFTCfg.NodeNum, config.PBFTCfg.Index, config.PBFTCfg.MaxFaultNode, config.PBFTCfg.View, txPool, blockPool, net, c, w, config.PBFTCfg.LogPath)
	if err != nil {
		l.Panic("Initialize pBFT consensus fail")
		return nil, err
	}

	client := &Client{
		isConsensus: config.PBFTCfg.IsConsensusNode,
		chain:       c,
		network:     net,
		consensus:   pbft,
		wallet:      w,
		txPool:      txPool,
		blockPool:   blockPool,
		config:      config,
		log:         l,
	}
	return client, nil
}

// Run executes the client operations, including running the various components.
func (c *Client) Run(wg *sync.WaitGroup, exitChan chan struct{}) error {
	// Start the p2p network
	c.log.Println("Run p2p net")
	go c.network.StartNode()

	// Run pBFT consensus
	if c.isConsensus {
		c.log.Println("Run pBFT consensus")
		go c.consensus.Run()
	}

	// Run transaction pool
	c.log.Println("Run Transaction Pool")
	go c.txPool.Run()

	// Start block synchronization
	c.log.Println("Run Block Pool")
	go c.blockPool.Run()

	defer wg.Done() // Mark the WaitGroup as done when this function exits
	for {
		fmt.Print("> ")
		var input string
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			input = scanner.Text()
		}

		// Split the input string into command and arguments
		cmd := strings.Fields(input)
		if len(cmd) == 0 {
			c.Usages() // Display usage instructions if no command is entered
		} else {
			if !c.handleCmd(cmd) {
				close(exitChan)
				return nil
			}
		}
	}
}

// handleCmd handle user input command
func (c *Client) handleCmd(cmd []string) bool {
	switch cmd[0] {
	case "q":
		return false // Exit the Run function
	case "h":
		c.Usages() // Display usage instructions
	case "tx":
		// Process transaction command
		if len(cmd) == 3 {
			amount, err := strconv.Atoi(cmd[1])
			if err != nil {
				fmt.Println("wrong amount")
			} else if toAddress := []byte(cmd[2]); err != nil {
				fmt.Println("wrong address")
			} else {
				c.createTransaction(amount, toAddress) // Create a transaction
			}
		} else {
			fmt.Println("Please input address and amount")
		}
	case "s":
		c.showStatus() // Display the status
	case "b":
		if len(cmd) == 2 {
			block := c.chain.FindBlock(cmd[1])
			if block == nil {
				fmt.Println("Block not found")
			} else {
				block.Show()
			}

		} else {
			fmt.Println("Please input block id")
		}
	default:
		fmt.Println("Unknown command, use \"help\" or \"h\" for usage")
	}
	return true
}

// createTransaction creates a transaction with specified amount and recipient address,
// adds it to the local transaction pool, and broadcasts it to connected peers.
func (c *Client) createTransaction(amount int, to []byte) {
	// Create a new transaction using the client's wallet and blockchain
	tx, err := blockchain.NewTransaction(c.wallet, c.chain, to, amount)
	if err != nil {
		return // If creating transaction fails, exit function
	}

	// Add the transaction to the local transaction pool
	c.txPool.AddTransaction(tx)

	// Marshal the transaction to JSON for broadcasting
	txByte, err := json.Marshal(tx)
	if err != nil {
		c.log.Println("Marshal transaction failed")
		return // If marshaling fails, exit function
	}

	// Create a message containing the transaction and its type
	txMessage := pool.TxMessage{
		Type:    pool.SendTxMsg,
		TxBytes: txByte,
	}
	payload, err := json.Marshal(txMessage)
	if err != nil {
		c.log.Println("Marshal TxMessage failed")
		return // If marshaling fails, exit function
	}

	// Create a P2P message containing the transaction payload and type
	msg := &p2pnet.Message{
		Type: p2pnet.TransactionMsg,
		Data: payload,
	}

	// Broadcast the transaction message to connected peers
	c.network.Broadcast(msg)
}

// getBalance get balance of the client
func (c *Client) getBalance() int {
	return blockchain.GetBalanceFromSet(c.chain.DataBase, c.wallet.GetAddress())
}

// showStatus displays the status of the blockchain, wallet, network, and consensus components.
func (c *Client) showStatus() {
	// Display wallet information
	fmt.Println("Address: ", string(c.wallet.GetAddress()))
	fmt.Println("Balance: ", c.getBalance())

	// Display blockchain status
	fmt.Println("\nBlockChain Status:")
	fmt.Println("Height: ", c.chain.GetHeight())
	tip := c.chain.GetTip()
	fmt.Println("Tip: ", tip)
	headBlock := c.chain.FindBlock(tip)
	headBlock.Show()

	// Display pool status (transaction and block pools)
	fmt.Println("\nPoolStatus: ")
	fmt.Println("TxPool count: ", c.txPool.Count())
	fmt.Println("BlockPool count: ", c.blockPool.Count())

	// Display pBFT consensus status
	fmt.Println("\npBFT consensus status: ")
	if c.consensus.IsPrimary() {
		fmt.Println("Is Primary Node: true")
	} else {
		fmt.Println("Is Primary Node: false")
	}
	fmt.Println("View: ", c.consensus.GetView())

	// Display network status
	fmt.Println("\nNetwork Status: ")
	fmt.Println("Host ID: ", c.network.Host.ID().String())
}

// Usages displays usage instructions for the client's command-line interface.
func (c *Client) Usages() {
	fmt.Println("Usages:")
	fmt.Println("h:  Show usage")
	fmt.Println("q:  Exit the blockchain client")
	fmt.Println("tx: tx <amount> <address>   create new transaction")
	fmt.Println("s:  Show current status of block chain")
	fmt.Println("b:  Search block by hash or height")
}
