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
	"github.com/blocktree/openwallet/log"
	"github.com/mr-tron/base58"
	"github.com/shopspring/decimal"
	"testing"
)

func TestSEROBlockScanner_GetBlocksInfo(t *testing.T) {
	block, err := tw.GetBlocksInfo(1583452)
	if err != nil {
		t.Errorf("GetBlocksInfo failed, err: %v", err)
		return
	}
	log.Infof("block: %+v", block)

}

func TestSEROBlockScanner_GetBlockByNumber(t *testing.T) {
	block, err := tw.GetBlockByNumber(1579905)
	if err != nil {
		t.Errorf("GetBlockByNumber failed, err: %v", err)
		return
	}
	log.Infof("block: %+v", block)
}

func TestSEROBlockScanner_GetBlockHeight(t *testing.T) {
	height, err := tw.GetBlockHeight()
	if err != nil {
		t.Errorf("GetBlockHeight failed, err: %v", err)
		return
	}
	log.Infof("height: %d", height)
}

func TestSEROBlockScanner_GetTransactionByHash(t *testing.T) {
	txid := "0x350e16c8ae3eb81c9cec2b718153efaa99f17dd04d60d13872a0c0869c0fa573"
	tw.GetTransactionByHash(txid)
}

func TestSEROBlockScanner_GetOut(t *testing.T) {
	root := "0xfb5548851afeaa42c07056c70bd0dd55df0ca08ac146cb2d8f4d069af556d385"
	r, err := tw.GetOut(root)
	if err != nil {
		t.Errorf("GetOut failed, err: %v", err)
		return
	}
	log.Infof("result: %+v", r)
}



func TestWalletManager_DecOut(t *testing.T) {

	blockHeight := []uint64{
		1583464,
		1583452,
	}

	tk := "5tb3GBhJks3QMpPsPVabRQG4ZuhjorGZvooQhif2uRcbwJq5ZsXpCFc78hEU9Wom38MrFqQbu7SXWG7foGYt7JV6"
	tkBytes, _ := base58.Decode(tk)

	for _, height := range blockHeight {
		blockInfo, err := tw.GetBlocksInfo(height)
		if err != nil {
			t.Errorf("GetBlocksInfo failed, err: %v", err)
			return
		}

		log.Infof("=========== Height - %d ===========", height)

		douts, _ := tw.DecOut(blockInfo.Outs, tkBytes)

		for i, o := range douts {
			//log.Debugf("o: %+v", o)
			if o.Asset.Tkn != nil {
				out := blockInfo.Outs[i]
				log.Infof("[%d]txid: %s", i, hex.EncodeToString(out.State.TxHash[:]))
				if out.State.OS.Out_O != nil {
					address := base58.Encode(out.State.OS.Out_O.Addr[:])
					log.Infof("[%d]Out_O.Addr: %s", i, address)
				}
				if out.State.OS.Out_Z != nil {
					address := base58.Encode(out.State.OS.Out_Z.PKr[:])
					log.Infof("[%d]Out_Z.PKr: %s", i, address)
				}

				currency, _ := tw.LocalIdToCurrency(o.Asset.Tkn.Currency)
				amount := decimal.NewFromBigInt(o.Asset.Tkn.Value.ToInt(), 0)

				log.Infof("[%d]Currency: %s", i, currency)
				log.Infof("[%d]Value: %s", i, amount.String())
			}
			for _, nilobj := range o.Nils {
				log.Infof("[%d]NIL: %+v", i, hex.EncodeToString(nilobj[:]))
			}

		}

		log.Infof("\n")
	}


}