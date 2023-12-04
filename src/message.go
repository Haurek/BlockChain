package main

import (
	"encoding/gob"
)

const (
	VersionInfo = byte(0x00)

	// protocol type
	Version    = byte(0x00)
	VersionAck = byte(0x01)

	Addr    = byte(0x02)
	GetAddr = byte(0x03)

	GetBlocks = byte(0x04)
	Inv       = byte(0x05)
	GetData   = byte(0x06)
	Blocks    = byte(0x07)

	GetHeaders = byte(0x08)
	Headers    = byte(0x09)
)

// Message is type exchange between peer
type Message struct {
	Type byte
	Data []byte
}

// WrapMsg wrap message to raw data
func WrapMsg(t byte, msg interface{}) ([]byte, error) {
	data, err := Serialize(msg)
	if err != nil {
		return nil, err
	}
	data = append([]byte{t}, data...)
	return data, nil
}

type VersionMsg struct {
	Version    byte
	BestHeight int
	TimeStamp  int64
}

type VersionAckMsg struct {
	Version    byte
	BestHeight int
	TimeStamp  int64
}

//type AddrMsg struct {
//	AddressList [][]byte
//	TimeStamp   int64
//}
//
//type GetAddrMsg struct {
//	TimeStamp int64
//}

type GetBlocksMsg struct {
	Tip       []byte
	Timestamp int64
}

type BlocksMsg struct {
}

type InvMsg struct {
}

type GetDataMsg struct {
}

//type GetHeadersMsg struct {
//}
//
//type HeadersMsg struct {
//}

func InitMsg() {
	gob.Register(VersionMsg{})
	gob.Register(VersionAckMsg{})
	//gob.Register(AddrMsg{})
	//gob.Register(GetAddrMsg{})
	gob.Register(GetBlocksMsg{})
	gob.Register(BlocksMsg{})
	gob.Register(InvMsg{})
	gob.Register(GetDataMsg{})
	//gob.Register(GetHeadersMsg{})
	//gob.Register(HeadersMsg{})
}
