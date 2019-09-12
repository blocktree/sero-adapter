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

package openwtester

import (
	"fmt"
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openw"
	"github.com/blocktree/openwallet/openwallet"
	"github.com/blocktree/sero-adapter/sero"
)

// SERO_CreateAssetsAccount
func SERO_CreateAssetsAccount(appID, walletID, password string, account *openwallet.AssetsAccount, wm *openw.WalletManager) (*openwallet.AssetsAccount, *openwallet.Address, error) {

	var (
		wallet *openwallet.Wallet
	)

	if len(account.Alias) == 0 {
		return nil, nil, fmt.Errorf("account alias is empty")
	}

	if len(account.Symbol) == 0 {
		return nil, nil, fmt.Errorf("account symbol is empty")
	}

	if account.Required == 0 {
		account.Required = 1
	}

	assets := openw.GetAssets(account.Symbol)
	seroMgr := assets.(*sero.WalletManager)

	wrapper, err := wm.NewWalletWrapper(appID, walletID)
	if err == nil {
		wallet = wrapper.GetWallet()
	}

	if wallet == nil {
		return nil, nil, fmt.Errorf("wallet not exist")
	}

	log.Debugf("wallet[%v] is trusted", wallet.WalletID)
	//使用私钥创建子账户
	key, err := wrapper.HDKey(password)
	if err != nil {
		return nil, nil, err
	}

	newAccIndex := wallet.AccountIndex + 1

	account, err = seroMgr.CreateAccount(account.Alias, key, uint64(newAccIndex))
	if err != nil {
		return nil, nil, err
	}

	wallet.AccountIndex = newAccIndex

	account.AddressIndex = -1

	//组合拥有者
	account.OwnerKeys = []string{
		account.PublicKey,
	}

	if len(account.PublicKey) == 0 {
		return nil, nil, fmt.Errorf("account publicKey is empty")
	}

	//保存钱包到本地应用数据库
	db, err := wm.OpenDB(appID)
	if err != nil {
		return nil, nil, err
	}

	tx, err := db.Begin(true)
	if err != nil {
		return nil, nil, err
	}

	defer tx.Rollback()

	err = tx.Save(wallet)
	if err != nil {

		return nil, nil, err
	}

	err = tx.Save(account)
	if err != nil {
		return nil, nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, nil, err
	}

	log.Debug("new account create success:", account.AccountID)

	addresses, err := wm.CreateAddress(appID, walletID, account.GetAccountID(), 1)
	if err != nil {
		log.Debug("new address create failed, unexpected error:", err)
	}

	var addr *openwallet.Address
	if len(addresses) > 0 {
		addr = addresses[0]
		account.AddressIndex++
	}

	return account, addr, nil
}

