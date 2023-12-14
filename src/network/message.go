package p2pnet

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"hash/crc32"
	"io"
)

type MessageType uint8

// message received handler type
const (
	BlockMsg MessageType = iota
	TransactionMsg
	ConsensusMsg
)

// Message Generic message type, transmission in P2P network
// Type represents the received handler
// Data is the payload of the message
type Message struct {
	Type MessageType `json:"msg_type"`
	Data []byte      `json:"data"`
}

// PackMessage packs the message into a binary format.
func PackMessage(msg *Message) ([]byte, error) {
	dataBuf := bytes.NewBuffer(nil)

	msgBuf, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	lenBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(lenBuf, uint64(len(msgBuf)))
	dataBuf.Write(lenBuf)
	dataBuf.Write(msgBuf)

	data := dataBuf.Bytes()
	checksum := crc32.ChecksumIEEE(data)
	checksumBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(checksumBuf, checksum)
	dataBuf.Reset()

	dataBuf.Write(data)
	dataBuf.Write(checksumBuf)

	return dataBuf.Bytes(), nil
}

// UnpackMessage unpacks the binary data into a Message.
func UnpackMessage(rw *bufio.Reader) (*Message, error) {
	lenBuf := make([]byte, 8)
	_, err := io.ReadFull(rw, lenBuf)
	if err != nil {
		return nil, err
	}
	msgLen := binary.BigEndian.Uint64(lenBuf)
	msgBuf := make([]byte, msgLen)
	_, err = io.ReadFull(rw, msgBuf)
	if err != nil {
		return nil, err
	}

	// verify checksum
	checkBuf := make([]byte, 4)
	_, err = io.ReadFull(rw, checkBuf)
	if err != nil {
		return nil, err
	}
	readCk := binary.BigEndian.Uint32(checkBuf)

	dataBuf := bytes.NewBuffer(nil)
	dataBuf.Write(lenBuf)
	dataBuf.Write(msgBuf)
	ck := crc32.ChecksumIEEE(dataBuf.Bytes())
	if readCk != ck {
		return nil, errors.New("verify checksum failed")
	}

	var msg Message
	err = json.Unmarshal(msgBuf, &msg)
	return &msg, err
}
