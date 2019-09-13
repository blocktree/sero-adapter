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
	"github.com/asdine/storm"
	"github.com/astaxie/beego/config"
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openwallet"
	"github.com/blocktree/sero-adapter/client"
	"path/filepath"
	bolt "go.etcd.io/bbolt"
	"time"
)

//CurveType 曲线类型
func (wm *WalletManager) CurveType() uint32 {
	return wm.Config.CurveType
}

//FullName 币种全名
func (wm *WalletManager) FullName() string {
	return "aeternity"
}

//Symbol 币种标识
func (wm *WalletManager) Symbol() string {
	return wm.Config.Symbol
}

//Decimal 小数位精度
func (wm *WalletManager) Decimal() int32 {
	return 18
}

//BalanceModelType 余额模型类别
func (wm *WalletManager) BalanceModelType() openwallet.BalanceModelType {
	return openwallet.BalanceModelTypeAddress
}

//GetAddressDecode 地址解析器
func (wm *WalletManager) GetAddressDecoderV2() openwallet.AddressDecoderV2 {
	return wm.Decoder
}

//GetTransactionDecoder 交易单解析器
func (wm *WalletManager) GetTransactionDecoder() openwallet.TransactionDecoder {
	return wm.TxDecoder
}

//GetBlockScanner 获取区块链
func (wm *WalletManager) GetBlockScanner() openwallet.BlockScanner {

	return wm.Blockscanner
}

//LoadAssetsConfig 加载外部配置
func (wm *WalletManager) LoadAssetsConfig(c config.Configer) error {

	wm.Config.ServerAPI = c.String("serverAPI")
	wm.Config.FixGas, _ = c.Int64("fixGas")
	wm.WalletClient = client.NewClient(wm.Config.ServerAPI, false)
	wm.Config.DataDir = c.String("dataDir")

	//数据文件夹
	wm.Config.makeDataDir()

	//加载未花数据库
	unspentdbfile := filepath.Join(wm.Config.dbPath, wm.Config.unspentFile)
	unspentdb, err := storm.Open(
		unspentdbfile,
		storm.BoltOptions(
			0600,
			&bolt.Options{
				Timeout: 5 * time.Second,
				//ReadOnly: true,
			}),
	)
	if err != nil {
		return err
	}

	blockchaindb, err := storm.Open(filepath.Join(wm.Config.dbPath, wm.Config.BlockchainFile))
	if err != nil {
		return err
	}

	wm.unspentDB = unspentdb
	wm.blockChainDB = blockchaindb
	wm.Decoder.Client = wm.WalletClient

	return nil
}

//InitAssetsConfig 初始化默认配置
func (wm *WalletManager) InitAssetsConfig() (config.Configer, error) {
	return config.NewConfigData("ini", []byte(wm.Config.DefaultConfig))
}

//GetAssetsLogger 获取资产账户日志工具
func (wm *WalletManager) GetAssetsLogger() *log.OWLogger {
	return wm.Log
}

//GetSmartContractDecoder 获取智能合约解析器
func (wm *WalletManager) GetSmartContractDecoder() openwallet.SmartContractDecoder {
	return wm.ContractDecoder
}

