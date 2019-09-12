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

package commands

import (
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/blocktree/go-openw-cli/openwcli"
	"github.com/blocktree/go-openw-sdk/openwsdk"
	"github.com/blocktree/openwallet/console"
	"github.com/blocktree/openwallet/hdkeystore"
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/owtp"
	"github.com/blocktree/sero-adapter/sero"
)

var (
	globalCLI *openwcli.CLI
	seroMgr   *sero.WalletManager
)

func init() {
	seroMgr = sero.NewWalletManager()
}

// LoadSEROConfig 加载sero-adapter的配置
func LoadSEROConfig(path string) error {

	c, err := config.NewConfig("ini", path)
	if err != nil {
		return err
	}
	return seroMgr.LoadAssetsConfig(c)
}

//NewAccountFlow
func NewAccountFlow(cli *openwcli.CLI) error {

	//:选择钱包
	wallet, err := cli.SelectWalletStep()
	if err != nil {
		return err
	}

	//:输入钱包密码
	// 等待用户输入密码
	password, err := console.InputPassword(false, 3)
	if err != nil {
		return err
	}

	//:输入账户别名
	// 等待用户输入钱包名字
	name, err := console.InputText("Enter account's name: ", true)
	if err != nil {
		return err
	}

	//:输入币种类别
	// 等待用户输入钱包名字
	symbol, err := console.InputText("Enter account's symbol: ", true)
	if err != nil {
		return err
	}

	//创建新账户
	_, _, err = SERO_CreateAccountOnServer(cli, name, password, symbol, wallet)
	if err != nil {
		return err
	}

	return nil
}

//SERO_CreateAccountOnServer
func SERO_CreateAccountOnServer(cli *openwcli.CLI, name, password, symbol string, wallet *openwsdk.Wallet) (*openwsdk.Account, []*openwsdk.Address, error) {

	var (
		key            *hdkeystore.HDKey
		//selectedSymbol *openwsdk.Symbol
		retAccount     *openwsdk.Account
		retAddresses   []*openwsdk.Address
		err            error
		retErr         error
	)

	if len(name) == 0 {
		return nil, nil, fmt.Errorf("acount name is empty. ")
	}

	if len(password) == 0 {
		return nil, nil, fmt.Errorf("wallet password is empty. ")
	}

	//selectedSymbol, err = cli.GetSymbolInfo(symbol)
	//if err != nil {
	//	return nil, nil, err
	//}

	keystore := hdkeystore.NewHDKeystore(
		cli.GetConfig().GetKeyDir(),
		hdkeystore.StandardScryptN,
		hdkeystore.StandardScryptP,
	)

	fileName := fmt.Sprintf("%s-%s.key", wallet.Alias, wallet.WalletID)

	key, err = keystore.GetKey(
		wallet.WalletID,
		fileName,
		password,
	)
	if err != nil {
		return nil, nil, err
	}

	newAccIndex := wallet.AccountIndex + 1

	account, err := seroMgr.CreateAccount(name, key, uint64(newAccIndex))
	if err != nil {
		return nil, nil, err
	}

	newaccount := &openwsdk.Account{
		Symbol:       account.Symbol,
		AccountID:    account.AccountID,
		PublicKey:    account.PublicKey,
		Alias:        account.Alias,
		ReqSigs:      int64(account.Required),
		WalletID:     account.WalletID,
		AccountIndex: int64(account.Index),
		AddressIndex: int64(account.AddressIndex),
		HdPath:       account.HDPath,
	}

	//登记钱包的openw-server
	err = cli.APINode().CreateNormalAccount(newaccount, true,
		func(status uint64, msg string, account *openwsdk.Account, addresses []*openwsdk.Address) {
			if status == owtp.StatusSuccess {
				log.Infof("create [%s] account successfully", symbol)
				log.Infof("new accountID: %s", account.AccountID)
				if len(addresses) > 0 {
					log.Infof("new address: %s", addresses[0].Address)
				}

				retAccount = account
				retAddresses = addresses
			} else {
				log.Error("create account on server failed, unexpected error:", msg)
				retErr = fmt.Errorf(msg)
			}
		})

	if err != nil {
		return nil, nil, err
	}
	if retErr != nil {
		return nil, nil, retErr
	}

	return retAccount, retAddresses, nil
}

//SignRawTransaction 签名交易单
func SERO_SignRawTransaction(rawTx *openwsdk.RawTransaction, key *hdkeystore.HDKey) error {

	account, err := globalCLI.GetAccountByAccountID(rawTx.AccountID)
	if err != nil {
		return err
	}

	childKey, err := key.DerivedKeyWithPath(account.HdPath, seroMgr.CurveType())
	keyBytes, err := childKey.GetPrivateKeyBytes()
	if err != nil {
		return err
	}

	signature, err := seroMgr.SignTxWithSk(rawTx.RawHex, keyBytes)
	if err != nil {
		return err
	}

	rawTx.RawHex = signature.Raw

	return nil
}
