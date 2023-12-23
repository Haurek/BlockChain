package consensus

import (
	"BlockChain/src/blockchain"
	"BlockChain/src/mycrypto"
	p2pnet "BlockChain/src/network"
	"BlockChain/src/pool"
	"BlockChain/src/utils"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"
)

const (
	TxPoolPollingTime  = 10
	PrimaryNodeTimeout = 20
	TxPoolFull         = 1
)

// PBFT type
// @param fsm: finite-state machine
// @param msgLog: Message msgLog
// @param msgQueue: Message queue
// @param pBFTPeers: peerID -> public key, public key used for verify
// @param privateKey: used for message sign
// @param isStart: pBFT consensus start flag
// @param isPrimary: primary node flag
// @param view: current view
type PBFT struct {
	id     string
	engine *PBFTEngine
	msgLog *MsgLog

	net        *p2pnet.P2PNet
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	chain      *blockchain.Chain
	blockPool  *pool.BlockPool
	txPool     *pool.TxPool

	isStart   bool
	isPrimary bool
	isRunning bool

	view         uint64
	index        uint64
	leaderIndex  uint64
	nodeNum      uint64
	maxFaultNode uint64

	sealerTimer  *time.Timer
	primaryTimer *time.Timer
	consensusMsg chan *PBFTMessage
	lock         sync.Mutex
	log          *log.Logger
}

// NewPBFT create pBFT engine
func NewPBFT(num, index uint64, f uint64, v uint64, tp *pool.TxPool, bp *pool.BlockPool, net *p2pnet.P2PNet, chain *blockchain.Chain, wallet *blockchain.Wallet, logPath string) (*PBFT, error) {
	// initialize logger
	l := utils.NewLogger("[pbft] ", logPath)

	pbft := &PBFT{
		engine:       NewEngine(),
		msgLog:       NewMsgLog(),
		net:          net,
		chain:        chain,
		publicKey:    wallet.GetPublicKey(),
		privateKey:   wallet.GetPrivateKey(),
		blockPool:    bp,
		txPool:       tp,
		isStart:      false,
		isRunning:    false,
		view:         v,
		nodeNum:      num,
		index:        index,
		maxFaultNode: f,
		log:          l,
		consensusMsg: make(chan *PBFTMessage),
	}

	pbft.id = string(wallet.GetAddress())
	pbft.leaderIndex = (v + chain.BestHeight) % pbft.nodeNum
	if pbft.leaderIndex == pbft.index {
		pbft.isPrimary = true
	}
	pbft.sealerTimer = time.NewTimer(TxPoolPollingTime * time.Second)
	pbft.primaryTimer = time.NewTimer(PrimaryNodeTimeout * time.Second)
	pbft.sealerTimer.Stop()
	pbft.primaryTimer.Stop()

	return pbft, nil
}

// OnReceive receive callback func
func (pbft *PBFT) OnReceive(t p2pnet.MessageType, msgBytes []byte, peerID string) {
	if t != p2pnet.ConsensusMsg {
		return
	}
	// Unmarshal message byte
	var msg PBFTMessage
	err := json.Unmarshal(msgBytes, &msg)
	if err != nil {
		pbft.log.Println("Unmarshal PBFTMessage fail")
		return
	}
	// run consensus engine
	pbft.consensusMsg <- &msg
}

func (pbft *PBFT) Run() {
	pbft.log.Println("Run consensus")
	pbft.isStart = true
	// register callback func
	pbft.net.RegisterCallback(p2pnet.ConsensusMsg, pbft.OnReceive)

	if pbft.isPrimary {
		pbft.sealerTimer.Reset(TxPoolPollingTime * time.Second)
	} else {
		pbft.primaryTimer.Reset(PrimaryNodeTimeout * time.Second)
	}

	// run PBFTEngine
	for {
		select {
		// receive message
		case msg := <-pbft.consensusMsg:
			pbft.log.Println("Receive a PBFT message")
			if !pbft.isStart {
				continue
			}
			// run FSM handle message
			pbft.NextState(msg)
		case <-pbft.primaryTimer.C:
			// primary node timeout
			if pbft.isPrimary || !pbft.isStart {
				pbft.primaryTimer.Reset(PrimaryNodeTimeout * time.Second)
				continue
			} else {
				// raise view change
			}
		case <-pbft.sealerTimer.C:
			pbft.lock.Lock()
			if !pbft.isPrimary {
				// not primary node, stop timer
				pbft.lock.Unlock()
				pbft.sealerTimer.Stop()
				continue
			}
			if !pbft.isStart || pbft.isRunning {
				// is running or not start, reset
				pbft.lock.Unlock()
				pbft.sealerTimer.Reset(10 * time.Second)
				continue
			}
			pbft.lock.Unlock()

			// check TxPool
			if pbft.txPool.Count() >= TxPoolFull {
				// pack block into prepare message
				msg, err := pbft.PBFTSealer()
				if err != nil {
					pbft.log.Println(err)
					pbft.sealerTimer.Reset(10 * time.Second)
					continue
				}
				// serialize PBFTMessage
				serialized, err := json.Marshal(msg)
				if err != nil {
					pbft.log.Println(err)
					pbft.sealerTimer.Reset(10 * time.Second)
					continue
				}
				// pack P2P network message
				p2pMessage := &p2pnet.Message{
					Type: p2pnet.ConsensusMsg,
					Data: serialized,
				}
				pbft.net.Broadcast(p2pMessage)
				time.Sleep(10 * time.Millisecond)

				// run PBFTEngine
				pbft.NextState(msg)
			} else {
				pbft.sealerTimer.Reset(10 * time.Second)
			}
		}
	}
}

// PBFTSealer primary node pack block
func (pbft *PBFT) PBFTSealer() (*PBFTMessage, error) {
	// get txs from pool
	txMap := pbft.txPool.GetTransactions()
	var txs []*blockchain.Transaction
	for _, tx := range txMap {
		txs = append(txs, tx)
	}

	// pack block
	newBlock := blockchain.NewBlock(pbft.chain.Tip, txs, pbft.chain.BestHeight+1)
	blockData, err := json.Marshal(newBlock)
	if err != nil {
		return nil, errors.New("Marshal block data fail")
	}

	// sign
	signature, err := mycrypto.Sign(pbft.privateKey, newBlock.Header.Hash)
	if err != nil {
		return nil, errors.New("Sign message fail")
	}

	// pack PrePrepareMessage
	prepare := PrepareMessage{
		ID:        pbft.id,
		Height:    newBlock.Header.Height,
		BlockHash: newBlock.Header.Hash,
		Block:     blockData,
		View:      pbft.view,
		Sign:      signature,
		PubKey:    mycrypto.PublicKey2Bytes(pbft.publicKey),
	}

	payload, err := json.Marshal(prepare)
	if err != nil {
		return nil, errors.New("marshal error")
	}
	pbftMessage := &PBFTMessage{
		Type: PrepareMsg,
		Data: payload,
	}
	return pbftMessage, nil
}

// Start sets the PBFT process as started by setting the corresponding flag.
func (pbft *PBFT) Start() {
	pbft.lock.Lock()
	defer pbft.lock.Unlock()
	pbft.isStart = true // Set the PBFT process as started
}

// Stop sets the PBFT process as stopped by setting the corresponding flag.
func (pbft *PBFT) Stop() {
	pbft.lock.Lock()
	defer pbft.lock.Unlock()
	pbft.isStart = false // Set the PBFT process as stopped
}

// GetView retrieves the current view of the PBFT process.
func (pbft *PBFT) GetView() uint64 {
	pbft.lock.Lock()
	defer pbft.lock.Unlock()
	return pbft.view // Retrieve the current view
}

// IsPrimary checks if the PBFT node is acting as a primary node.
func (pbft *PBFT) IsPrimary() bool {
	pbft.lock.Lock()
	defer pbft.lock.Unlock()
	return pbft.isPrimary // Check if the PBFT node is a primary node
}
