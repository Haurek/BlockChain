package client

import (
	"BlockChain/src/blockchain"
	"BlockChain/src/consensus"
	"BlockChain/src/network"
	"BlockChain/src/pool"
	"BlockChain/src/state"
	"BlockChain/src/utils"
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Client struct {
	chain      *blockchain.Chain
	network    *p2pnet.P2PNet
	consensus  *consensus.PBFT
	worldState *state.WorldState
	wallet     *blockchain.Wallet
	txPool     *pool.TxPool
	blockPool  *pool.BlockPool
	config     *Config
	log        *log.Logger
	//startNodeDone chan struct{}
}

// CreateClient create a new client
func CreateClient(config *Config, c *blockchain.Chain, w *blockchain.Wallet) (*Client, error) {
	// initialize log
	l := utils.NewLogger("[client] ", config.ClientCfg.LogPath)

	// initialize the network
	net := p2pnet.CreateNode(config.P2PNetCfg.PriKeyPath, config.P2PNetCfg.ListenAddr, config.P2PNetCfg.LogPath)

	// initialize TxPool
	txPool := pool.NewTxPool(net, config.TxPoolCfg.LogPath)

	// initialize the world state
	ws := &state.WorldState{
		BlockHeight:  c.BestHeight,
		Tip:          c.Tip,
		IsPrimary:    config.PBFTCfg.IsPrimary,
		SelfID:       net.Host.ID().String(),
		View:         config.PBFTCfg.View,
		CheckPoint:   c.BestHeight,
		WaterHead:    config.PBFTCfg.WaterHead,
		MaxFaultNode: config.PBFTCfg.MaxFaultNode,
	}
	// initialize BlockPool
	blockPool := pool.NewBlockPool(net, c, ws, config.BlockPoolCfg.LogPath)

	// initialize the consensus
	pbft, err := consensus.NewPBFT(ws, txPool, net, c, config.PBFTCfg.LogPath)
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
		//startNodeDone: make(chan struct{}),
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
			case "q":
				close(exitChan)
				return nil
			case "help":
			case "h":
				c.Usages()
			case "balance":
			case "b":
				// TODO
				fmt.Printf("Your balance: %d\n", c.GetBalance())
			case "request":
			case "r":
				c.ProposalNewBlock()
			case "address":
				fmt.Println(hex.EncodeToString(c.wallet.GetAddress()))
			case "transaction":
			case "tx":
				if len(parts) == 3 {
					amount, err := strconv.Atoi(parts[1])
					if err != nil {
						fmt.Println("wrong amount")
					} else if toAddress, err := hex.DecodeString(parts[2]); err != nil {
						fmt.Println("wrong address")
					} else {
						c.CreateTransaction(amount, toAddress)
					}
				} else {
					fmt.Println("Please input address and amount")
				}
			case "tip":
				if b := c.chain.FindBlock(c.chain.Tip); b != nil {
					b.Show()
				} else {
					fmt.Println("Uninitialized database")
				}
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
	// TODO
	return 0
}

// ProposalNewBlock proposal new block when TxPool is full
func (c *Client) ProposalNewBlock() {
	// check TxPool
	if c.txPool.Count() == 0 {
		// tx pool is empty
		return
	}

	// get Tx from TxPool and verify Tx
	txs := c.txPool.GetTransactions()
	var packTxs []*blockchain.Transaction
	count := 0
	for id, tx := range txs {
		if !blockchain.VerifyTransaction(c.chain, tx) {
			// transaction is invalid
			c.txPool.RemoveTransaction(id)
			continue
		}
		packTxs = append(packTxs, tx)
		count++
		if count == c.config.ChainCfg.MaxTxPerBlock {
			break
		}
	}
	// pack TxsBytes into Request message
	txBytes, err := json.Marshal(packTxs)
	if err != nil {
		return
	}
	reqMsg := consensus.RequestMessage{
		Timestamp: time.Now().Unix(),
		ClientID:  c.worldState.SelfID,
		TxsBytes:  txBytes,
	}
	payload, err := json.Marshal(reqMsg)
	if err != nil {
		return
	}
	pbftMessage := consensus.PBFTMessage{
		Type: consensus.RequestMsg,
		Data: payload,
	}
	serialized, err := json.Marshal(pbftMessage)
	if err != nil {
		return
	}
	p2pMsg := p2pnet.Message{
		Type: p2pnet.ConsensusMsg,
		Data: serialized,
	}

	// send request message to primary node
	c.network.BroadcastToPeer(&p2pMsg, c.worldState.PrimaryID)
}

func (c *Client) Usages() {
	fmt.Println("Usages:")
	fmt.Println("exit, q: Exit the block chain")
	fmt.Println("balance, b: Check balance")
}
