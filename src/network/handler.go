package p2pnet

import (
	"bufio"
	"github.com/libp2p/go-libp2p/core/network"
)

// handleStream handler peer stream
func (node *P2PNet) handleStream(stream network.Stream) {
	// create P2PStream struct
	p2pStaeam := &P2PStream{
		peerID:          stream.Conn().RemotePeer().String(),
		stream:          stream,
		MessageChan:     make(chan *Message, 0),
		closeReadStrem:  make(chan struct{}, 1),
		closeWriteStrem: make(chan struct{}, 1),
	}
	node.Lock()
	node.peerTable[p2pStaeam.peerID] = p2pStaeam
	node.Unlock()

	go node.recvData(p2pStaeam)
	go node.sendData(p2pStaeam)
}

func (node *P2PNet) recvData(stream *P2PStream) {
	rw := bufio.NewReader(stream.stream)

outLoop:
	for {
		select {
		// receive close flag
		case <-stream.closeReadStrem:
			stream.stream.Conn().Close()
			node.Lock()
			delete(node.peerTable, stream.peerID)
			node.Unlock()
			return
		default:
			node.log.Printf("receive message from: %s", stream.peerID)
			msg, err := UnpackMessage(rw)
			if err != nil {
				break outLoop
			}
			node.log.Printf("receive message type: %s", msg.Type)
			callback := node.callBacks[msg.Type]
			if callback != nil {
				node.log.Printf("Run callback for message type: %s", msg.Type)
				go callback(msg.Type, msg.Data, stream.peerID)
			} else {
				node.log.Printf("unknown Message Type")
			}
		}
	}

	stream.stream.Conn().Close()
	node.Lock()
	delete(node.peerTable, stream.peerID)
	node.Unlock()

	select {
	case stream.closeWriteStrem <- struct{}{}:
	default:
	}
}

func (node *P2PNet) sendData(stream *P2PStream) {
	rw := bufio.NewWriter(stream.stream)

outLoop:
	for {
		select {
		// receive close flag
		case <-stream.closeWriteStrem:
			stream.stream.Conn().Close()
			node.Lock()
			delete(node.peerTable, stream.peerID)
			node.Unlock()
			return
		// send message
		case msg := <-stream.MessageChan:
			node.log.Printf("send message to: %s", stream.peerID)
			msgBuf, err := PackMessage(msg)
			node.log.Printf("send message type: %s", msg.Type)
			if err != nil {
				break outLoop
			}
			_, err = rw.Write(msgBuf)
			if err != nil {
				break
			}
			rw.Flush()
		}
	}

	stream.stream.Conn().Close()
	node.Lock()
	delete(node.peerTable, stream.peerID)
	node.Unlock()

	select {
	case stream.closeReadStrem <- struct{}{}:
	default:
	}
}
