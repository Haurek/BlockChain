package main

import (
	"fmt"
	"github.com/libp2p/go-libp2p/core/network"
	"io/ioutil"
	"log"
	"time"
)

// handleStream handler peer receive stream
func handleStream(stream network.Stream, chain *Chain) {
	data, err := ioutil.ReadAll(stream)
	if err != nil {
		log.Panic(err)
	}

	t := data[0]
	payload := data[1:]
	fmt.Println("receive protocol type: ", t)
	switch t {
	case Version:
		go handleVersion(stream, payload, chain)
	case VersionAck:
		go handleVersionAck(stream, payload, chain)
	//case Addr:
	//	go handleAddr(stream, payload)
	//case GetAddr:
	//	go handleGetAddr(stream, payload)
	case GetBlocks:
		go handleGetBlocks(stream, payload, chain)
	case Inv:
		go handleInv(stream, payload, chain)
	case GetData:
		go handleGetData(stream, payload, chain)
	case Blocks:
		go handleBlocks(stream, payload, chain)
		//case GetHeaders:
		//	go handleGetHeaders(stream, payload, chain)
		//case Headers:
		//	go handleHeaders(stream, payload, chain)
	}
}

// handleVersion handle version message
func handleVersion(stream network.Stream, payload []byte, chain *Chain) {
	// receive version message
	var version VersionMsg
	err := Deserialize(payload, &version)
	HandleError(err)

	// check version
	if version.Version != VersionInfo {
		panic("unsupported version type")
	}

	// check best height of chain
	// get chain height
	chain.Lock.Lock()
	defer chain.Lock.Unlock()
	currentHeight := chain.BestHeight

	// return version ack message with best height
	versionAck := VersionAckMsg{
		Version:    VersionInfo,
		BestHeight: currentHeight,
		TimeStamp:  time.Now().Unix(),
	}
	data, err := WrapMsg(VersionAck, versionAck)
	HandleError(err)
	SendMessage(stream, data)
}

// handleVersionAck handle version ack message
func handleVersionAck(stream network.Stream, payload []byte, chain *Chain) {
	// rec
}

//// handleAddr handler addr message
//func handleAddr(stream network.Stream, payload []byte) {
//}
//
//// handleGetAddr handler getaddr message
//func handleGetAddr(stream network.Stream, payload []byte) {
//}

// handleGetBlocks handle getblocks message
func handleGetBlocks(stream network.Stream, payload []byte, chain *Chain) {
	// get peer chain tip

	// check current chain tip

	// if current chain is longer, send inv message

}

// handleInv handle inv message
func handleInv(stream network.Stream, payload []byte, chain *Chain) {
	// receive inv and get hash list

	// send getdata message ready for block synchronize

}

// handleGetData handle getdata message
func handleGetData(stream network.Stream, payload []byte, chain *Chain) {
	// receive getdata message and send blocks

}

// handleBlocks handle blocks message
func handleBlocks(stream network.Stream, payload []byte, chain *Chain) {
	// receive block from peer

	// add block to chain

}

//// handleGetHeaders handle getheaders message
//func handleGetHeaders(stream network.Stream, payload []byte, chain *Chain) {
//	// for SPV node send getheaders message to get blocks header
//
//}
//
//// handleHeaders handle headers message
//func handleHeaders(stream network.Stream, payload []byte, chain *Chain) {
//
//}
