package consensus

import (
	"BlockChain/src/blockchain"
	"BlockChain/src/mycrypto"
	p2pnet "BlockChain/src/network"
	"BlockChain/src/pool"
	"BlockChain/src/utils"
	"bytes"
	"encoding/json"
	"sync"
	"time"
)

type State int32

const (
	RequestState State = iota
	PrePrepareState
	PrepareState
	CommitState
	ReplyState
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

func NewFSM(s State) *PBFTFSM {
	return &PBFTFSM{
		currentState: s,
	}
}

// NextState run FSM change state when receive a PBFTMessage
func (pbft *PBFT) NextState(msg *PBFTMessage) {
	data, msgType := msg.SplitMessage()
	switch pbft.fsm.currentState {
	case RequestState:
		// primary node in this state wait for client request
		// receive RequestMessage
		if request, ok := data.(RequestMessage); ok {
			if pbft.isPrimary {
				// verify Txs
				var txs []*blockchain.Transaction
				err := json.Unmarshal(request.TxsBytes, &txs)
				if err != nil {
					return
				}
				for _, tx := range txs {
					if !blockchain.VerifyTransaction(pbft.chain, tx) {
						return
					}
				}
				// receive a request, add current seqNum
				pbft.AddCheckPoint()
				// generate digest
				reqBytes, err := json.Marshal(request)
				if err != nil {
					return
				}
				digest := utils.Sha256Hash(reqBytes)
				// sign
				signature, err := mycrypto.Sign(pbft.privateKey, digest[:])
				if err != nil {
					return
				}
				// pack PrePrepareMessage
				pubKeyByte := mycrypto.PublicKey2Bytes(pbft.publicKey)
				preprepare := PrePrepareMessage{
					View:   pbft.view,
					SeqNum: pbft.checkPoint,
					Digest: digest,
					ReqMsg: reqBytes,
					Sign:   signature,
					PubKey: pubKeyByte,
				}
				// add pre-prepare message to msgLog
				pbft.msgLog.AddMessage(pbft.checkPoint, PrePrepareMsg, preprepare)
				// change the state
				// primary broadcast pre-prepare message and change state to prepare state
				pbft.SetState(PrepareState)

				//Broadcast to replica node
				// pack PBFTMessage
				pbftMessage := PBFTMessage{
					Type: PrePrepareMsg,
					Data: preprepare,
				}
				// serialize PBFTMessage
				serialized, err := json.Marshal(pbftMessage)
				if err != nil {
					return
				}
				// pack P2P network message
				p2pMessage := &p2pnet.Message{
					Type: p2pnet.ConsensusMsg,
					Data: serialized,
				}
				//for replicaPeerID, _ := range pbft.pBFTPeers {
				//	if replicaPeerID == pbft.selfID {
				//		continue
				//	}
				//	pbft.net.BroadcastToPeer(p2pMessage, replicaPeerID)
				//}
				pbft.net.Broadcast(p2pMessage)
			} else {
				// client node in this state
				pbft.AddCheckPoint()
				pbft.msgLog.AddMessage(pbft.checkPoint, RequestMsg, request)
				// change state
				pbft.SetState(ReplyState)
			}
		} else {
			// not a request message
			return
		}
	case PrePrepareState:
		// replica node in this state wait primary node pre-prepare message
		if preprepare, ok := data.(PrePrepareMessage); ok {
			// receive pre-prepare message
			// verify message
			//primaryNodePubKey := pbft.pBFTPeers[pbft.primaryID]
			pubKey := mycrypto.Bytes2PublicKey(preprepare.PubKey)
			digest := utils.Sha256Hash(preprepare.ReqMsg)
			// verify Txs
			reqByte := preprepare.ReqMsg
			var request RequestMessage
			err := json.Unmarshal(reqByte, &request)
			var txs []*blockchain.Transaction
			err = json.Unmarshal(request.TxsBytes, &txs)
			if err != nil {
				return
			}
			for _, tx := range txs {
				if !blockchain.VerifyTransaction(pbft.chain, tx) {
					return
				}
			}
			if bytes.Equal(digest, preprepare.Digest) == false {
				// check digest fail
				return
			} else if pbft.view != preprepare.View {
				// check view fail
				return
			} else if pbft.checkPoint+1 != preprepare.SeqNum {
				// check seqNum fail
				return
			} else if pbft.msgLog.GetPrePrepareLog(preprepare.SeqNum) != nil {
				// already receive pre-prepare message
				return
			} else if preprepare.SeqNum < pbft.checkPoint || preprepare.SeqNum > pbft.checkPoint+pbft.ws.WaterHead {
				//
				return
			} else if !mycrypto.Verify(pubKey, preprepare.Digest, preprepare.Sign) {
				// verify signature fail
				return
			} else {
				// receive a pre-prepare message, add current seqNum
				pbft.AddCheckPoint()
				// add message into msgLog
				pbft.msgLog.AddMessage(pbft.checkPoint, msgType, data)

				// pack Prepare Message
				// sign
				signature, err := mycrypto.Sign(pbft.privateKey, digest[:])
				pubKeyBytes := mycrypto.PublicKey2Bytes(pbft.publicKey)
				if err != nil {
					return
				}
				prepare := PrepareMessage{
					View:      pbft.view,
					SeqNum:    pbft.checkPoint,
					Digest:    digest,
					ReplicaID: pbft.selfID,
					Sign:      signature,
					PubKey:    pubKeyBytes,
				}
				// add self prepare message into msgLog
				pbft.msgLog.AddMessage(pbft.checkPoint, PrepareMsg, prepare)
				// change the state to PrepareState
				pbft.SetState(PrepareState)

				// Broadcast to replica node
				// pack PBFTMessage
				pbftMessage := PBFTMessage{
					Type: PrepareMsg,
					Data: prepare,
				}
				// serialize PBFTMessage
				serialized, err := json.Marshal(pbftMessage)
				if err != nil {
					return
				}
				// pack P2P network message
				p2pMessage := &p2pnet.Message{
					Type: p2pnet.ConsensusMsg,
					Data: serialized,
				}
				//for replicaPeerID, _ := range pbft.pBFTPeers {
				//	if replicaPeerID == pbft.selfID {
				//		continue
				//	}
				//	pbft.net.BroadcastToPeer(p2pMessage, replicaPeerID)
				//}
				pbft.net.Broadcast(p2pMessage)
			}
		} else {
			return
		}
	case PrepareState:
		if prepare, ok := data.(PrepareMessage); ok {
			// receive prepare message
			// verify message
			pubKey := mycrypto.Bytes2PublicKey(prepare.PubKey)
			if pbft.view != prepare.View {
				// check view fail
				return
			} else if pbft.checkPoint != prepare.SeqNum {
				// check seqNum fail
				return
			} else if !mycrypto.Verify(pubKey, prepare.Digest, prepare.Sign) {
				// verify signature fail
				return
			} else {
				if pbft.msgLog.HaveLog(prepare.SeqNum, PrepareMsg, prepare.ReplicaID) {
					// already have this message
					return
				}
				// add message to msgLog
				pbft.msgLog.AddMessage(pbft.checkPoint, msgType, data)
				prepareCount := pbft.msgLog.GetLogCount(pbft.checkPoint, PrepareMsg)
				selfPrepareMsg := pbft.msgLog.GetPrepareLog(pbft.checkPoint, pbft.selfID)
				pubKeyBytes := mycrypto.PublicKey2Bytes(pbft.publicKey)
				// check already receive prepare message
				if prepareCount == 2*pbft.maxFaultNode+1 {
					// had received enough prepare message
					// send commit message
					commit := CommitMessage{
						View:      pbft.view,
						SeqNum:    pbft.checkPoint,
						Digest:    selfPrepareMsg.Digest,
						ReplicaID: pbft.selfID,
						Sign:      selfPrepareMsg.Sign,
						PubKey:    pubKeyBytes,
					}
					// add self commit message
					pbft.msgLog.AddMessage(pbft.checkPoint, CommitMsg, commit)
					// change state to commit state
					pbft.SetState(CommitState)

					// Broadcast to replica node
					// pack PBFTMessage
					pbftMessage := PBFTMessage{
						Type: CommitMsg,
						Data: commit,
					}
					// serialize PBFTMessage
					serialized, err := json.Marshal(pbftMessage)
					if err != nil {
						return
					}
					// pack P2P network message
					p2pMessage := &p2pnet.Message{
						Type: p2pnet.ConsensusMsg,
						Data: serialized,
					}
					//for replicaPeerID, _ := range pbft.pBFTPeers {
					//	if replicaPeerID == pbft.selfID {
					//		continue
					//	}
					//	pbft.net.BroadcastToPeer(p2pMessage, replicaPeerID)
					//}
					pbft.net.Broadcast(p2pMessage)
				}
			}
		} else {
			return
		}
	case CommitState:
		if commit, ok := data.(CommitMessage); ok {
			// receive commit message
			// verify message
			pubKey := mycrypto.Bytes2PublicKey(commit.PubKey)
			if pbft.view != commit.View {
				// check view fail
				return
			} else if pbft.checkPoint != commit.SeqNum {
				// check seqNum fail
				return
			} else if !mycrypto.Verify(pubKey, commit.Digest, commit.Sign) {
				// verify signature fail
				return
			} else {
				if pbft.msgLog.HaveLog(commit.SeqNum, CommitMsg, commit.ReplicaID) {
					// already have this message
					return
				}
				// add message to msgLog
				pbft.msgLog.AddMessage(pbft.checkPoint, msgType, data)
				commitCount := pbft.msgLog.GetLogCount(pbft.checkPoint, CommitMsg)
				//selfPrepareMsg := pbft.msgLog.GetPrepareLog(pbft.checkPoint, pbft.selfID)
				// check already receive commit message
				if commitCount == 2*pbft.maxFaultNode+1 {
					// had received enough commit message

					// send reply message to client
					preprepare := pbft.msgLog.GetPrePrepareLog(pbft.checkPoint)
					var request RequestMessage
					err := json.Unmarshal(preprepare.ReqMsg, &request)
					if err != nil {
						return
					}
					reply := ReplyMessage{
						View:      pbft.view,
						Timestamp: time.Now().Unix(),
						ClientID:  request.ClientID,
						ReplicaID: pbft.selfID,
					}
					// change state
					if pbft.isPrimary {
						// change to request state wait next request
						pbft.SetState(RequestState)
					} else {
						// replica node wait next pre-prepare message from primary node
						pbft.SetState(PrePrepareState)
					}
					// send reply message to client
					// pack PBFTMessage
					pbftMessage := PBFTMessage{
						Type: ReplyMsg,
						Data: reply,
					}
					// serialize PBFTMessage
					serialized, err := json.Marshal(pbftMessage)
					if err != nil {
						return
					}
					// pack P2P network message
					p2pMessage := &p2pnet.Message{
						Type: p2pnet.ConsensusMsg,
						Data: serialized,
					}
					//for replicaPeerID, _ := range pbft.pBFTPeers {
					//	if replicaPeerID == pbft.selfID {
					//		continue
					//	}
					//	pbft.net.BroadcastToPeer(p2pMessage, replicaPeerID)
					//}
					pbft.net.BroadcastToPeer(p2pMessage, request.ClientID)
				}
			}
		} else {
			return
		}
	case ReplyState:
		// client in this state receive reply message
		if reply, ok := data.(ReplyMessage); ok {
			pbft.msgLog.AddMessage(pbft.checkPoint, ReplyMsg, reply)
			replyCount := pbft.msgLog.GetLogCount(pbft.checkPoint, ReplyMsg)
			if replyCount == pbft.maxFaultNode+1 {
				// receive enough reply
				// broadcast new block
				request := pbft.msgLog.GetRequestLog(pbft.checkPoint)
				txBytes := request.TxsBytes
				var Txs []*blockchain.Transaction
				err := json.Unmarshal(txBytes, &Txs)
				if err != nil {
					// msgLog
				}
				newBlock := blockchain.NewBlock(pbft.chain.Tip, Txs, pbft.checkPoint)
				newBlockBytes, err := utils.Serialize(newBlock)
				if err != nil {
					// msgLog
				}
				newBlockMsg := pool.NewBlockMessage{
					Height: pbft.checkPoint,
					Hash:   newBlock.Header.Hash,
					Block:  newBlockBytes,
				}
				blockMsg := pool.BlockMessage{
					Type: pool.NewBlockBroadcastMsg,
					Data: newBlockMsg,
				}
				serializedData, err := json.Marshal(blockMsg)
				if err != nil {
					// msgLog
				}
				p2pMsg := p2pnet.Message{
					Type: p2pnet.BlockMsg,
					Data: serializedData,
				}
				pbft.net.Broadcast(&p2pMsg)
				// change state
				pbft.SetState(PrePrepareState)
			}
		} else {
			return
		}
	case CheckPointState:
	case ViewChangeState:
	case ViewChangeAckState:
	case NewViewState:
	default:
	}
}

func (pbft *PBFT) SetState(s State) {
	pbft.fsm.lock.Lock()
	defer pbft.fsm.lock.Unlock()
	pbft.fsm.currentState = s
}
