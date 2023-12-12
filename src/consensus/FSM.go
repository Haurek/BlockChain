package consensus

import (
	"sync"
)

type State int32

const (
	NoneState State = iota
	PrePrepareState
	PrepareState
	CommitState
	CheckPointState
	ViewChangeState
	ViewChangeAckState
	NewViewState
)

// PBFTFSM pBFT consensus finite-state machine
// handle state of pBFT consensus
type PBFTFSM struct {
	currentState State
	lock         sync.Mutex
}

func NewFSM() *PBFTFSM {
	return &PBFTFSM{
		currentState: NoneState,
	}
}

// NextState run FSM change state when receive a PBFTMessage
func (fsm *PBFTFSM) NextState(msg *PBFTMessage) {
	// verify message
	//if msg == nil ||
	//data, msgType := msg.SplitMessage()
	switch fsm.currentState {
	case NoneState:
	case PrePrepareState:
	case PrepareState:
	case CommitState:
	case CheckPointState:
	case ViewChangeState:
	case ViewChangeAckState:
	case NewViewState:
	default:
	}
}

func (fsm *PBFTFSM) SetState(s State) {
	fsm.lock.Lock()
	defer fsm.lock.Unlock()
	fsm.currentState = s
}
