/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */
package sero

import (
	"fmt"
	"github.com/blocktree/openwallet/common"
	"github.com/blocktree/openwallet/crypto"
	"github.com/blocktree/openwallet/openwallet"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/tidwall/gjson"
	"math/big"
)

type BlockData struct {
	BlockNumber  uint64   `json:"number" storm:"id"`
	BlockHash    string   `json:"hash"`
	ParentHash   string   `json:"parentHash"`
	Timestamp    uint64   `json:"timestamp"`
	transactions []string `json:"transactions"`
	blockInfo    *Block
}

func NewBlock(json *gjson.Result) *BlockData {
	obj := &BlockData{}
	//解析json
	obj.BlockNumber, _ = hexutil.DecodeUint64(gjson.Get(json.Raw, "number").String())
	obj.BlockHash = gjson.Get(json.Raw, "hash").String()
	obj.ParentHash = gjson.Get(json.Raw, "parentHash").String()
	obj.Timestamp, _ = hexutil.DecodeUint64(gjson.Get(json.Raw, "timestamp").String())

	txs := make([]string, 0)
	//txDetails := make([]*Transaction, 0)
	for _, tx := range gjson.Get(json.Raw, "transactions").Array() {
		txs = append(txs, tx.String())
	}

	obj.transactions = txs

	return obj
}

//GetOutputInfoByTxID 根据txid查找blockinfo的output
func (block *BlockData) GetOutputInfoByTxID(txid string) []Out {
	output := make([]Out, 0)
	if block.blockInfo != nil {
		for _, info := range block.blockInfo.Outs {
			id := info.State.TxHash
			if id == txid {
				output = append(output, info)
			}
		}
	}
	return output
}

//BlockHeader 区块链头
func (b *BlockData) BlockHeader(symbol string) *openwallet.BlockHeader {

	obj := openwallet.BlockHeader{}
	//解析json
	obj.Hash = b.BlockHash
	obj.Previousblockhash = b.ParentHash
	obj.Height = b.BlockNumber
	obj.Time = b.Timestamp
	obj.Symbol = symbol

	return &obj
}

//UnscanRecord 扫描失败的区块及交易
type UnscanRecord struct {
	ID          string `storm:"id"` // primary key
	BlockHeight uint64
	TxID        string
	Reason      string
}

//NewUnscanRecord new UnscanRecord
func NewUnscanRecord(height uint64, txID, reason string) *UnscanRecord {
	obj := UnscanRecord{}
	obj.BlockHeight = height
	obj.TxID = txID
	obj.Reason = reason
	obj.ID = common.Bytes2Hex(crypto.SHA256([]byte(fmt.Sprintf("%d_%s", height, txID))))
	return &obj
}

//Unspent 未花记录
type Unspent struct {
	Root     string `json:"root" storm:"id"`
	Height   uint64 `json:"height" storm:"index"`
	Currency string `json:"currency"`
	Value    string `json:"value"`
	Address  string `json:"address" storm:"index"`
	TK       string `json:"tk" storm:"index"`
	Sending  bool   `json:"sending"`
}

// NewUnspent 未花
func NewUnspent(json *gjson.Result) *Unspent {
	obj := &Unspent{}
	//解析json
	obj.Root = gjson.Get(json.Raw, "root").String()
	obj.Currency = gjson.Get(json.Raw, "currency").String()
	obj.Address = gjson.Get(json.Raw, "address").String()
	obj.Value = gjson.Get(json.Raw, "value").String()
	obj.TK = gjson.Get(json.Raw, "tk").String()
	obj.Sending = gjson.Get(json.Raw, "sending").Bool()
	obj.Height = gjson.Get(json.Raw, "height").Uint()

	return obj
}

// Nil 作废码
type Nil struct {
	Nil  string `json:"nil" storm:"id"`
	Root string `json:"root"`
}

type Out struct {
	Root  string
	State RootState
}

type RootState struct {
	OS     OutState
	TxHash string
	Num    uint64
}

type OutState struct {
	Index  uint64
	Out_O  *Out_O `rlp:"nil"`
	Out_Z  *Out_Z `rlp:"nil"`
	OutCM  *string
	RootCM *string
}

type Out_O struct {
	Addr  string
	Asset Asset
	Memo  string
}

type Out_Z struct {
	AssetCM string
	OutCM   string
	RPK     string
	EInfo   string
	PKr     string
	Proof   string
}

type TDOut struct {
	Asset Asset
	Memo  string
	Nils  []string
}

type DOut struct {
	Asset Asset
	Memo  Uint512
	Nil   Uint256
}

type Block struct {
	Num  string
	Hash string
	Outs []Out
	Nils []string
}

type Asset struct {
	Tkn *Token `rlp:"nil"`
}

type Token struct {
	Currency string
	Value    string
}

type Uint256 [32]byte
type Uint512 [64]byte
type Uint128 [16]byte
type PKr [96]byte
type U256 big.Int

func (b *U256) ToInt() *big.Int {
	return (*big.Int)(b)
}
