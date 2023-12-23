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

// MessageType represents the type of message received.
type MessageType uint8

// Constants representing different message types for handling.
const (
	BlockMsg       MessageType = iota // Message type for handling blocks
	TransactionMsg                    // Message type for handling transactions
	ConsensusMsg                      // Message type for handling consensus-related messages
)

// Message represents a generic message type transmitted over the P2P network.
// Type represents the received handler.
// Data is the payload of the message.
type Message struct {
	Type MessageType `json:"msg_type"`
	Data []byte      `json:"data"`
}

// PackMessage encodes the message into a binary format for network transmission.
func PackMessage(msg *Message) ([]byte, error) {
	dataBuf := bytes.NewBuffer(nil)

	// Convert message to JSON byte array
	msgBuf, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	// Write message length and content to buffer
	lenBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(lenBuf, uint64(len(msgBuf)))
	dataBuf.Write(lenBuf)
	dataBuf.Write(msgBuf)

	// Calculate and append CRC32 checksum to ensure data integrity
	data := dataBuf.Bytes()
	checksum := crc32.ChecksumIEEE(data)
	checksumBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(checksumBuf, checksum)
	dataBuf.Reset()

	dataBuf.Write(data)
	dataBuf.Write(checksumBuf)

	return dataBuf.Bytes(), nil
}

// UnpackMessage decodes the binary data into a Message struct.
func UnpackMessage(rw *bufio.Reader) (*Message, error) {
	lenBuf := make([]byte, 8)
	_, err := io.ReadFull(rw, lenBuf)
	if err != nil {
		return nil, err
	}

	// Retrieve message length and read the message content
	msgLen := binary.BigEndian.Uint64(lenBuf)
	msgBuf := make([]byte, msgLen)
	_, err = io.ReadFull(rw, msgBuf)
	if err != nil {
		return nil, err
	}

	// Verify CRC32 checksum for data integrity
	checkBuf := make([]byte, 4)
	_, err = io.ReadFull(rw, checkBuf)
	if err != nil {
		return nil, err
	}
	readChecksum := binary.BigEndian.Uint32(checkBuf)

	dataBuf := bytes.NewBuffer(nil)
	dataBuf.Write(lenBuf)
	dataBuf.Write(msgBuf)
	calculatedChecksum := crc32.ChecksumIEEE(dataBuf.Bytes())
	if readChecksum != calculatedChecksum {
		return nil, errors.New("checksum verification failed")
	}

	// Unmarshal the message content into a Message struct
	var msg Message
	err = json.Unmarshal(msgBuf, &msg)
	return &msg, err
}
