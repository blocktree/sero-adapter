/*
 * Copyright 2019 The openwallet Authors
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
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/blocktree/openwallet/common"
	"github.com/blocktree/openwallet/common/file"
	"github.com/blocktree/openwallet/hdkeystore"
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openwallet"
	"github.com/sero-cash/go-sero/common/hexutil"
	"path/filepath"
	"testing"
)

var (
	tw *WalletManager
)

func init() {

	tw = testNewWalletManager()
}

func testNewWalletManager() *WalletManager {
	wm := NewWalletManager()

	//读取配置
	absFile := filepath.Join("conf", "conf.ini")
	//log.Debug("absFile:", absFile)
	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		panic(err)
	}
	wm.LoadAssetsConfig(c)
	wm.WalletClient.Debug = true
	return wm
}

func testGetLocalKeyByKeyID(keyID string) *hdkeystore.HDKey {

	var (
		password = "12345678"
	)

	keydir := filepath.Join(tw.Config.DataDir, "key")

	localWallets, err := openwallet.GetWalletsByKeyDir(keydir)
	if err != nil {
		return nil
	}

	for _, w := range localWallets {

		if w.WalletID == keyID {
			keystore := hdkeystore.NewHDKeystore(
				keydir,
				hdkeystore.StandardScryptN,
				hdkeystore.StandardScryptP,
			)

			fileName := fmt.Sprintf("%s-%s.key", w.Alias, w.WalletID)

			key, _ := keystore.GetKey(
				w.WalletID,
				fileName,
				password,
			)

			return key
		}
	}

	return nil
}

func TestWalletManager_CreateWallet(t *testing.T) {

	keydir := filepath.Join(tw.Config.DataDir, "key")

	//创建目录
	file.MkdirAll(keydir)

	name := "sero"
	password := "12345678"

	_, filePath, err := tw.CreateWallet(name, password, keydir)
	if err != nil {
		t.Errorf("CreateWallet failed, error: %v", err)
		return
	}
	log.Infof("key file: %s", filePath)
}

func TestWalletManager_CreateAccount(t *testing.T) {
	key := testGetLocalKeyByKeyID("W3KUZGQqWYmPUanYfHth5A4tMuuZ2Uo7t5")
	name := "sero"
	account, err := tw.CreateAccount(name, key, 0)
	if err != nil {
		t.Errorf("CreateAccount failed, error: %v", err)
		return
	}
	log.Infof("account: %+v", account)

	/*
		{
		WalletID:W3KUZGQqWYmPUanYfHth5A4tMuuZ2Uo7t5
		Alias:sero
		AccountID:5tb3GBhJks3QMpPsPVabRQG4ZuhjorGZvooQhif2uRcbwJq5ZsXpCFc78hEU9Wom38MrFqQbu7SXWG7foGYt7JV6
		Index:0 HDPath:m/44'/88'/0'
		PublicKey:5tb3GBhJks3QMpPsPVabRQG4ZuhjorGZvooQhif2uRcc2QFvRZmxxsotxNkMjmgCyMFVpnFcgmfLfGudUDyV9ts8
		OwnerKeys:[]
		ContractAddress:
		Required:1 Symbol:SERO
		AddressIndex:-1
		Balance: IsTrust:false
		ExtParam: ModelType:0 core:<nil>}
	*/
}

func TestWalletManager_CreateAddress(t *testing.T) {

	key := testGetLocalKeyByKeyID("W3KUZGQqWYmPUanYfHth5A4tMuuZ2Uo7t5")
	name := "sero"
	account, err := tw.CreateAccount(name, key, 0)
	if err != nil {
		t.Errorf("CreateAccount failed, error: %v", err)
		return
	}

	addr, err := tw.CreateAddress(account, 0)
	if err != nil {
		t.Errorf("CreateAccount failed, error: %v", err)
		return
	}

	fmt.Printf("%s\n", addr.Address)
}


func TestWalletManager_CreateFixAddress(t *testing.T) {

	key := testGetLocalKeyByKeyID("W3KUZGQqWYmPUanYfHth5A4tMuuZ2Uo7t5")
	name := "sero"
	account, err := tw.CreateAccount(name, key, 0)
	if err != nil {
		t.Errorf("CreateAccount failed, error: %v", err)
		return
	}

	rnd, _ := hex.DecodeString("0f15135ce7333a2c11962803cd1ab905dd640954ebbebf1af3925c5347ea720a")
	log.Infof("rnd: %s", rnd)

	addr, err := tw.CreateFixAddress(account, rnd, 0)
	if err != nil {
		t.Errorf("CreateAccount failed, error: %v", err)
		return
	}

	fmt.Printf("%s\n", addr.Address)
}

func TestWalletManager_GasPrice(t *testing.T) {
	gasPrice, err := tw.GasPrice()
	if err != nil {
		t.Errorf("GasPrice failed, error: %v", err)
		return
	}
	log.Infof("gasPrice: %s", common.BigIntToDecimals(gasPrice, tw.Decimal()).String())
}

func TestWalletManager_Seed2Sk(t *testing.T) {
	seed, _ := hdkeystore.GenerateSeed(32)
	sk, err := tw.LocalSeed2Sk(hexutil.Encode(seed))
	if err != nil {
		t.Errorf("Seed2Sk failed, error: %v", err)
		return
	}
	log.Infof("sk: %s", sk)
}

func TestWalletManager_ListUnspent(t *testing.T) {
	tk := "2kfDs5Ptb1nybNnJx2TTBcRiWpmsb5wrzowQfhFjv4J8jEGSMxu7xxVSYAY32RGdefCbucDKPtiqJYjtrnksiiYL"
	currency := "AIPP"
	unspent, err := tw.ListUnspent(tk, currency, 0, -1)
	if err != nil {
		t.Errorf("ListUnspent failed, error: %v", err)
		return
	}
	for _, utxo := range unspent {
		log.Infof("utxo = %+v", utxo)
	}
}

func TestWalletManager_ListUnspentByAddress(t *testing.T) {
	address := "DuxPidNrhmk7xPMaQvB4uaD8Sssqd2CX1DFvVLUcBcfR7F7VrU6cnXbWhQMfM1tuMN4HmDGCqpyjWLPTBcqnJicwuJMx7EwNDDvdTETpy6x6a5A5Bmok1Yh7SV5aJ6TwNWb"
	currency := "SERO"
	unspent, err := tw.ListUnspentByAddress(address, currency, 0, -1)
	if err != nil {
		t.Errorf("ListUnspent failed, error: %v", err)
		return
	}
	for _, utxo := range unspent {
		log.Infof("utxo = %+v", utxo)
	}
}

func TestWalletManager_CreateRootAccount(t *testing.T) {
	key := testGetLocalKeyByKeyID("W3KUZGQqWYmPUanYfHth5A4tMuuZ2Uo7t5")
	name := "sero"
	account, err := tw.CreateRootAccount(name, key)
	if err != nil {
		t.Errorf("CreateAccount failed, error: %v", err)
		return
	}
	log.Infof("account: %+v", account)
}

