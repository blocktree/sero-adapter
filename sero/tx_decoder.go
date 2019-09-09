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
		usedUTXO     = make([]*Unspent, 0)
		outputAddrs  = make(map[string]decimal.Decimal)
		balance      = decimal.Zero
		totalSend    = decimal.Zero
		receive    = decimal.Zero
		feesRate     = decimal.Zero
		accountID    = rawTx.Account.AccountID
		destinations = make([]string, 0)
		accountTotalSent = decimal.Zero
		txFrom           = make([]string, 0)
		txTo             = make([]string, 0)
		currency = ""
		coinDecimals = int32(0)
	)

	if rawTx.Coin.IsContract {
		currency = rawTx.Coin.Contract.Address
		coinDecimals = int32(rawTx.Coin.Contract.Decimals)
	} else {
		currency = rawTx.Coin.Symbol
		coinDecimals = decoder.wm.Decimal()
	}

	//查找账户的代币utxo
	unspents, err := decoder.wm.ListUnspent(accountID, currency)
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
		outputAddrs[addr] = deamount

		//计算账户的实际转账amount
		addresses, findErr := wrapper.GetAddressList(0, -1, "AccountID", accountID, "Address", addr)
		if findErr != nil || len(addresses) == 0 {
			accountTotalSent = accountTotalSent.Add(deamount)
		}

		txTo = append(txTo, fmt.Sprintf("%s:%s", addr, amount))
	}

	feesRate, _ = decimal.NewFromString(rawTx.FeeRate)
	fees, feesRate, err := decoder.wm.EstimateFee(feesRate)
	receive = totalSend

	if rawTx.Coin.IsContract {
		//查找账户的主币的utxo
		mainUnspents, err := decoder.wm.ListUnspent(accountID, rawTx.Coin.Symbol)
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
			txFrom = append(txFrom, fmt.Sprintf("%s:%s", u.Address, ua.String()))
			if balance.GreaterThanOrEqual(totalSend) {
				break
			}
		}
	}

	if balance.LessThan(totalSend) {
		return openwallet.Errorf(openwallet.ErrInsufficientBalanceOfAccount, "The %s balance: %s is not enough! ", currency, balance.String())
	}

	//UTXO如果大于设定限制，则分拆成多笔交易单发送
	if len(usedUTXO) > MaxTxInputs {
		errStr := fmt.Sprintf("The transaction is use max inputs over: %d", MaxTxInputs)
		return fmt.Errorf(errStr)
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

	txStruct, err := decoder.wm.GenTxParam(changeAddress, accountID, currency, coinDecimals, feesRate, usedUTXO, outputAddrs)

	rawTx.RawHex = txStruct.Raw

	if rawTx.Signatures == nil {
		rawTx.Signatures = make(map[string][]*openwallet.KeySignature)
	}

	//装配签名
	keySigs := make([]*openwallet.KeySignature, 0)

	addr, err := wrapper.GetAddress(changeAddress)
	if err != nil {
		return err
	}

	signature := openwallet.KeySignature{
		EccType: decoder.wm.Config.CurveType,
		Nonce:   "",
		Address: addr,
		Message: txStruct.Raw,
	}

	keySigs = append(keySigs, &signature)

	accountTotalSent = decimal.Zero.Sub(accountTotalSent)

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

	//keySignatures := rawTx.Signatures[rawTx.Account.AccountID]
	for accountID, keySignatures := range rawTx.Signatures {

		decoder.wm.Log.Debug("accountID:", accountID)

		if keySignatures != nil {
			for _, keySignature := range keySignatures {

				childKey, err := key.DerivedKeyWithPath(account.HDPath, keySignature.EccType)
				keyBytes, err := childKey.GetPrivateKeyBytes()
				if err != nil {
					return err
				}

				//decoder.wm.Log.Debug("privateKey:", hex.EncodeToString(keyBytes))

				signature, err := decoder.wm.SignTxWithSk(keySignature.Message, keyBytes)
				if err != nil {
					return err
				}

				keySignature.Signature = signature.Raw
			}
		}

		rawTx.Signatures[accountID] = keySignatures
	}

	decoder.wm.Log.Info("transaction hash sign success")

	return nil
}

//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (decoder *TransactionDecoder) VerifyRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
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
	return nil, nil
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

	return tx, nil
}