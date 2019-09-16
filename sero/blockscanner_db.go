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
	"errors"
	"github.com/asdine/storm/q"
)

const (
	blockchainBucket  = "blockchain" // blockchain dataset
	NilKeyBucket = "nilkey"
)

//SaveLocalBlockHead 记录区块高度和hash到本地
func (bs *SEROBlockScanner) SaveLocalBlockHead(blockHeight uint64, blockHash string) error {

	//获取本地区块高度

	bs.wm.blockChainDB.Set(blockchainBucket, "blockHeight", &blockHeight)
	bs.wm.blockChainDB.Set(blockchainBucket, "blockHash", &blockHash)

	return nil
}

//GetLocalBlockHead 获取本地记录的区块高度和hash
func (bs *SEROBlockScanner) GetLocalBlockHead() (uint64, string) {

	var (
		blockHeight uint64
		blockHash   string
	)

	////获取本地区块高度

	bs.wm.blockChainDB.Get(blockchainBucket, "blockHeight", &blockHeight)
	bs.wm.blockChainDB.Get(blockchainBucket, "blockHash", &blockHash)

	return blockHeight, blockHash
}

//SaveLocalBlock 记录本地新区块
func (bs *SEROBlockScanner) SaveLocalBlock(blockHeader *BlockData) error {


	bs.wm.blockChainDB.Save(blockHeader)

	return nil
}

//GetLocalBlock 获取本地区块数据
func (bs *SEROBlockScanner) GetLocalBlock(height uint64) (*BlockData, error) {

	var (
		blockHeader BlockData
	)

	err := bs.wm.blockChainDB.One("BlockNumber", height, &blockHeader)
	if err != nil {
		return nil, err
	}

	return &blockHeader, nil
}


//获取未扫记录
func (bs *SEROBlockScanner) GetUnscanRecords() ([]*UnscanRecord, error) {

	var list []*UnscanRecord
	err := bs.wm.blockChainDB.All(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

//SaveUnscanRecord 保存交易记录到钱包数据库
func (bs *SEROBlockScanner) SaveUnscanRecord(record *UnscanRecord) error {

	if record == nil {
		return errors.New("the unscan record to save is nil")
	}

	////获取本地区块高度

	return bs.wm.blockChainDB.Save(record)
}

//DeleteUnscanRecord 删除指定高度的未扫记录
func (bs *SEROBlockScanner) DeleteUnscanRecord(height uint64) error {
	//获取本地区块高度


	var list []*UnscanRecord
	err := bs.wm.blockChainDB.Find("BlockHeight", height, &list)
	if err != nil {
		return err
	}

	for _, r := range list {
		bs.wm.blockChainDB.DeleteStruct(r)
	}

	return nil
}


//SaveUnspent 记录新的未花
func (bs *SEROBlockScanner) SaveUnspent(utxo *Unspent, nilKeys []string) error {

	db := bs.wm.unspentDB
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	err = tx.Save(utxo)
	if err != nil {
		return err
	}

	//nil与utxo关联，保存
	for _, nilkey := range nilKeys {
		err = tx.Set(NilKeyBucket, nilkey, utxo.Root)
	}

	return tx.Commit()
}

//DeleteUnspent 删除已使用的未花
func (bs *SEROBlockScanner) DeleteUnspent(nilKey string) error {

	db := bs.wm.unspentDB

	//先判断nilKey是否存在
	exsit, err := db.KeyExists(NilKeyBucket, nilKey)
	if err != nil {
		return nil
	}

	if !exsit {
		return nil
	}

	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	var root string
	err = tx.Get(NilKeyBucket, nilKey, &root)
	if err != nil {
		return err
	}

	var utxo Unspent
	err = tx.One("Root", root, &utxo)
	if err != nil {
		return err
	}

	//删除utxo记录
	err = tx.DeleteStruct(&utxo)
	if err != nil {
		return err
	}

	//bs.wm.Log.Infof("delete utxo = %s", root)

	//删除utxo与nil的关联记录
	err = tx.Delete(NilKeyBucket, nilKey)
	if err != nil {
		return err
	}

	//bs.wm.Log.Infof("delete nilKey = %s", nilKey)

	return tx.Commit()
}


//DeleteUnspent 删除已使用的未花
func (bs *SEROBlockScanner) DeleteUnspentByHeight(height uint64) error {

	db := bs.wm.unspentDB

	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	var list []*Unspent
	err = db.Find("Height", height, &list)
	if err != nil {
		return err
	}

	for _, u := range list {
		tx.DeleteStruct(u)
	}

	return tx.Commit()
}

// LockUnspent 锁定发送中utxo
func (wm *WalletManager) LockUnspent(root string) error {

	db := wm.unspentDB

	var utxo Unspent
	err := db.One("Root", root, &utxo)
	if err != nil {
		return err
	}

	utxo.Sending = true
	err = db.Save(&utxo)
	if err != nil {
		return err
	}

	return nil
}


// UnlockUnspent 解锁已超时发送的未花记录
func (wm *WalletManager) UnlockUnspent(currentHeight uint64) error {

	db := wm.unspentDB

	var (
		utxo []*Unspent
		err error
	)

	err = wm.unspentDB.Select(
		q.And(
			q.Eq("Sending", true),
		)).Find(&utxo)
	if err != nil {
		return err
	}

	for _, u := range utxo {
		confirms := currentHeight - u.Height
		//发送中的未花确认数
		if confirms > MinConfirms {
			u.Sending = false
			err = db.Save(&utxo)
			if err != nil {
				return err
			}
		}
	}
	return nil
}