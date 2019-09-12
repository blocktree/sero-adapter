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
	"github.com/blocktree/openwallet/openwallet"
	"github.com/shopspring/decimal"
)

type ContractDecoder struct {
	*openwallet.SmartContractDecoderBase
	wm *WalletManager
}

//NewContractDecoder 智能合约解析器
func NewContractDecoder(wm *WalletManager) *ContractDecoder {
	decoder := ContractDecoder{}
	decoder.wm = wm
	return &decoder
}

func (decoder *ContractDecoder) GetTokenBalanceByAddress(contract openwallet.SmartContract, address ...string) ([]*openwallet.TokenBalance, error) {

	var tokenBalanceList []*openwallet.TokenBalance

	for _, addr := range address {
		utxo, err := decoder.wm.ListUnspentByAddress(addr, contract.Address, 0, -1)
		if err != nil {
			return nil, nil
		}

		obj := &openwallet.Balance{}
		tb := decimal.Zero
		for _, u := range utxo {

			b, _ := decimal.NewFromString(u.Value)
			tb = tb.Add(b.Shift(-int32(contract.Decimals)))

		}
		obj.Symbol = contract.Symbol
		obj.Address = addr
		obj.ConfirmBalance = tb.String()
		obj.Balance = obj.ConfirmBalance

		tokenBalance := &openwallet.TokenBalance{
			Contract: &contract,
			Balance:  obj,
		}

		tokenBalanceList = append(tokenBalanceList, tokenBalance)
	}

	return tokenBalanceList, nil
}
