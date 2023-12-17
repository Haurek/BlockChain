package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"goland/x/crypto/ripemd160"
)

func Int2Bytes(value int64) []byte {
	buffer := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buffer, value)
	Bytes := make([]byte, n)
	copy(Bytes, buffer[:n])
	return Bytes
}

// CalculateChecksum calculate checksum for address generate
func CalculateChecksum(payload []byte) []byte {
	firstHash := Sha256Hash(payload)
	secondHash := Sha256Hash(firstHash[:])
	// the first 4 byte are checksum
	return secondHash[:4]
}

func Sha256Hash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func Ripemd160Hash(data []byte) []byte {
	ripemd160Hasher := ripemd160.New()
	ripemd160Hasher.Write(data)
	hash := ripemd160Hasher.Sum(nil)
	return hash[:]
}

func HandleError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

// Serialize data
func Serialize(data interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(data)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Deserialize data
func Deserialize(raw []byte, target interface{}) error {
	buffer := bytes.NewBuffer(raw)
	decoder := gob.NewDecoder(buffer)
	err := decoder.Decode(target)
	if err != nil {
		return err
	}
	return nil
}
