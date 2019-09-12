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
	"fmt"
	"github.com/blocktree/openwallet/openwallet"
	"github.com/shopspring/decimal"
	"strings"
	"time"
)

type TransactionDecoder struct {
	openwallet.TransactionDecoderBase
	wm *WalletManager //钱包管理者
}

//NewTransactionDecoder 交易单解析器
func NewTransactionDecoder(wm *WalletManager) *TransactionDecoder {
	decoder := TransactionDecoder{}
	decoder.wm = wm
	return &decoder
}

//CreateRawTransaction 创建交易单
func (decoder *TransactionDecoder) CreateRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	var (
		usedUTXO         = make([]*Unspent, 0)
		outputAddrs      = make([]Out_O, 0)
		balance          = decimal.Zero
		totalSend        = decimal.Zero
		receive          = decimal.Zero
		feesRate         = decimal.Zero
		accountID        = rawTx.Account.AccountID
		destinations     = make([]string, 0)
		accountTotalSent = decimal.Zero
		txFrom           = make([]string, 0)
		txTo             = make([]string, 0)
		currency         = ""
		coinDecimals     = int32(0)
	)

	if rawTx.Coin.IsContract {
		currency = rawTx.Coin.Contract.Address
		coinDecimals = int32(rawTx.Coin.Contract.Decimals)
	} else {
		currency = rawTx.Coin.Symbol
		coinDecimals = decoder.wm.Decimal()
	}

	//查找账户的代币utxo
	unspents, err := decoder.wm.ListUnspent(accountID, currency, 0, 50)
	if err != nil {
		return err
	}

	if len(unspents) == 0 {
		return openwallet.Errorf(openwallet.ErrInsufficientBalanceOfAccount, "[%s] %s balance is not enough", accountID, currency)
	}

	if len(rawTx.To) == 0 {
		return fmt.Errorf("Receiver addresses is empty!")
	}

	//计算总发送金额
	for addr, amount := range rawTx.To {
		deamount, _ := decimal.NewFromString(amount)
		totalSend = totalSend.Add(deamount)
		destinations = append(destinations, addr)

		output := Out_O{
			Asset: Asset{
				Tkn: &Token{
					Currency: currency,
					Value:    deamount.Shift(coinDecimals).String(),
				},
			},
			Addr: addr,
		}

		outputAddrs = append(outputAddrs, output)

		//计算账户的实际转账amount
		addresses, findErr := wrapper.GetAddressList(0, -1, "AccountID", accountID, "Address", addr)
		if findErr != nil || len(addresses) == 0 {
			accountTotalSent = accountTotalSent.Add(deamount)
		}

		if rawTx.Coin.IsContract {
			txTo = append(txTo, fmt.Sprintf("%s:%s", addr, deamount.Shift(coinDecimals)))
		} else {
			txTo = append(txTo, fmt.Sprintf("%s:%s", addr, amount))
		}
	}

	feesRate, _ = decimal.NewFromString(rawTx.FeeRate)
	fees, feesRate, err := decoder.wm.EstimateFee(feesRate)
	receive = totalSend

	if rawTx.Coin.IsContract {
		//查找账户的主币的utxo
		mainUnspents, err := decoder.wm.ListUnspent(accountID, rawTx.Coin.Symbol, 0, 50)
		if err != nil {
			return err
		}

		if len(mainUnspents) == 0 {
			return openwallet.Errorf(openwallet.ErrInsufficientBalanceOfAccount, "[%s] %s balance is not enough", accountID, currency)
		}

		//计算SERO手续费是否足够
		seroBalance := decimal.Zero
		for _, u := range mainUnspents {

			if u.Sending == false {
				ua, _ := decimal.NewFromString(u.Value)
				ua = ua.Shift(-decoder.wm.Decimal())
				seroBalance = seroBalance.Add(ua)
				usedUTXO = append(usedUTXO, u)
				if seroBalance.GreaterThanOrEqual(fees) {
					break
				}

				//UTXO如果大于设定限制，则分拆成多笔交易单发送
				if len(usedUTXO) > MaxTxInputs {
					errStr := fmt.Sprintf("The transaction is use max inputs over: %d", MaxTxInputs)
					return fmt.Errorf(errStr)
				}
			}
		}

		if seroBalance.LessThan(fees) {
			return openwallet.Errorf(openwallet.ErrInsufficientBalanceOfAccount, "The %s balance: %s is not enough! ", currency, balance.String())
		}
	} else {
		totalSend = totalSend.Add(fees)
		accountTotalSent = accountTotalSent.Add(fees)
	}

	//计算一个可用于支付的余额
	for _, u := range unspents {

		if u.Sending == false {
			ua, _ := decimal.NewFromString(u.Value)
			ua = ua.Shift(-coinDecimals)
			balance = balance.Add(ua)
			usedUTXO = append(usedUTXO, u)
			if rawTx.Coin.IsContract {
				txFrom = append(txFrom, fmt.Sprintf("%s:%s", u.Address, ua.Shift(coinDecimals).String()))
			} else {
				txFrom = append(txFrom, fmt.Sprintf("%s:%s", u.Address, ua.String()))
			}
			if balance.GreaterThanOrEqual(totalSend) {
				break
			}
		}
	}

	if balance.LessThan(totalSend) {
		return openwallet.Errorf(openwallet.ErrInsufficientBalanceOfAccount, "The %s balance: %s is not enough! ", currency, balance.String())
	}

	//取账户最后一个地址
	changeAddress := usedUTXO[0].Address

	changeAmount := balance.Sub(totalSend)
	rawTx.FeeRate = feesRate.StringFixed(decoder.wm.Decimal())
	rawTx.Fees = fees.StringFixed(decoder.wm.Decimal())

	decoder.wm.Log.Std.Notice("-----------------------------------------------")
	decoder.wm.Log.Std.Notice("From Account: %s", accountID)
	decoder.wm.Log.Std.Notice("To Address: %s", strings.Join(destinations, ", "))
	decoder.wm.Log.Std.Notice("Use: %v", balance.StringFixed(decoder.wm.Decimal()))
	decoder.wm.Log.Std.Notice("Fees: %v", fees.StringFixed(decoder.wm.Decimal()))
	decoder.wm.Log.Std.Notice("Receive: %v", receive.StringFixed(decoder.wm.Decimal()))
	decoder.wm.Log.Std.Notice("Change: %v", changeAmount.StringFixed(decoder.wm.Decimal()))
	decoder.wm.Log.Std.Notice("Change Address: %v", changeAddress)
	decoder.wm.Log.Std.Notice("-----------------------------------------------")

	txStruct, err := decoder.wm.GenTxParam(changeAddress, accountID, coinDecimals, feesRate, usedUTXO, outputAddrs)
	if err != nil {
		return err
	}

	rawTx.RawHex = txStruct.Raw

	if rawTx.Signatures == nil {
		rawTx.Signatures = make(map[string][]*openwallet.KeySignature)
	}

	//装配签名
	keySigs := make([]*openwallet.KeySignature, 0)

	for _, u := range usedUTXO {

		addr, err := wrapper.GetAddress(u.Address)
		if err != nil {
			return err
		}

		signature := openwallet.KeySignature{
			EccType: decoder.wm.Config.CurveType,
			Nonce:   "",
			Address: addr,
			Message: u.Root,
		}

		keySigs = append(keySigs, &signature)

	}

	accountTotalSent = decimal.Zero.Sub(accountTotalSent)

	if rawTx.Coin.IsContract {
		accountTotalSent = accountTotalSent.Shift(coinDecimals)
	}

	rawTx.Signatures[rawTx.Account.AccountID] = keySigs
	rawTx.IsBuilt = true
	rawTx.TxAmount = accountTotalSent.StringFixed(decoder.wm.Decimal())
	rawTx.TxFrom = txFrom
	rawTx.TxTo = txTo

	return nil
}

//SignRawTransaction 签名交易单
func (decoder *TransactionDecoder) SignRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	account, err := wrapper.GetAssetsAccountInfo(rawTx.Account.AccountID)
	if err != nil {
		return err
	}

	key, err := wrapper.HDKey()
	if err != nil {
		return err
	}

	childKey, err := key.DerivedKeyWithPath(account.HDPath, decoder.wm.CurveType())
	keyBytes, err := childKey.GetPrivateKeyBytes()
	if err != nil {
		return err
	}

	//decoder.wm.Log.Debug("privateKey:", hex.EncodeToString(keyBytes))

	signature, err := decoder.wm.SignTxWithSk(rawTx.RawHex, keyBytes)
	if err != nil {
		return err
	}

	rawTx.RawHex = signature.Raw

	decoder.wm.Log.Info("transaction hash sign success")

	return nil
}

//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (decoder *TransactionDecoder) VerifyRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	rawTx.IsCompleted = true
	return nil
}

//CreateSummaryRawTransaction 创建汇总交易，返回原始交易单数组
func (decoder *TransactionDecoder) CreateSummaryRawTransaction(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransaction, error) {
	var (
		rawTxWithErrArray []*openwallet.RawTransactionWithError
		rawTxArray        = make([]*openwallet.RawTransaction, 0)
		err               error
	)
	rawTxWithErrArray, err = decoder.CreateSummaryRawTransactionWithError(wrapper, sumRawTx)
	if err != nil {
		return nil, err
	}
	for _, rawTxWithErr := range rawTxWithErrArray {
		if rawTxWithErr.Error != nil {
			continue
		}
		rawTxArray = append(rawTxArray, rawTxWithErr.RawTx)
	}
	return rawTxArray, nil
}

// CreateSummaryRawTransactionWithError 创建汇总交易，返回能原始交易单数组（包含带错误的原始交易单）
func (decoder *TransactionDecoder) CreateSummaryRawTransactionWithError(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransactionWithError, error) {

	var (
		outputAddrs      = make([]Out_O, 0)
		balance          = decimal.Zero
		feesRate         = decimal.Zero
		accountID        = sumRawTx.Account.AccountID
		accountTotalSent = decimal.Zero
		txFrom           = make([]string, 0)
		txTo             = make([]string, 0)
		currency         = ""
		coinDecimals     = int32(0)
		sumAmount        = decimal.Zero
		minTransfer, _   = decimal.NewFromString(sumRawTx.MinTransfer)
		rawTxArray       = make([]*openwallet.RawTransactionWithError, 0)
	)

	if sumRawTx.Coin.IsContract {
		currency = sumRawTx.Coin.Contract.Address
		coinDecimals = int32(sumRawTx.Coin.Contract.Decimals)
	} else {
		currency = sumRawTx.Coin.Symbol
		coinDecimals = decoder.wm.Decimal()
	}

	if len(sumRawTx.SummaryAddress) == 0 {
		return nil, fmt.Errorf("summary address is empty!")
	}

	//查找账户的代币utxo
	tokenUnspents, err := decoder.wm.ListUnspent(accountID, currency, sumRawTx.AddressStartIndex, sumRawTx.AddressLimit)
	if err != nil {
		return nil, err
	}

	if len(tokenUnspents) == 0 {
		return nil, openwallet.Errorf(openwallet.ErrInsufficientBalanceOfAccount, "[%s] %s balance is not enough", accountID, currency)
	}

	//合计所有utxo
	for _, u := range tokenUnspents {

		if u.Sending == false {
			ua, _ := decimal.NewFromString(u.Value)
			ua = ua.Shift(-coinDecimals)
			balance = balance.Add(ua)

			if sumRawTx.Coin.IsContract {
				txFrom = append(txFrom, fmt.Sprintf("%s:%s", u.Address, ua.Shift(coinDecimals).String()))
			} else {
				txFrom = append(txFrom, fmt.Sprintf("%s:%s", u.Address, ua.String()))
			}
		}
	}

	feesRate, _ = decimal.NewFromString(sumRawTx.FeeRate)
	fees, feesRate, err := decoder.wm.EstimateFee(feesRate)

	if sumRawTx.Coin.IsContract {

		sumAmount = balance

		//查找账户的代币utxo
		symbolUnspents, err := decoder.wm.ListUnspent(accountID, decoder.wm.Symbol(), sumRawTx.AddressStartIndex, sumRawTx.AddressLimit)
		if err != nil {
			return nil, err
		}

		//查找足够付费的utxo
		supportUnspent, supportErr := decoder.getUTXOSatisfyAmount(symbolUnspents, fees)
		if supportErr != nil {
			return nil, supportErr
		}

		//手续费地址utxo作为输入
		tokenUnspents = append(tokenUnspents, supportUnspent)

		supportAmount, _ := decimal.NewFromString(supportUnspent.Value)
		supportAmount = supportAmount.Shift(-decoder.wm.Decimal())

		//多余的主币找零到汇总地址
		if supportAmount.GreaterThan(fees) {
			surplus := supportAmount.Sub(fees)
			output := Out_O{
				Asset: Asset{
					Tkn: &Token{
						Currency: sumRawTx.Coin.Symbol,
						Value:    surplus.Shift(decoder.wm.Decimal()).String(),
					},
				},
				Addr: sumRawTx.SummaryAddress,
			}
			outputAddrs = append(outputAddrs, output)
		}

	} else {
		sumAmount = balance.Sub(fees)
	}

	sumOutput := Out_O{
		Asset: Asset{
			Tkn: &Token{
				Currency: currency,
				Value:    sumAmount.Shift(coinDecimals).String(),
			},
		},
		Addr: sumRawTx.SummaryAddress,
	}

	outputAddrs = append(outputAddrs, sumOutput)

	//计算账户的实际转账amount
	addresses, findErr := wrapper.GetAddressList(0, -1, "AccountID", accountID, "Address", sumRawTx.SummaryAddress)
	if findErr != nil || len(addresses) == 0 {
		accountTotalSent = sumAmount
	}

	if sumRawTx.Coin.IsContract {
		txTo = append(txTo, fmt.Sprintf("%s:%s", sumRawTx.SummaryAddress, sumAmount.Shift(coinDecimals)))
	} else {
		txTo = append(txTo, fmt.Sprintf("%s:%s", sumRawTx.SummaryAddress, sumAmount.String()))
	}

	//超过最低转账额才发送
	if balance.LessThan(minTransfer) {
		return rawTxArray, nil
	}

	decoder.wm.Log.Std.Notice("-----------------------------------------------")
	decoder.wm.Log.Std.Notice("From Account: %s", accountID)
	decoder.wm.Log.Std.Notice("Summary Address: %s", sumRawTx.SummaryAddress)
	decoder.wm.Log.Std.Notice("Summary Amount: %v", sumAmount.String())
	decoder.wm.Log.Std.Notice("Fees: %v", fees.String())
	decoder.wm.Log.Std.Notice("-----------------------------------------------")

	changeAddress := tokenUnspents[0].Address

	txStruct, err := decoder.wm.GenTxParam(changeAddress, accountID, coinDecimals, feesRate, tokenUnspents, outputAddrs)
	if err != nil {
		return nil, err
	}

	//创建一笔交易单
	rawTx := &openwallet.RawTransaction{
		Coin:     sumRawTx.Coin,
		Account:  sumRawTx.Account,
		FeeRate:  sumRawTx.FeeRate,
		To:       map[string]string{sumRawTx.SummaryAddress: sumAmount.String()},
		Fees:     fees.String(),
		Required: 1,
	}

	rawTx.RawHex = txStruct.Raw

	if rawTx.Signatures == nil {
		rawTx.Signatures = make(map[string][]*openwallet.KeySignature)
	}

	//装配签名
	keySigs := make([]*openwallet.KeySignature, 0)

	for _, u := range tokenUnspents {

		addr, err := wrapper.GetAddress(u.Address)
		if err != nil {
			return nil, err
		}

		signature := openwallet.KeySignature{
			EccType: decoder.wm.Config.CurveType,
			Nonce:   "",
			Address: addr,
			Message: u.Root,
		}

		keySigs = append(keySigs, &signature)

	}

	accountTotalSent = decimal.Zero.Sub(accountTotalSent)

	if sumRawTx.Coin.IsContract {
		accountTotalSent = accountTotalSent.Shift(coinDecimals)
	}

	rawTx.Signatures[rawTx.Account.AccountID] = keySigs
	rawTx.IsBuilt = true
	rawTx.TxAmount = accountTotalSent.String()
	rawTx.TxFrom = txFrom
	rawTx.TxTo = txTo

	rawTxWithErr := &openwallet.RawTransactionWithError{
		RawTx: rawTx,
		Error: nil,
	}

	rawTxArray = append(rawTxArray, rawTxWithErr)

	return rawTxArray, nil

}

//SendRawTransaction 广播交易单
func (decoder *TransactionDecoder) SubmitRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) (*openwallet.Transaction, error) {

	if len(rawTx.RawHex) == 0 {
		return nil, fmt.Errorf("transaction hex is empty")
	}

	if !rawTx.IsCompleted {
		return nil, fmt.Errorf("transaction is not completed validation")
	}

	txid, err := decoder.wm.CommitTx(rawTx.RawHex)
	if err != nil {
		decoder.wm.Log.Warningf("[Sid: %s] submit raw hex: %s", rawTx.Sid, rawTx.RawHex)
		return nil, err
	}

	rawTx.TxID = txid
	rawTx.IsSubmit = true

	decimals := int32(0)
	fees := "0"
	if rawTx.Coin.IsContract {
		decimals = int32(rawTx.Coin.Contract.Decimals)
		fees = "0"
	} else {
		decimals = int32(decoder.wm.Decimal())
		fees = rawTx.Fees
	}

	//记录一个交易单
	tx := &openwallet.Transaction{
		From:       rawTx.TxFrom,
		To:         rawTx.TxTo,
		Amount:     rawTx.TxAmount,
		Coin:       rawTx.Coin,
		TxID:       rawTx.TxID,
		Decimal:    decimals,
		AccountID:  rawTx.Account.AccountID,
		Fees:       fees,
		SubmitTime: time.Now().Unix(),
	}

	tx.WxID = openwallet.GenTransactionWxID(tx)

	//广播成功后标记utxo已发送
	keySignatures := rawTx.Signatures[rawTx.Account.AccountID]
	if keySignatures != nil {
		for _, keySignature := range keySignatures {
			lockErr := decoder.wm.LockUnspent(keySignature.Message)
			if lockErr != nil {
				decoder.wm.Log.Errorf("LockUnspent failed, error: %v", lockErr)
			}
		}
	}

	return tx, nil
}

// getAssetsAccountUnspentSatisfyAmount
func (decoder *TransactionDecoder) getUTXOSatisfyAmount(unspents []*Unspent, amount decimal.Decimal) (*Unspent, *openwallet.Error) {

	var utxo *Unspent

	if unspents != nil {
		for _, u := range unspents {
			if !u.Sending {
				ua, _ := decimal.NewFromString(u.Value)
				if ua.GreaterThanOrEqual(amount) {
					utxo = u
					break
				}
			}
		}
	}

	if utxo == nil {
		return nil, openwallet.Errorf(openwallet.ErrInsufficientBalanceOfAccount, "account have not available utxo")
	}

	return utxo, nil
}
