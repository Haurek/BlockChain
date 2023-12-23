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
// handle state of pBFT consensus
type PBFTEngine struct {
	currentState State
	lock         sync.Mutex
}

// NewEngine create a new engine
func NewEngine() *PBFTEngine {
	return &PBFTEngine{
		currentState: PrePrepareState,
	}
}

// NextState run engine change state when receive a PBFTMessage
func (pbft *PBFT) NextState(msg *PBFTMessage) {
	data, _ := msg.SplitMessage()
	switch pbft.engine.currentState {
	case PrePrepareState:
		// replica node in this state wait primary node pre-prepare message
		if prepare, ok := data.(PrepareMessage); ok {
			// receive pre-prepare message from primary node
			// start consensus
			pbft.lock.Lock()
			pbft.isRunning = true
			pbft.primaryTimer.Stop()
			pbft.lock.Unlock()

			next, err := pbft.handlePrepare(&prepare)
			if next {
				// change state to prepare state
				pbft.SetState(PrepareState)
			} else if err != nil {
				pbft.log.Println(err)
				pbft.lock.Lock()
				pbft.isRunning = false
				if !pbft.isPrimary {
					pbft.primaryTimer.Reset(PrimaryNodeTimeout * time.Second)
				}
				pbft.lock.Unlock()
			}
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
		}
	case CommitState:
		if commit, ok := data.(CommitMessage); ok {
			next, err := pbft.handleCommit(&commit)
			if next {
				// reset State
				pbft.ResetState()
				// update status
				pbft.lock.Lock()
				// primary node change
				pbft.leaderIndex = (pbft.view + pbft.chain.BestHeight) % pbft.nodeNum
				if pbft.leaderIndex == pbft.index {
					pbft.isPrimary = true
					pbft.primaryTimer.Stop()
					pbft.sealerTimer.Reset(TxPoolPollingTime * time.Second)
				} else {
					pbft.isPrimary = false
					pbft.primaryTimer.Reset(PrimaryNodeTimeout * time.Second)
				}
				pbft.isRunning = false
				pbft.lock.Unlock()
			} else if err != nil {
				pbft.log.Println(err)
			}
		}
	case ViewChangeState:
	default:
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
			// empty Tx, run view change
			// TODO
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

		// cache block
		pbft.blockPool.AddBlock(&block)

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
	} else if !pbft.blockPool.HaveBlock(sign.BlockHash) {
		// check cache block hash
		return false, errors.New("unsigned block")
	} else {
		// add message to cache
		pbft.msgLog.AddMessage(SignMsg, *sign)
		pbft.log.Println("sign count: ", pbft.msgLog.Count(SignMsg))
		if pbft.msgLog.Count(SignMsg) == 2*pbft.maxFaultNode+1 {
			pbft.log.Println("Already receive enough sign message, broadcast commit message")
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
	} else if !pbft.blockPool.HaveBlock(commit.BlockHash) {
		// check cache block hash
		return false, errors.New("unsigned block")
	} else {
		// add message to cache
		pbft.msgLog.AddMessage(CommitMsg, *commit)
		pbft.log.Println("commit count: ", pbft.msgLog.Count(CommitMsg))

		// check already receive commit message
		if pbft.msgLog.Count(CommitMsg) == 2*pbft.maxFaultNode+1 {
			// had received enough commit message
			pbft.log.Println("Already receive enough commit message")
			// add block to chain
			pbft.blockPool.Reindex()
			return true, nil
		}
	}
	return false, nil
}

// SetState sets the state of the PBFT consensus engine to the provided state.
func (pbft *PBFT) SetState(s State) {
	pbft.engine.lock.Lock()         // Lock to ensure thread safety during state modification
	defer pbft.engine.lock.Unlock() // Release the lock after function execution
	pbft.engine.currentState = s    // Update the current state to the provided state
}

// GetState retrieves the current state of the PBFT consensus engine.
func (pbft *PBFT) GetState() State {
	pbft.engine.lock.Lock()         // Lock to ensure thread safety during state retrieval
	defer pbft.engine.lock.Unlock() // Release the lock after function execution
	return pbft.engine.currentState // Return the current state
}

// ResetState resets the state of the PBFT consensus engine to the PrePrepareState.
func (pbft *PBFT) ResetState() {
	pbft.engine.lock.Lock()                    // Lock to ensure thread safety during state reset
	defer pbft.engine.lock.Unlock()            // Release the lock after function execution
	pbft.engine.currentState = PrePrepareState // Set the current state to PrePrepareState
}

// packBroadcastMessage prepares a P2P network message by packaging a PBFT message with its corresponding type.
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
