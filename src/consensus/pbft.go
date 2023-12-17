package consensus

import (
	"BlockChain/src/blockchain"
	p2pnet "BlockChain/src/network"
	"BlockChain/src/pool"
	"BlockChain/src/state"
	"BlockChain/src/utils"
	"crypto/ecdsa"
	"encoding/json"
	"log"
	"sync"
)

// PBFT type
// @param fsm: finite-state machine
// @param msgLog: Message msgLog
// @param msgQueue: Message queue
// @param pBFTPeers: peerID -> public key, public key used for verify
// @param privateKey: used for message sign
// @param txPool: Transaction Pool
// @param isStart: pBFT consensus start flag
// @param isPrimary: primary node flag
// @param view: current view
type PBFT struct {
	fsm              *PBFTFSM
	msgLog           *MsgLog
	net              *p2pnet.P2PNet
	ws               *state.WorldState
	msgQueue         *MessageQueue
	primaryID        string
	selfID           string
	privateKey       *ecdsa.PrivateKey
	publicKey        *ecdsa.PublicKey
	txPool           *pool.TxPool
	chain            *blockchain.Chain
	isStart          bool
	isPrimary        bool
	view             uint64
	stableCheckPoint uint64
	checkPoint       uint64
	maxFaultNode     uint64
	lock             sync.Mutex
	log              *log.Logger
}

// MessageQueue queue of message received from network
type MessageQueue struct {
	messages chan *PBFTMessage
}

// NewMessageQueue creat a message queue
func NewMessageQueue() *MessageQueue {
	return &MessageQueue{
		messages: make(chan *PBFTMessage, 100),
	}
}

func (q *MessageQueue) Enqueue(msg *PBFTMessage) {
	select {
	case q.messages <- msg:
		return
	default:
		// unable to enqueue, drop the message
		return
	}
}

func (q *MessageQueue) Dequeue() <-chan *PBFTMessage {
	return q.messages
}

// NewPBFT create pBFT engine
func NewPBFT(ws *state.WorldState, txPool *pool.TxPool, net *p2pnet.P2PNet, chain *blockchain.Chain, logPath string) (*PBFT, error) {
	// initialize logger
	l := utils.NewLogger("[pbft] ", logPath)

	var fsm *PBFTFSM
	if ws.IsPrimary {
		l.Println("Initialize primary node")
		fsm = NewFSM(RequestState)
	} else {
		l.Println("Initialize replica node")
		fsm = NewFSM(PrePrepareState)
	}
	pbft := &PBFT{
		fsm:      fsm,
		msgLog:   NewMsgLog(ws.WaterHead),
		net:      net,
		ws:       ws,
		msgQueue: NewMessageQueue(),
		//pBFTPeers: make(map[string]*ecdsa.PublicKey),
		primaryID: ws.PrimaryID,
		selfID:    net.Host.ID().String(),
		isPrimary: ws.IsPrimary,
		txPool:    txPool,
		chain:     chain,
		isStart:   false,
		log:       l,
	}

	return pbft, nil
}

// OnReceive receive callback func
func (pbft *PBFT) OnReceive(t p2pnet.MessageType, msgBytes []byte, peerID string) {
	if t != p2pnet.ConsensusMsg {
		// unable to handler
		return
	}
	// Unmarshal message byte
	var msg PBFTMessage
	err := json.Unmarshal(msgBytes, &msg)
	if err != nil {
		pbft.log.Println("Unmarshal PBFTMessage failed")
		return
	}
	// add to message queue
	pbft.msgQueue.Enqueue(&msg)
}

func (pbft *PBFT) Run() {
	pbft.log.Println("Run consensus")
	pbft.isStart = true
	// register callback func
	pbft.net.RegisterCallback(p2pnet.ConsensusMsg, pbft.OnReceive)

	// run consensus
	for {
		select {
		// receive message
		case msg := <-pbft.msgQueue.Dequeue():
			pbft.log.Println("Receive a PBFT message")
			if !pbft.isStart {
				continue
			}
			// run FSM handle message
			pbft.NextState(msg)
		}
	}
}

func (pbft *PBFT) Start() {
	pbft.lock.Lock()
	defer pbft.lock.Unlock()
	pbft.isStart = true
}

func (pbft *PBFT) Stop() {
	pbft.lock.Lock()
	defer pbft.lock.Unlock()
	pbft.isStart = false
}

func (pbft *PBFT) AddCheckPoint() {
	pbft.lock.Lock()
	defer pbft.lock.Unlock()
	pbft.checkPoint++
}
