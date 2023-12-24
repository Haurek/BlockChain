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
	ViewTimeout = 20
)

// PBFT type
type PBFT struct {
	id     string      // pBFT node ID
	engine *PBFTEngine // consensus engine
	msgLog *MsgLog     // cache of consensus messages

	net        *p2pnet.P2PNet    // network layer
	privateKey *ecdsa.PrivateKey // private key
	publicKey  *ecdsa.PublicKey  // public key
	chain      *blockchain.Chain // block chain
	blockPool  *pool.BlockPool   // block pool
	txPool     *pool.TxPool      // tx pool

	isStart   bool // flag of start
	isPrimary bool // flag of primary node
	isRunning bool // flag of running

	view         uint64 // current view
	index        uint64 // node index
	leaderIndex  uint64 // primary node index
	nodeNum      uint64 // total consensus node number
	maxFaultNode uint64 // max pBFT fault node number

	viewChangeTimer *time.Timer       // view change timer
	consensusMsg    chan *PBFTMessage // consensus message channel
	lock            sync.Mutex
	log             *log.Logger
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
	// set primary node
	pbft.leaderIndex = (v + chain.BestHeight) % pbft.nodeNum
	if pbft.leaderIndex == pbft.index {
		pbft.isPrimary = true
	}
	// initialize timer
	pbft.viewChangeTimer = time.NewTimer(ViewTimeout * time.Second)
	pbft.viewChangeTimer.Stop()
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

	// run PBFTEngine
	for {
		select {
		// receive message
		case msg := <-pbft.consensusMsg:
			pbft.log.Println("Receive a consensus message")
			if !pbft.isStart {
				continue
			}
			// run FSM handle message
			pbft.NextState(msg)
		case <-pbft.viewChangeTimer.C:
			// primary node timeout or running timeout
			// raise view change
			pbft.log.Println("View change timeout, run view change")

			pbft.lock.TryLock()
			pbft.viewChangeTimer.Stop()
			pbft.isRunning = true
			pbft.SetState(ViewChangeState)
			pbft.lock.Unlock()

			// sign
			signature, err := mycrypto.Sign(pbft.privateKey, pbft.chain.Tip)
			if err != nil {
				pbft.log.Println("Sign view change message fail")
				continue
			}
			msg := ViewChangeMessage{
				ID:        pbft.id,
				Height:    pbft.chain.BestHeight,
				BlockHash: pbft.chain.Tip,
				View:      pbft.view,
				ToView:    (pbft.view + 1) % pbft.nodeNum,
				Sign:      signature,
				PubKey:    mycrypto.PublicKey2Bytes(pbft.publicKey),
			}
			p2pmsg, err := pbft.packBroadcastMessage(ViewChangeMsg, msg)
			if err != nil {
				pbft.log.Println(err)
			}
			// broadcast
			pbft.net.Broadcast(p2pmsg)
			var pbftMsg PBFTMessage
			json.Unmarshal(p2pmsg.Data, &pbftMsg)
			// run consensus engine
			pbft.NextState(&pbftMsg)

		case <-pbft.txPool.FullSignal:
			// receive TxPool interrupt
			pbft.log.Println("TxPool full, pack into block...")
			pbft.lock.Lock()
			if !pbft.isStart || pbft.isRunning {
				// is running or not start, ignore
				pbft.lock.Unlock()
				continue
			}
			if !pbft.isPrimary {
				// not primary node
				// start viewChange timer, wait primary node prepare message
				pbft.lock.Unlock()
				pbft.viewChangeTimer.Reset(ViewTimeout * time.Second)
				continue
			}
			pbft.lock.Unlock()

			// primary node pack Txs into block and send prepare message
			msg, err := pbft.PBFTSealer()
			if err != nil {
				pbft.log.Println(err)
				continue
			}
			// serialize PBFTMessage
			serialized, err := json.Marshal(msg)
			if err != nil {
				pbft.log.Println(err)
				continue
			}
			// pack P2P network message
			p2pMessage := &p2pnet.Message{
				Type: p2pnet.ConsensusMsg,
				Data: serialized,
			}
			// broadcast
			pbft.log.Println("Broadcast prepare message")
			pbft.net.Broadcast(p2pMessage)
			time.Sleep(10 * time.Millisecond)

			// send prepare message to self engine
			pbft.NextState(msg)
		}
	}
}

// PBFTSealer primary node pack block
func (pbft *PBFT) PBFTSealer() (*PBFTMessage, error) {
	pbft.log.Println("Primary node pack block...")
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
