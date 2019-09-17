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

package openwtester

import (
	"github.com/astaxie/beego/config"
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openw"
	"github.com/blocktree/openwallet/openwallet"
	"path/filepath"
	"testing"
)

////////////////////////// 测试单个扫描器 //////////////////////////

type subscriberSingle struct {
	manager *openw.WalletManager
}

//BlockScanNotify 新区块扫描完成通知
func (sub *subscriberSingle) BlockScanNotify(header *openwallet.BlockHeader) error {
	//log.Notice("header:", header)
	return nil
}

//BlockTxExtractDataNotify 区块提取结果通知
func (sub *subscriberSingle) BlockExtractDataNotify(sourceKey string, data *openwallet.TxExtractData) error {
	log.Notice("account:", sourceKey)

	for i, input := range data.TxInputs {
		log.Std.Notice("data.TxInputs[%d]: %+v", i, input)
	}

	for i, output := range data.TxOutputs {
		log.Std.Notice("data.TxOutputs[%d]: %+v", i, output)
	}

	log.Std.Notice("data.Transaction: %+v", data.Transaction)

	balance, err := sub.manager.GetAssetsAccountBalance(testApp, "", sourceKey)
	if err != nil {
		log.Errorf("GetAssetsAccountBalance failed, err: %v", err)
	}

	log.Std.Notice("balance: %s", balance.Balance)

	if data.Transaction.Coin.IsContract {

		balance, err := sub.manager.GetAssetsAccountTokenBalance(testApp, "", sourceKey, data.Transaction.Coin.Contract)
		if err != nil {
			log.Errorf("GetAssetsAccountTokenBalance failed, err: %v", err)
		}

		log.Std.Notice("%s balance: %s", balance.Contract.Token, balance.Balance.Balance)
	}

	return nil
}

func test_scanTargetFunc(target openwallet.ScanTarget) (string, bool) {
	addr, err := tw.GetAddress(testApp, "", "", target.Address)
	if err != nil {
		return "", false
	}
	if addr == nil {
		return "", false
	}
	return addr.AccountID, true
}

func TestSubscribeAddress(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
		symbol     = "SERO"
	)

	assetsMgr, err := openw.GetAssetsAdapter(symbol)
	if err != nil {
		log.Error(symbol, "is not support")
		return
	}

	//读取配置
	absFile := filepath.Join(configFilePath, symbol+".ini")

	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return
	}
	assetsMgr.LoadAssetsConfig(c)

	assetsLogger := assetsMgr.GetAssetsLogger()
	if assetsLogger != nil {
		assetsLogger.SetLogFuncCall(true)
	}

	scanner := assetsMgr.GetBlockScanner()

	if scanner == nil {
		log.Error(symbol, "is not support block scan")
		return
	}

	scanner.SetBlockScanTargetFunc(test_scanTargetFunc)

	sub := subscriberSingle{manager: tw}
	scanner.AddObserver(&sub)

	wrapper := &walletWrapper{wm: tw}
	scanner.SetBlockScanWalletDAI(wrapper)

	//scanner.SetRescanBlockHeight(1658105)

	scanner.Run()

	<-endRunning
}

func TestSubscribeScanBlock(t *testing.T) {

	var (
		symbol     = "SERO"
		addrs      = map[string]string{
			"7EHTPNYhKNuULtwQEgFK3NuYbf3qAGNoowRHo5BHZij3mdB7WJxZ4oRJt91HbVL88pxDmBV159MsTjiwzRMD7FgqideToxcNK63VPU7LJ9ff37kJ38Yx41cSBXgdAhFRwJy": "2kfDs5Ptb1nybNnJx2TTBcRiWpmsb5wrzowQfhFjv4J8jEGSMxu7xxVSYAY32RGdefCbucDKPtiqJYjtrnksiiYL",
		}
	)

	//GetSourceKeyByAddress 获取地址对应的数据源标识
	scanTargetFunc := func(scanTarget openwallet.ScanTarget) (string, bool) {
		key, ok := addrs[scanTarget.Address]
		if !ok {
			return "", false
		}
		return key, true
	}

	assetsMgr, err := openw.GetAssetsAdapter(symbol)
	if err != nil {
		log.Error(symbol, "is not support")
		return
	}

	//读取配置
	absFile := filepath.Join(configFilePath, symbol+".ini")

	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return
	}
	assetsMgr.LoadAssetsConfig(c)

	assetsLogger := assetsMgr.GetAssetsLogger()
	if assetsLogger != nil {
		assetsLogger.SetLogFuncCall(true)
	}

	//log.Debug("already got scanner:", assetsMgr)
	scanner := assetsMgr.GetBlockScanner()
	if scanner == nil {
		log.Error(symbol, "is not support block scan")
		return
	}

	scanner.SetBlockScanTargetFunc(scanTargetFunc)
	//scanner.SetBlockScanTargetFunc(test_scanTargetFunc)

	sub := subscriberSingle{manager: tw}
	scanner.AddObserver(&sub)

	scanner.ScanBlock(1659125)
}
