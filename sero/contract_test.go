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
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openwallet"
	"testing"
)

func TestContractDecoder_GetTokenBalanceByAddress(t *testing.T) {

	address := "7EHTPNYhKNuULtwQEgFK3NuYbf3qAGNoowRHo5BHZij3mdB7WJxZ4oRJt91HbVL88pxDmBV159MsTjiwzRMD7FgqideToxcNK63VPU7LJ9ff37kJ38Yx41cSBXgdAhFRwJy"
	contract := openwallet.SmartContract{
		Address:  "AIPP",
		Symbol:   "SERO",
		Name:     "AIPP",
		Token:    "AIPP",
		Decimals: 18,
	}

	balance, err := tw.ContractDecoder.GetTokenBalanceByAddress(contract, address)
	if err != nil {
		t.Errorf("GetTokenBalanceByAddress failed, unexpected error: %v", err)
		return
	}
	for _, b := range balance {
		log.Infof("contractID: %s", openwallet.GenContractID("SERO", "AIPP"))
		log.Infof("balance: %+v", b.Balance)
	}
}