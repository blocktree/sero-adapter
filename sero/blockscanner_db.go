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
	"encoding/hex"
	"errors"
	"path/filepath"

	"github.com/asdine/storm"
)

const (
	blockchainBucket  = "blockchain" // blockchain dataset
	NilKeyBucket = "nilkey"
)

//SaveLocalBlockHead 记录区块高度和hash到本地
func (bs *SEROBlockScanner) SaveLocalBlockHead(blockHeight uint64, blockHash string) error {

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(bs.wm.Config.dbPath, bs.wm.Config.BlockchainFile))
	if err != nil {
		return err
	}
	defer db.Close()

	db.Set(blockchainBucket, "blockHeight", &blockHeight)
	db.Set(blockchainBucket, "blockHash", &blockHash)

	return nil
}

//GetLocalBlockHead 获取本地记录的区块高度和hash
func (bs *SEROBlockScanner) GetLocalBlockHead() (uint64, string) {

	var (
		blockHeight uint64
		blockHash   string
	)

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(bs.wm.Config.dbPath, bs.wm.Config.BlockchainFile))
	if err != nil {
		return 0, ""
	}
	defer db.Close()

	db.Get(blockchainBucket, "blockHeight", &blockHeight)
	db.Get(blockchainBucket, "blockHash", &blockHash)

	return blockHeight, blockHash
}

//SaveLocalBlock 记录本地新区块
func (bs *SEROBlockScanner) SaveLocalBlock(blockHeader *BlockData) error {

	db, err := storm.Open(filepath.Join(bs.wm.Config.dbPath, bs.wm.Config.BlockchainFile))
	if err != nil {
		return err
	}
	defer db.Close()

	db.Save(blockHeader)

	return nil
}

//GetLocalBlock 获取本地区块数据
func (bs *SEROBlockScanner) GetLocalBlock(height uint64) (*BlockData, error) {

	var (
		blockHeader BlockData
	)

	db, err := storm.Open(filepath.Join(bs.wm.Config.dbPath, bs.wm.Config.BlockchainFile))
	if err != nil {
		return nil, err
	}
	defer db.Close()

	err = db.One("Height", height, &blockHeader)
	if err != nil {
		return nil, err
	}

	return &blockHeader, nil
}

//SaveUnscanRecord 保存交易记录到钱包数据库
func (bs *SEROBlockScanner) SaveUnscanRecord(record *UnscanRecord) error {

	if record == nil {
		return errors.New("the unscan record to save is nil")
	}

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(bs.wm.Config.dbPath, bs.wm.Config.BlockchainFile))
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Save(record)
}

//DeleteUnscanRecord 删除指定高度的未扫记录
func (bs *SEROBlockScanner) DeleteUnscanRecord(height uint64) error {
	//获取本地区块高度
	db, err := storm.Open(filepath.Join(bs.wm.Config.dbPath, bs.wm.Config.BlockchainFile))
	if err != nil {
		return err
	}
	defer db.Close()

	var list []*UnscanRecord
	err = db.Find("BlockHeight", height, &list)
	if err != nil {
		return err
	}

	for _, r := range list {
		db.DeleteStruct(r)
	}

	return nil
}


//SaveUnspent 记录新的未花
func (bs *SEROBlockScanner) SaveUnspent(utxo *Unspent, nilKeys [][]byte) error {

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
		nilkeyhex := hex.EncodeToString(nilkey[:])
		err = tx.Set(NilKeyBucket, nilkeyhex, utxo.Root)
	}

	return tx.Commit()
}

//DeleteUnspent 删除已使用的未花
func (bs *SEROBlockScanner) DeleteUnspent(nilKey string) error {

	db := bs.wm.unspentDB

	//先判断nilKey是否存在
	exsit, err := db.KeyExists(NilKeyBucket, nilKey)
	if err != nil {
		return err
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

	var utxo *Unspent
	err = tx.One("Root", root, &utxo)
	if err != nil {
		return err
	}

	//删除utxo记录
	err = tx.DeleteStruct(utxo)
	if err != nil {
		return err
	}

	//删除utxo与nil的关联记录
	err = tx.Delete(NilKeyBucket, nilKey)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// LockUnspent 锁定发送中utxo
func (bs *SEROBlockScanner) LockUnspent(utxo []*Unspent) error {

	db := bs.wm.unspentDB
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	for _, u := range utxo {
		u.Sending = true
		err = tx.Save(u)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}