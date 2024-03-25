package blockchain

import (
	"bytes"
	"encoding/gob"

	"github.com/AayushChhabra42/Golang-Blockchain/wallet"
)

type TxInput struct {
	ID     []byte
	Out    int
	Sig    []byte
	PubKey []byte
}

type TxOutput struct {
	Value      int
	PubKeyHash []byte
}

type TxOutputs struct {
	Outputs []TxOutput
}

func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := wallet.PublicKeyHash(in.PubKey)
	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

func (out *TxOutput) Lock(address []byte) {
	pubkeyHash := wallet.Base58Decode(address)
	pubkeyHash = pubkeyHash[1 : len(pubkeyHash)-4]
	out.PubKeyHash = pubkeyHash
}

func (out *TxOutput) IsLockedWithKey(pubkeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubkeyHash) == 0
}

func NewTXOutput(value int, address string) *TxOutput {
	txo := &TxOutput{value, nil}
	txo.Lock([]byte(address))

	return txo
}

func (outs TxOutputs) Serialize() []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(outs)
	if err != nil {
		panic(err)
	}

	return buff.Bytes()
}

func DeserializeOutputs(data []byte) TxOutputs {
	var outputs TxOutputs

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	if err != nil {
		panic(err)
	}

	return outputs
}
