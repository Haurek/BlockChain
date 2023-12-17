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
			pbft.log.Println("Receive a request")
			if pbft.isPrimary {
				// verify Txs
				var txs []*blockchain.Transaction
				err := json.Unmarshal(request.TxsBytes, &txs)
				if err != nil {
					pbft.log.Println("Unmarshal Transaction bytes error")
					return
				}
				for _, tx := range txs {
					if !blockchain.VerifyTransaction(pbft.chain, tx) {
						pbft.log.Printf("Verify Transaction %s fail", tx.ID)
						return
					}
				}
				pbft.log.Println("Verify Transactions successfully")

				// receive a request, add current seqNum
				pbft.AddCheckPoint()

				// generate digest
				reqBytes, err := json.Marshal(request)
				if err != nil {
					pbft.log.Println("Marshal request message error")
					return
				}
				digest := utils.Sha256Hash(reqBytes)

				// sign
				signature, err := mycrypto.Sign(pbft.privateKey, digest[:])
				if err != nil {
					pbft.log.Println("Sign message error")
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

				//Broadcast to replica node
				// pack PBFTMessage
				payload, err := json.Marshal(preprepare)
				if err != nil {
					pbft.log.Println("Marshal pre-prepare message error")
					return
				}
				pbftMessage := PBFTMessage{
					Type: PrePrepareMsg,
					Data: payload,
				}
				// serialize PBFTMessage
				serialized, err := json.Marshal(pbftMessage)
				if err != nil {
					pbft.log.Println("Marshal PBFTMessage error")
					return
				}
				// pack P2P network message
				p2pMessage := &p2pnet.Message{
					Type: p2pnet.ConsensusMsg,
					Data: serialized,
				}

				// change the state
				// primary broadcast pre-prepare message and change state to prepare state
				pbft.SetState(PrepareState)
				pbft.log.Println("Set state to PrepareState")

				pbft.log.Println("Broadcast pre-prepare message to replica node")
				pbft.net.Broadcast(p2pMessage)
			} else {
				// client node in this state
				pbft.AddCheckPoint()
				pbft.msgLog.AddMessage(pbft.checkPoint, RequestMsg, request)
				// change state
				pbft.log.Println("Client change state to ReplyState")
				pbft.SetState(ReplyState)
			}
		} else {
			pbft.log.Println("Unknown message type")
			return
		}
	case PrePrepareState:
		// replica node in this state wait primary node pre-prepare message
		if preprepare, ok := data.(PrePrepareMessage); ok {
			pbft.log.Println("Receive a pre-prepare message")
			// receive pre-prepare message
			// verify message
			//primaryNodePubKey := pbft.pBFTPeers[pbft.primaryID]
			pubKey := mycrypto.Bytes2PublicKey(preprepare.PubKey)
			digest := utils.Sha256Hash(preprepare.ReqMsg)

			// verify Txs
			reqByte := preprepare.ReqMsg
			var request RequestMessage
			err := json.Unmarshal(reqByte, &request)
			if err != nil {
				pbft.log.Println("Unmarshal request message bytes error")
				return
			}
			var txs []*blockchain.Transaction
			err = json.Unmarshal(request.TxsBytes, &txs)
			if err != nil {
				pbft.log.Println("Unmarshal Transaction bytes error")
				return
			}
			for _, tx := range txs {
				if !blockchain.VerifyTransaction(pbft.chain, tx) {
					pbft.log.Printf("Verify Transaction %s fail", tx.ID)
					return
				}
			}
			pbft.log.Println("Verify Transaction successfully")
			if bytes.Equal(digest, preprepare.Digest) == false {
				pbft.log.Println("Verify digest fail")
				return
			} else if pbft.view != preprepare.View {
				pbft.log.Println("Not in current view")
				return
			} else if pbft.checkPoint+1 != preprepare.SeqNum {
				pbft.log.Println("Not in current checkpoint")
				return
			} else if pbft.msgLog.GetPrePrepareLog(preprepare.SeqNum) != nil {
				pbft.log.Println("already receive this pre-prepare message")
				return
			} else if preprepare.SeqNum < pbft.checkPoint || preprepare.SeqNum > pbft.checkPoint+pbft.ws.WaterHead {
				pbft.log.Println("sequence number not in current range")
				return
			} else if !mycrypto.Verify(pubKey, preprepare.Digest, preprepare.Sign) {
				pbft.log.Println("Verify signature fail")
				return
			} else {
				// receive a pre-prepare message, add current seqNum
				pbft.AddCheckPoint()
				// add message into msgLog
				pbft.msgLog.AddMessage(pbft.checkPoint, msgType, data)

				// pack Prepare Message
				// sign
				pbft.log.Println("Sign digest")
				signature, err := mycrypto.Sign(pbft.privateKey, digest[:])
				if err != nil {
					pbft.log.Println("Sign digest fail")
					return
				}
				pubKeyBytes := mycrypto.PublicKey2Bytes(pbft.publicKey)
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

				// Broadcast to replica node
				// pack PBFTMessage
				payload, err := json.Marshal(prepare)
				if err != nil {
					pbft.log.Println("Marshal prepare message error")
					return
				}
				pbftMessage := PBFTMessage{
					Type: PrepareMsg,
					Data: payload,
				}
				// serialize PBFTMessage
				serialized, err := json.Marshal(pbftMessage)
				if err != nil {
					pbft.log.Println("Marshal PBFTMessage fail")
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

				// change the state to PrepareState
				pbft.log.Println("Change state to PrepareState")
				pbft.SetState(PrepareState)

				pbft.log.Println("Broadcast prepare message")
				pbft.net.Broadcast(p2pMessage)
			}
		} else {
			pbft.log.Println("Unknown message type")
			return
		}
	case PrepareState:
		if prepare, ok := data.(PrepareMessage); ok {
			pbft.log.Println("Receive a prepare message")
			// receive prepare message
			// verify message
			pubKey := mycrypto.Bytes2PublicKey(prepare.PubKey)
			if pbft.view != prepare.View {
				pbft.log.Println("Not in current view")
				return
			} else if pbft.checkPoint != prepare.SeqNum {
				pbft.log.Println("Not in current checkpoint")
				return
			} else if !mycrypto.Verify(pubKey, prepare.Digest, prepare.Sign) {
				pbft.log.Println("Verify digest fail")
				return
			} else if pbft.msgLog.HaveLog(prepare.SeqNum, PrepareMsg, prepare.ReplicaID) {
				pbft.log.Println("Already receive this prepare message")
				return
			} else {
				// add message to msgLog
				pbft.msgLog.AddMessage(pbft.checkPoint, msgType, data)
				prepareCount := pbft.msgLog.GetLogCount(pbft.checkPoint, PrepareMsg)
				selfPrepareMsg := pbft.msgLog.GetPrepareLog(pbft.checkPoint, pbft.selfID)
				pubKeyBytes := mycrypto.PublicKey2Bytes(pbft.publicKey)
				// check already receive prepare message
				if prepareCount == 2*pbft.maxFaultNode+1 {
					pbft.log.Println("Already receive enough prepare message, broadcast commit message")
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

					// Broadcast to replica node
					// pack PBFTMessage
					payload, err := json.Marshal(commit)
					if err != nil {
						pbft.log.Println("Marshal commit message error")
						return
					}
					pbftMessage := PBFTMessage{
						Type: CommitMsg,
						Data: payload,
					}
					// serialize PBFTMessage
					serialized, err := json.Marshal(pbftMessage)
					if err != nil {
						pbft.log.Println("Marshal PBFTMessage fail")
						return
					}
					// pack P2P network message
					p2pMessage := &p2pnet.Message{
						Type: p2pnet.ConsensusMsg,
						Data: serialized,
					}

					// change state to commit state
					pbft.log.Println("Change state to CommitState")
					pbft.SetState(CommitState)

					pbft.log.Println("Broadcast commit message")
					pbft.net.Broadcast(p2pMessage)
				}
			}
		} else {
			pbft.log.Println("Unknown message type")
			return
		}
	case CommitState:
		if commit, ok := data.(CommitMessage); ok {
			pbft.log.Println("Receive a commit message")
			// receive commit message
			// verify message
			pubKey := mycrypto.Bytes2PublicKey(commit.PubKey)
			if pbft.view != commit.View {
				pbft.log.Println("Not in current view")
				return
			} else if pbft.checkPoint != commit.SeqNum {
				pbft.log.Println("Not in current checkpoint")
				return
			} else if !mycrypto.Verify(pubKey, commit.Digest, commit.Sign) {
				pbft.log.Println("Verify digest fail")
				return
			} else if pbft.msgLog.HaveLog(commit.SeqNum, CommitMsg, commit.ReplicaID) {
				pbft.log.Println("Already receive this commit message")
				return
			} else {
				// add message to msgLog
				pbft.msgLog.AddMessage(pbft.checkPoint, msgType, data)
				commitCount := pbft.msgLog.GetLogCount(pbft.checkPoint, CommitMsg)
				//selfPrepareMsg := pbft.msgLog.GetPrepareLog(pbft.checkPoint, pbft.selfID)
				// check already receive commit message
				if commitCount == 2*pbft.maxFaultNode+1 {
					// had received enough commit message
					pbft.log.Println("Already receive enough commit message")
					// send reply message to client
					preprepare := pbft.msgLog.GetPrePrepareLog(pbft.checkPoint)
					var request RequestMessage
					err := json.Unmarshal(preprepare.ReqMsg, &request)
					if err != nil {
						pbft.log.Println("Unmarshal pre-prepare message fail")
						return
					}
					reply := ReplyMessage{
						View:      pbft.view,
						Timestamp: time.Now().Unix(),
						ClientID:  request.ClientID,
						ReplicaID: pbft.selfID,
					}

					// send reply message to client
					// pack PBFTMessage
					payload, err := json.Marshal(reply)
					if err != nil {
						pbft.log.Println("Marshal reply message error")
						return
					}
					pbftMessage := PBFTMessage{
						Type: ReplyMsg,
						Data: payload,
					}
					// serialize PBFTMessage
					serialized, err := json.Marshal(pbftMessage)
					if err != nil {
						pbft.log.Println("Marshal PBFTMessage fail")
						return
					}
					// pack P2P network message
					p2pMessage := &p2pnet.Message{
						Type: p2pnet.ConsensusMsg,
						Data: serialized,
					}

					// change state
					if pbft.isPrimary {
						// change to request state wait next request
						pbft.log.Println("Change state to RequestState")
						pbft.SetState(RequestState)
					} else {
						// replica node wait next pre-prepare message from primary node
						pbft.log.Println("Change state to PrePrepareState")
						pbft.SetState(PrePrepareState)
					}

					pbft.log.Println("Broadcast reply message to client: ", request.ClientID)
					pbft.net.BroadcastToPeer(p2pMessage, request.ClientID)
				}
			}
		} else {
			pbft.log.Println("Unknown message type")
			return
		}
	case ReplyState:
		// client in this state receive reply message
		if reply, ok := data.(ReplyMessage); ok {
			pbft.log.Println("Receive a reply message")
			pbft.msgLog.AddMessage(pbft.checkPoint, ReplyMsg, reply)
			replyCount := pbft.msgLog.GetLogCount(pbft.checkPoint, ReplyMsg)
			if replyCount == pbft.maxFaultNode+1 {
				pbft.log.Println("Receive enough reply message")
				// receive enough reply
				// broadcast new block
				request := pbft.msgLog.GetRequestLog(pbft.checkPoint)
				txBytes := request.TxsBytes
				var Txs []*blockchain.Transaction
				err := json.Unmarshal(txBytes, &Txs)
				if err != nil {
					pbft.log.Println("Unmarshal transaction bytes fail")
				}

				newBlock := blockchain.NewBlock(pbft.chain.Tip, Txs, pbft.checkPoint)
				newBlockBytes, err := json.Marshal(newBlock)
				if err != nil {
					pbft.log.Println("Marshal block fail")
				}
				blockMsg, err := pool.CreateBlockMessage(pool.NewBlockBroadcastMsg, pbft.checkPoint, newBlock.Header.Hash, newBlockBytes)
				if err != nil {
					pbft.log.Println("Create new block message fail")
				}
				serializedData, err := json.Marshal(blockMsg)
				if err != nil {
					pbft.log.Println("Marshal new block message fail")
				}
				p2pMsg := p2pnet.Message{
					Type: p2pnet.BlockMsg,
					Data: serializedData,
				}

				// change state
				pbft.log.Println("Change state to PrePrepareMessage")
				pbft.SetState(PrePrepareState)

				pbft.log.Println("Broadcast new block message")
				pbft.net.Broadcast(&p2pMsg)
			}
		} else {
			pbft.log.Println("Unknown message type")
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

func (pbft *PBFT) ResetSate() {
	pbft.fsm.lock.Lock()
	defer pbft.fsm.lock.Unlock()
	if pbft.isPrimary {
		pbft.fsm.currentState = RequestState
	} else {
		pbft.fsm.currentState = PrePrepareState
	}
}
