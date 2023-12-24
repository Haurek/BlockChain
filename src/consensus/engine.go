package consensus

import (
	"BlockChain/src/blockchain"
	"BlockChain/src/mycrypto"
	p2pnet "BlockChain/src/network"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

type State int32

const (
	PrePrepareState State = iota
	PrepareState
	CommitState
	ViewChangeState
)

// PBFTEngine pBFT consensus engine
type PBFTEngine struct {
	currentState State // current state of pBFT node
	lock         sync.Mutex
}

// NewEngine create a new engine
func NewEngine() *PBFTEngine {
	return &PBFTEngine{
		currentState: PrePrepareState, // initialize state
	}
}

// NextState run engine change state when receive a PBFTMessage
func (pbft *PBFT) NextState(msg *PBFTMessage) {
	data, _ := msg.SplitMessage()
	switch pbft.engine.currentState {
	case PrePrepareState:
		// node in this state wait primary node prepare message
		if prepare, ok := data.(PrepareMessage); ok {
			// receive prepare message from primary node
			// set status and start consensus
			pbft.lock.Lock()
			pbft.isRunning = true
			// receive primary prepare message, start timer
			pbft.viewChangeTimer.Reset(ViewTimeout * time.Second)
			pbft.lock.Unlock()

			next, err := pbft.handlePrepare(&prepare)
			if next {
				// change state to prepare state, wait sign message
				pbft.SetState(PrepareState)
			} else if err != nil {
				pbft.log.Println(err)
			}
		} else if viewChange, ok := data.(ViewChangeMessage); ok {
			// receive view change message
			pbft.handleViewChange(&viewChange)
		}
	case PrepareState:
		if sign, ok := data.(SignMessage); ok {
			next, err := pbft.handleSign(&sign)
			if next {
				// change state to commit state
				pbft.SetState(CommitState)
			} else if err != nil {
				pbft.log.Println(err)
			}
		} else if viewChange, ok := data.(ViewChangeMessage); ok {
			// receive view change message
			pbft.handleViewChange(&viewChange)
		}
	case CommitState:
		if commit, ok := data.(CommitMessage); ok {
			next, err := pbft.handleCommit(&commit)
			if next {
				pbft.log.Println("Finish consensus")
				// finish consensus, reset State
				pbft.ResetState()
				// update status
				pbft.lock.Lock()
				// change primary node
				pbft.leaderIndex = (pbft.view + pbft.chain.BestHeight) % pbft.nodeNum
				if pbft.leaderIndex == pbft.index {
					pbft.isPrimary = true
				} else {
					pbft.isPrimary = false
				}
				// stop running
				pbft.viewChangeTimer.Stop()
				pbft.isRunning = false
				// clear log cache
				if pbft.chain.BestHeight%pbft.nodeNum == 0 {
					pbft.msgLog.ClearLog()
				}
				pbft.lock.Unlock()
			} else if err != nil {
				pbft.log.Println(err)
			}
		} else if viewChange, ok := data.(ViewChangeMessage); ok {
			// receive view change message
			pbft.handleViewChange(&viewChange)
		}
	case ViewChangeState:
		if viewChange, ok := data.(ViewChangeMessage); ok {
			// receive view change message
			next, err := pbft.handleViewChange(&viewChange)
			if next {
				// reset state
				pbft.ResetState()
				// update status
				pbft.lock.Lock()
				// change primary node
				pbft.leaderIndex = (pbft.view + pbft.chain.BestHeight) % pbft.nodeNum
				if pbft.leaderIndex == pbft.index {
					pbft.isPrimary = true
				} else {
					pbft.isPrimary = false
				}
				// stop running
				pbft.viewChangeTimer.Stop()
				pbft.isRunning = false
				// clear log cache
				pbft.msgLog.ClearLog()
				pbft.lock.Unlock()
			} else if err != nil {
				pbft.log.Println(err)
			}

		}
	default:
		pbft.log.Println("Unknown message")
	}
}

func (pbft *PBFT) handlePrepare(prepare *PrepareMessage) (bool, error) {
	pbft.log.Println("Receive a prepare message from: ", prepare.ID)
	// verify
	pubKey := mycrypto.Bytes2PublicKey(prepare.PubKey)
	if pbft.view != prepare.View {
		// check view
		return false, errors.New("not in current view")
	} else if !mycrypto.Verify(pubKey, prepare.BlockHash, prepare.Sign) {
		// verify signature fail
		return false, errors.New("verify signature fail")
	} else if pbft.msgLog.HaveLog(PrepareMsg, prepare.ID) {
		return false, errors.New("already receive this prepare message")
	} else {
		var block blockchain.Block
		err := json.Unmarshal(prepare.Block, &block)
		if err != nil {
			return false, errors.New("unmarshal block error")
		}
		if !bytes.Equal(block.Header.PrevHash, pbft.chain.Tip) {
			return false, errors.New("block previous hash not match")
		}
		if len(block.Transactions) == 0 {
			// empty Tx, raise view change
			pbft.viewChangeTimer.Stop()
			pbft.SetState(ViewChangeState)
		}
		// verify transaction
		for _, tx := range block.Transactions {
			if !blockchain.VerifyTransaction(pbft.chain, tx) {
				// transaction is invalid
				return false, errors.New("tx verify error")
			}
			// update tx pool
			pbft.txPool.RemoveTransaction(hex.EncodeToString(tx.ID))
		}
		pbft.log.Println("Verify block successfully")

		// add block to log cache
		pbft.msgLog.CacheBlock(&block)

		// add message to cache
		pbft.msgLog.AddMessage(PrepareMsg, *prepare)

		// sign
		signature, err := mycrypto.Sign(pbft.privateKey, prepare.BlockHash)
		if err != nil {
			return false, errors.New("sign message fail")
		}

		// pack self sign message
		msg := SignMessage{
			ID:        pbft.id,
			Height:    prepare.Height,
			BlockHash: prepare.BlockHash,
			View:      pbft.view,
			Sign:      signature,
			PubKey:    mycrypto.PublicKey2Bytes(pbft.publicKey),
		}

		p2pMessage, err := pbft.packBroadcastMessage(SignMsg, msg)
		if err != nil {
			return false, err
		}

		// add message to cache
		pbft.msgLog.AddMessage(SignMsg, msg)

		// broadcast prepare message
		pbft.log.Println("Broadcast sign message")
		pbft.net.Broadcast(p2pMessage)
	}
	return true, nil
}

func (pbft *PBFT) handleSign(sign *SignMessage) (bool, error) {
	pbft.log.Println("Receive a sign message from: ", sign.ID)
	// receive prepare message
	// verify message
	pubKey := mycrypto.Bytes2PublicKey(sign.PubKey)
	if pbft.view != sign.View {
		// check view
		return false, errors.New("not in current view")
	} else if !mycrypto.Verify(pubKey, sign.BlockHash, sign.Sign) {
		// verify signature fail
		return false, errors.New("verify signature fail")
	} else if pbft.msgLog.HaveLog(SignMsg, sign.ID) {
		// check already receive message from the peer
		return false, errors.New("already receive this prepare message")
	} else if !pbft.msgLog.HaveBlock(sign.BlockHash) {
		// check cache block hash
		return false, errors.New("unsigned block")
	} else {
		// add message to cache
		pbft.msgLog.AddMessage(SignMsg, *sign)
		pbft.log.Println("Verify sign message successfully, sign count: ", pbft.msgLog.Count(SignMsg))
		if pbft.msgLog.Count(SignMsg) == 2*pbft.maxFaultNode+1 {
			pbft.log.Println("Already receive enough sign message")
			// had received enough prepare message

			// get self Sign message
			selfSign := pbft.msgLog.GetSignLog(pbft.id)

			// pack commit message
			msg := CommitMessage{
				ID:        pbft.id,
				Height:    selfSign.Height,
				BlockHash: selfSign.BlockHash,
				View:      pbft.view,
				Sign:      selfSign.Sign,
				PubKey:    mycrypto.PublicKey2Bytes(pbft.publicKey),
			}

			p2pMessage, err := pbft.packBroadcastMessage(CommitMsg, msg)
			if err != nil {
				return false, err
			}

			// add message to cache
			pbft.msgLog.AddMessage(CommitMsg, msg)

			// broadcast prepare message
			pbft.log.Println("Broadcast commit message")
			pbft.net.Broadcast(p2pMessage)

			return true, nil
		}
	}
	return false, nil
}

func (pbft *PBFT) handleCommit(commit *CommitMessage) (bool, error) {
	pbft.log.Println("Receive a commit message from: ", commit.ID)
	// receive commit message
	// verify
	pubKey := mycrypto.Bytes2PublicKey(commit.PubKey)
	if pbft.view != commit.View {
		return false, errors.New("not in current view")
	} else if !mycrypto.Verify(pubKey, commit.BlockHash, commit.Sign) {
		return false, errors.New("verify digest fail")
	} else if pbft.msgLog.HaveLog(CommitMsg, commit.ID) {
		return false, errors.New("already receive this commit message")
	} else if !pbft.msgLog.HaveBlock(commit.BlockHash) {
		// check cache block hash
		return false, errors.New("unsigned block")
	} else {
		// add message to cache
		pbft.msgLog.AddMessage(CommitMsg, *commit)
		pbft.log.Println("Verify commit message successfully, commit count: ", pbft.msgLog.Count(CommitMsg))

		// check already receive commit message
		if pbft.msgLog.Count(CommitMsg) == 2*pbft.maxFaultNode+1 {
			// had received enough commit message
			pbft.log.Println("Already receive enough commit message")
			// add block to block pool
			pbft.log.Println("Add block to chain")
			block := pbft.msgLog.GetBlock(commit.BlockHash)
			pbft.blockPool.AddBlock(block)

			return true, nil
		}
	}
	return false, nil
}

func (pbft *PBFT) handleViewChange(viewChange *ViewChangeMessage) (bool, error) {
	pbft.log.Println("Receive a view change message from: ", viewChange.ID)
	// receive commit message
	// verify
	pubKey := mycrypto.Bytes2PublicKey(viewChange.PubKey)
	if pbft.chain.BestHeight > viewChange.Height {
		return false, errors.New("expired request")
	} else if (pbft.view+1)%pbft.nodeNum != viewChange.ToView {
		return false, errors.New("invalid view change")
	} else if !mycrypto.Verify(pubKey, viewChange.BlockHash, viewChange.Sign) {
		return false, errors.New("verify digest fail")
	} else if pbft.msgLog.HaveLog(ViewChangeMsg, viewChange.ID) {
		return false, errors.New("already receive this view change message")
	} else {
		// add message to cache
		pbft.msgLog.AddMessage(ViewChangeMsg, *viewChange)
		pbft.log.Println("Verify viewChange message successfully")
		pbft.log.Printf("to view: %d, count: %d", viewChange.ToView, pbft.msgLog.ViewChangeCount(viewChange.ToView))

		// not in view change state, pass
		if pbft.GetState() != ViewChangeState {
			return false, nil
		}

		// check already receive view change message
		if pbft.msgLog.ViewChangeCount(viewChange.ToView) == 2*pbft.maxFaultNode+1 {
			// had received enough view change message
			pbft.log.Println("Already receive enough view change message")
			// change view
			pbft.lock.Lock()
			pbft.view = viewChange.ToView
			pbft.lock.Unlock()
			return true, nil
		}
	}
	return false, nil
}

// SetState sets the state of the PBFT consensus engine
func (pbft *PBFT) SetState(s State) {
	pbft.engine.lock.Lock()
	defer pbft.engine.lock.Unlock()
	pbft.engine.currentState = s
}

// GetState retrieves the current state of the PBFT consensus engine.
func (pbft *PBFT) GetState() State {
	pbft.engine.lock.Lock()
	defer pbft.engine.lock.Unlock()
	return pbft.engine.currentState
}

// ResetState resets the state of the PBFT consensus engine to the PrePrepareState.
func (pbft *PBFT) ResetState() {
	pbft.engine.lock.Lock()
	defer pbft.engine.lock.Unlock()
	pbft.engine.currentState = PrePrepareState
}

// packBroadcastMessage pack a P2P network message by packaging a PBFT message with its corresponding type.
func (pbft *PBFT) packBroadcastMessage(t PBFTMsgType, msg interface{}) (*p2pnet.Message, error) {
	var payload []byte
	var err error

	// Determine the PBFT message type and marshal the message to JSON payload
	switch t {
	case PrepareMsg:
		if m, ok := msg.(PrepareMessage); ok {
			payload, err = json.Marshal(m)
			if err != nil {
				return nil, err
			}
		}
	case SignMsg:
		if m, ok := msg.(SignMessage); ok {
			payload, err = json.Marshal(m)
			if err != nil {
				return nil, err
			}
		}
	case CommitMsg:
		if m, ok := msg.(CommitMessage); ok {
			payload, err = json.Marshal(m)
			if err != nil {
				return nil, err
			}
		}
	case ViewChangeMsg:
		if m, ok := msg.(ViewChangeMessage); ok {
			payload, err = json.Marshal(m)
			if err != nil {
				return nil, err
			}
		}
	}

	// Create a PBFTMessage containing the type and data payload
	pbftMessage := PBFTMessage{
		Type: t,
		Data: payload,
	}

	// Serialize the PBFTMessage to JSON
	serialized, err := json.Marshal(pbftMessage)
	if err != nil {
		return nil, err
	}

	// Pack the PBFT message into a P2P network message
	p2pMessage := &p2pnet.Message{
		Type: p2pnet.ConsensusMsg,
		Data: serialized,
	}
	return p2pMessage, nil
}
