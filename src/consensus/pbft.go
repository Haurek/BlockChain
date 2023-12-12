package consensus

import (
	p2pnet "BlockChain/src/network"
	"BlockChain/src/pool"
	"BlockChain/src/state"
	"encoding/json"
	"sync"
	"time"
)

// PBFT type
type PBFT struct {
	fsm              *PBFTFSM
	log              *MsgLog
	net              *p2pnet.P2PNet
	ws               *state.WorldState
	msgQueue         *MessageQueue
	replicaPeer      map[string]string
	txPool           *pool.TxPool
	isStart          bool
	view             uint64
	currentView      uint64
	stableCheckPoint uint64
	CheckPoint       uint64

	stateTimeout *time.Timer
	sync.Mutex
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
func NewPBFT(ws *state.WorldState, txPool *pool.TxPool, net *p2pnet.P2PNet) (*PBFT, error) {
	pbft := &PBFT{
		fsm:         NewFSM(),
		log:         NewMsgLog(ws.WaterHead),
		net:         net,
		ws:          ws,
		msgQueue:    NewMessageQueue(),
		replicaPeer: make(map[string]string),
		txPool:      txPool,
		isStart:     false,
	}
	// initialize from world state
	// TODO
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
		return
	}
	// add to message queue
	pbft.msgQueue.Enqueue(&msg)
}

func (pbft *PBFT) Run() {
	// register callback func
	pbft.net.RegisterCallback(p2pnet.ConsensusMsg, pbft.OnReceive)

	// run consensus
	for {
		select {
		// receive message
		case msg := <-pbft.msgQueue.Dequeue():
			if !pbft.isStart {
				continue
			}
			// run FSM handle message
			pbft.fsm.NextState(msg)
		}
	}
}

// ProposalBlockRoutine proposal new block when TxPool is full
func (pbft *PBFT) ProposalBlockRoutine() {

}
