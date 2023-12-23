package blockchain

import "bytes"

// TXoutput represents the Unspent Transaction Output (UTXO) output type
type TXoutput struct {
	Value         int    `json:"value"`         // Value holds the value of the transaction output
	ToAddress     []byte `json:"toAddress"`     // ToAddress contains the address to which the output is sent
	PublicKeyHash []byte `json:"publicKeyHash"` // PublicKeyHash is the hashed public key related to the address
}

// NewTXoutput creates a new transaction output with given value and address
func NewTXoutput(value int, address []byte) *TXoutput {
	if value < 0 {
		panic("wrong value")
	}
	output := &TXoutput{}
	output.Value = value
	output.ToAddress = address
	output.Lock(address) // Lock the output with the provided address

	return output
}

// Lock encrypts the output with the provided address to create a public key hash
func (output *TXoutput) Lock(address []byte) []byte {
	output.PublicKeyHash = Address2PublicKeyHash(address) // Generate the public key hash from the address
	return output.PublicKeyHash
}

// CanBeUnlocked verifies if the output can be unlocked with the provided address
func (output *TXoutput) CanBeUnlocked(address []byte) bool {
	return bytes.Equal(output.PublicKeyHash, Address2PublicKeyHash(address)) // Check if the address's hash matches the output's hash
}
