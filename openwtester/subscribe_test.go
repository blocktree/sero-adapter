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
	log.Notice("header:", header)
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

	return nil
}

func TestSubscribeAddress(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
		symbol     = "SERO"
		addrs      = map[string]string{
			"QU2DzDboaMoh5bG4M4RXRwQu1BsngLFTr6LhWga5iSB27ay8U6SNnwkcX72ceCGfLCXVJwD8KwXwbtXMiXw5sNek7SXcXQKTKCvwH469oHSirRVKhQVs534BMYHSQ6Wn8VC": "5tb3GBhJks3QMpPsPVabRQG4ZuhjorGZvooQhif2uRcbwJq5ZsXpCFc78hEU9Wom38MrFqQbu7SXWG7foGYt7JV6",
			"mLeM72Gvyf2iddRNFwfYj4zTAX5CPtJ7rTDy5YJe6Vhn6x3f6dFdK1HHULYpq6xjNiwk8zCCxkYWheyBrbnWiL2dDWiWeH9AYhQ4RM3mAevLyvxbufP1Eo3UuFqLvBTmMFJ": "5tb3GBhJks3QMpPsPVabRQG4ZuhjorGZvooQhif2uRcbwJq5ZsXpCFc78hEU9Wom38MrFqQbu7SXWG7foGYt7JV6",
			"28HjcmZXRBLboNrdSEzGMaSyW8Uz4UotbA4gjHUQWaWT2RCM785eAEs5WAsfp1MTS5N85Wncwf9N4ChAjYohA33h5f3fWnm5WHrrQp2g677uAd3YyZJrPpEvMLiK84h9AxKD": "5tb3GBhJks3QMpPsPVabRQG4ZuhjorGZvooQhif2uRcbwJq5ZsXpCFc78hEU9Wom38MrFqQbu7SXWG7foGYt7JV6",
		}
	)

	//GetSourceKeyByAddress 获取地址对应的数据源标识
	scanTargetFunc := func(target openwallet.ScanTarget) (string, bool) {
		key, ok := addrs[target.Address]
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

	scanner := assetsMgr.GetBlockScanner()
	scanner.SetRescanBlockHeight(1583452)

	if scanner == nil {
		log.Error(symbol, "is not support block scan")
		return
	}

	scanner.SetBlockScanTargetFunc(scanTargetFunc)

	sub := subscriberSingle{}
	scanner.AddObserver(&sub)

	scanner.Run()

	<-endRunning
}
