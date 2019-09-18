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
	"github.com/blocktree/openwallet/openw"
	"testing"

	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openwallet"
)

func testGetAssetsAccountBalance(tm *openw.WalletManager, walletID, accountID string) {
	balance, err := tm.GetAssetsAccountBalance(testApp, walletID, accountID)
	if err != nil {
		log.Error("GetAssetsAccountBalance failed, unexpected error:", err)
		return
	}
	log.Info("balance:", balance)
}

func testGetAssetsAccountTokenBalance(tm *openw.WalletManager, walletID, accountID string, contract openwallet.SmartContract) {
	balance, err := tm.GetAssetsAccountTokenBalance(testApp, walletID, accountID, contract)
	if err != nil {
		log.Error("GetAssetsAccountTokenBalance failed, unexpected error:", err)
		return
	}
	log.Info("token balance:", balance.Balance)
}

func testCreateTransactionStep(tm *openw.WalletManager, walletID, accountID, to, amount, feeRate string, contract *openwallet.SmartContract) (*openwallet.RawTransaction, error) {

	//err := tm.RefreshAssetsAccountBalance(testApp, accountID)
	//if err != nil {
	//	log.Error("RefreshAssetsAccountBalance failed, unexpected error:", err)
	//	return nil, err
	//}

	rawTx, err := tm.CreateTransaction(testApp, walletID, accountID, amount, to, feeRate, "", contract)

	if err != nil {
		log.Error("CreateTransaction failed, unexpected error:", err)
		return nil, err
	}

	return rawTx, nil
}

func testCreateSummaryTransactionStep(
	tm *openw.WalletManager,
	walletID, accountID, summaryAddress, minTransfer, retainedBalance, feeRate string,
	start, limit int,
	contract *openwallet.SmartContract,
	feeSupportAccount *openwallet.FeesSupportAccount) ([]*openwallet.RawTransactionWithError, error) {

	rawTxArray, err := tm.CreateSummaryRawTransactionWithError(testApp, walletID, accountID, summaryAddress, minTransfer,
		retainedBalance, feeRate, start, limit, contract, feeSupportAccount)

	if err != nil {
		log.Error("CreateSummaryTransaction failed, unexpected error:", err)
		return nil, err
	}

	return rawTxArray, nil
}

func testSignTransactionStep(tm *openw.WalletManager, rawTx *openwallet.RawTransaction) (*openwallet.RawTransaction, error) {

	_, err := tm.SignTransaction(testApp, rawTx.Account.WalletID, rawTx.Account.AccountID, "12345678", rawTx)
	if err != nil {
		log.Error("SignTransaction failed, unexpected error:", err)
		return nil, err
	}

	//log.Infof("rawTx: %+v", rawTx)
	return rawTx, nil
}

func testVerifyTransactionStep(tm *openw.WalletManager, rawTx *openwallet.RawTransaction) (*openwallet.RawTransaction, error) {

	//log.Info("rawTx.Signatures:", rawTx.Signatures)

	_, err := tm.VerifyTransaction(testApp, rawTx.Account.WalletID, rawTx.Account.AccountID, rawTx)
	if err != nil {
		log.Error("VerifyTransaction failed, unexpected error:", err)
		return nil, err
	}

	//log.Infof("rawTx: %+v", rawTx)
	return rawTx, nil
}

func testSubmitTransactionStep(tm *openw.WalletManager, rawTx *openwallet.RawTransaction) (*openwallet.RawTransaction, error) {

	tx, err := tm.SubmitTransaction(testApp, rawTx.Account.WalletID, rawTx.Account.AccountID, rawTx)
	if err != nil {
		log.Error("SubmitTransaction failed, unexpected error:", err)
		return nil, err
	}

	log.Std.Info("tx: %+v", tx)
	log.Info("wxID:", tx.WxID)
	log.Info("txID:", rawTx.TxID)

	return rawTx, nil
}

func TestTransfer(t *testing.T) {

	targets := []string{
		//"wUysiDBai1jZ9Tab93DH3uRy6TMFH6DpBA6aQQZSXCAu5zNumRdx5QujBj3uoz39roP11oyEsQXCawG3Gi43x7dP8mJeAC5hG1TdVSbMmuhuc3DmvTbLKyNjU1w8TgRnyWp",
		//"wkxJT5KkTLw6fbS2VCbw8CcZgUkgUtgDDPkCxjFPkaCSHCk8ADDij7NRwSE6agXCSdSHUHUh5L2gX3wYb8PTLo1KPpEXv1Fz2V6nB6Hs4ia22SZEA1GjPCfYV8nZnkTJ4HD",
		//"wpdAANC4M1wxvzdwR6wBD8tDDc3E4utZcNF57t54W6tKuztjkiNGHZbxpuMpMUWWexdarY9i5RzPFvCzY5NEPmz5L9jb7ch68Xy2H3zrAsggVGNFcApS82GXpVVdEa8YiHL",
		//"wzfxZiuTMgL1ScM5GvVzLjC7L9ocAAsfFdvsEUX68TazmPQNkryxuF5ULpqyt3rzMT2Rhe3vet2aSKcLLcth48vf3nZis7KMX6WicxzWoBS8fSXs7fWSFPofwL5q58DCBZb",
		//"xDgAwuEGgRY3TJ41kd6rqvhmvzP1okRrndiftafScP66m6zuNDRy8ZouFniDaNqs4YssUthD4YHzdfMQ48CvRsXHTsqZnuByoeYL336Prg5z1u2Cb8CVfwkQSY5nLPJm8Ro",
		//"xbGdoyR2HWWcBqVFD8XkGJxYwMB441Dr1i38vBAqxTbFAzRQryx76Bg3nVQPHrMNUxwdhFhwJwkhAbhe7x25pxq6hiHod63zT9KwMNQ1fuEqEQbe5AZv5UezH5SNM7PAAk3",
		//"xeaCps7qd9Gzidt2zDb1Y2gg3uBhvFWbWuP265E22mcMxLsoYJXfJrMfRjrv2TerAjFtqtsnNWuaguwhSzAYm9DCm6tLL6aYx6Wo5mrzB3vUQSYU9F4cK2MmnFFiY5Sqe14",
		//"xfiETMrZGzYtBJnRFt1PkSaf34ddEfsQ8YjLzpch5y96U5EEgTz168GfAhUDNMAvyXCq2hJXEBJkoCoj3cpUHXQpBGkGVc2h6LSmDNCVbQupHzmjGrQ3o2W42YVmNwiU898",
		//"yhdrgq6LB5oJZW4pCPKgwdySVuVqE8n8eKcreNaCDE2ELZb76EZxg5GwFP1wvgpPTd5HW8Y6i2TofrKsmBGRCoGWKuUempkFvZv1d4hhZ48trkar5qDJsxykFwKk2T7dn4R",
		//"zMmrYEfnPwMXjA9ufXu7jGGdXgNezu1vVK4gtfX5Ve5ssKeZ426m9RtwSzvU8thMB3ckzJtTrVErgeBPZZsvucnjrLej3YbeEbbKzhH3XDHW9ohgtSwkiCXrLdKeWu7zkj9",
		//"zPyHaAzYdgDYwY6wiX6rUA6QbKWTDfAi4ZoNLWhzpnwhnXniiEjdrEs8FRbpCCPEeKnmy2meZcuQsMwB6p5wX6z5CEtanBFw6pgrmpN9WyeFg2soe8wZRgGjtW8Eg8KPS9L",
		//"zYyTvchFSiTnJJ7vXhudH1xzViUtUERxp7RYAQrGXLLkvi5YbJdFVfaVwtBCzM4EPm73fNKShL2EjFHKAiiAuXbCLrr7dq485ToqfnXk3KJdB8DPZbjcrCuCPNxS4rhp2Ym",
		//"zhhmsjsKwSHkq8ZDyiKs4W2wiYZ11KkBUUnR9Yuizevv4hFxBjcg7C7uTZ5fGypuB5574Qm1Mp1SSWEzGBFtBPvnSN3w8TLtvGYML5MBhDcYisuxpo1kaQ92xavcRTYohYi",
		//"zxPiJnvqXCYgkSYfjPPg2y8EBbSTVE41We2Y3LxhaMMqzsr6i9vE8hS8kQAXw9DML3nMiwxqE9StWPVLC9ibQCbPfxWzBuM6VfvEtnLPKiDe9Kgo7Ap9ARKzUM1LKChcDM7",
		//"zxtHrpvVY1nAMjBcKGV6fsjaLUn7thruPWHMr2ckRhW7ZUphgN9RbaUpnr5pu4t3xJEAEqhUcaVSztpFZS9f2Cr6nxr9dBvVUgRGA9czvMBWCmiWPYb8z1XSiruJQgWC5XL",

		"7EHTPNYhKNuULtwQEgFK3NuYbf3qAGNoowRHo5BHZij3mdB7WJxZ4oRJt91HbVL88pxDmBV159MsTjiwzRMD7FgqideToxcNK63VPU7LJ9ff37kJ38Yx41cSBXgdAhFRwJy",
	}

	walletID := "W3TuDqe8VShgyPcg2dw4FRrNQbmxxiGPTJ"
	accountID := "4gbLYX9shEoABGKaZrbTmAfHRXPKkVK6wudFEp7miNFZ7F9ZCL6t38Nr6tSr8GDS11tNZn7iwghsbt2qs6P1bkje"

	testGetAssetsAccountBalance(tw, walletID, accountID)

	for _, to := range targets {

		rawTx, err := testCreateTransactionStep(tw, walletID, accountID, to, "0.012345", "", nil)
		if err != nil {
			return
		}

		//log.Std.Info("rawTx: %+v", rawTx)

		_, err = testSignTransactionStep(tw, rawTx)
		if err != nil {
			return
		}

		_, err = testVerifyTransactionStep(tw, rawTx)
		if err != nil {
			return
		}

		_, err = testSubmitTransactionStep(tw, rawTx)
		if err != nil {
			return
		}

	}

}

func TestTransfer_Token(t *testing.T) {

	targets := []string{
		//"wUysiDBai1jZ9Tab93DH3uRy6TMFH6DpBA6aQQZSXCAu5zNumRdx5QujBj3uoz39roP11oyEsQXCawG3Gi43x7dP8mJeAC5hG1TdVSbMmuhuc3DmvTbLKyNjU1w8TgRnyWp",
		//"wkxJT5KkTLw6fbS2VCbw8CcZgUkgUtgDDPkCxjFPkaCSHCk8ADDij7NRwSE6agXCSdSHUHUh5L2gX3wYb8PTLo1KPpEXv1Fz2V6nB6Hs4ia22SZEA1GjPCfYV8nZnkTJ4HD",
		//"wpdAANC4M1wxvzdwR6wBD8tDDc3E4utZcNF57t54W6tKuztjkiNGHZbxpuMpMUWWexdarY9i5RzPFvCzY5NEPmz5L9jb7ch68Xy2H3zrAsggVGNFcApS82GXpVVdEa8YiHL",
		//"wzfxZiuTMgL1ScM5GvVzLjC7L9ocAAsfFdvsEUX68TazmPQNkryxuF5ULpqyt3rzMT2Rhe3vet2aSKcLLcth48vf3nZis7KMX6WicxzWoBS8fSXs7fWSFPofwL5q58DCBZb",
		//"xDgAwuEGgRY3TJ41kd6rqvhmvzP1okRrndiftafScP66m6zuNDRy8ZouFniDaNqs4YssUthD4YHzdfMQ48CvRsXHTsqZnuByoeYL336Prg5z1u2Cb8CVfwkQSY5nLPJm8Ro",
		//"xbGdoyR2HWWcBqVFD8XkGJxYwMB441Dr1i38vBAqxTbFAzRQryx76Bg3nVQPHrMNUxwdhFhwJwkhAbhe7x25pxq6hiHod63zT9KwMNQ1fuEqEQbe5AZv5UezH5SNM7PAAk3",
		//"xeaCps7qd9Gzidt2zDb1Y2gg3uBhvFWbWuP265E22mcMxLsoYJXfJrMfRjrv2TerAjFtqtsnNWuaguwhSzAYm9DCm6tLL6aYx6Wo5mrzB3vUQSYU9F4cK2MmnFFiY5Sqe14",
		//"xfiETMrZGzYtBJnRFt1PkSaf34ddEfsQ8YjLzpch5y96U5EEgTz168GfAhUDNMAvyXCq2hJXEBJkoCoj3cpUHXQpBGkGVc2h6LSmDNCVbQupHzmjGrQ3o2W42YVmNwiU898",
		//"yhdrgq6LB5oJZW4pCPKgwdySVuVqE8n8eKcreNaCDE2ELZb76EZxg5GwFP1wvgpPTd5HW8Y6i2TofrKsmBGRCoGWKuUempkFvZv1d4hhZ48trkar5qDJsxykFwKk2T7dn4R",
		//"zMmrYEfnPwMXjA9ufXu7jGGdXgNezu1vVK4gtfX5Ve5ssKeZ426m9RtwSzvU8thMB3ckzJtTrVErgeBPZZsvucnjrLej3YbeEbbKzhH3XDHW9ohgtSwkiCXrLdKeWu7zkj9",
		//"zPyHaAzYdgDYwY6wiX6rUA6QbKWTDfAi4ZoNLWhzpnwhnXniiEjdrEs8FRbpCCPEeKnmy2meZcuQsMwB6p5wX6z5CEtanBFw6pgrmpN9WyeFg2soe8wZRgGjtW8Eg8KPS9L",
		//"zYyTvchFSiTnJJ7vXhudH1xzViUtUERxp7RYAQrGXLLkvi5YbJdFVfaVwtBCzM4EPm73fNKShL2EjFHKAiiAuXbCLrr7dq485ToqfnXk3KJdB8DPZbjcrCuCPNxS4rhp2Ym",
		//"zhhmsjsKwSHkq8ZDyiKs4W2wiYZ11KkBUUnR9Yuizevv4hFxBjcg7C7uTZ5fGypuB5574Qm1Mp1SSWEzGBFtBPvnSN3w8TLtvGYML5MBhDcYisuxpo1kaQ92xavcRTYohYi",
		//"zxPiJnvqXCYgkSYfjPPg2y8EBbSTVE41We2Y3LxhaMMqzsr6i9vE8hS8kQAXw9DML3nMiwxqE9StWPVLC9ibQCbPfxWzBuM6VfvEtnLPKiDe9Kgo7Ap9ARKzUM1LKChcDM7",
		//"zxtHrpvVY1nAMjBcKGV6fsjaLUn7thruPWHMr2ckRhW7ZUphgN9RbaUpnr5pu4t3xJEAEqhUcaVSztpFZS9f2Cr6nxr9dBvVUgRGA9czvMBWCmiWPYb8z1XSiruJQgWC5XL",

		"7EHTPNYhKNuULtwQEgFK3NuYbf3qAGNoowRHo5BHZij3mdB7WJxZ4oRJt91HbVL88pxDmBV159MsTjiwzRMD7FgqideToxcNK63VPU7LJ9ff37kJ38Yx41cSBXgdAhFRwJy",
	}

	walletID := "W3TuDqe8VShgyPcg2dw4FRrNQbmxxiGPTJ"
	accountID := "4gbLYX9shEoABGKaZrbTmAfHRXPKkVK6wudFEp7miNFZ7F9ZCL6t38Nr6tSr8GDS11tNZn7iwghsbt2qs6P1bkje"

	contract := openwallet.SmartContract{
		Address:  "AIPP",
		Symbol:   "SERO",
		Name:     "AIPP",
		Token:    "AIPP",
		Decimals: 18,
	}

	testGetAssetsAccountBalance(tw, walletID, accountID)

	testGetAssetsAccountTokenBalance(tw, walletID, accountID, contract)

	for _, to := range targets {

		rawTx, err := testCreateTransactionStep(tw, walletID, accountID, to, "0.01", "", &contract)
		if err != nil {
			return
		}

		_, err = testSignTransactionStep(tw, rawTx)
		if err != nil {
			return
		}

		_, err = testVerifyTransactionStep(tw, rawTx)
		if err != nil {
			return
		}

		_, err = testSubmitTransactionStep(tw, rawTx)
		if err != nil {
			return
		}

	}

}

func TestSummary(t *testing.T) {

	walletID := "W3TuDqe8VShgyPcg2dw4FRrNQbmxxiGPTJ"
	accountID := "3D58HdM35ZJrMgAzzRduGy6mPVqc8yeGNFm5kNBU16tZYkf84N9C4uppHtJWfw6bEXMtkFgXTxPnw3kN9m7QhiX2"
	summaryAddress := "GUZ3K7iLmyuUnV4sGMZDG5E5GXd9dPanhHGAmxNnKVjfn9JvkMSPmm5PWXpmXjjhE2uzAWVaoPApqsLcqi9iawwinuoDp68vWC9Z4FPjpDMV4jVWL1kbVWLEAL6tzcN3gbv"

	testGetAssetsAccountBalance(tw, walletID, accountID)

	rawTxArray, err := testCreateSummaryTransactionStep(tw, walletID, accountID,
		summaryAddress, "", "", "",
		0, 100, nil, nil)
	if err != nil {
		log.Errorf("CreateSummaryTransaction failed, unexpected error: %v", err)
		return
	}

	//执行汇总交易
	for _, rawTxWithErr := range rawTxArray {

		if rawTxWithErr.Error != nil {
			log.Error(rawTxWithErr.Error.Error())
			continue
		}

		_, err = testSignTransactionStep(tw, rawTxWithErr.RawTx)
		if err != nil {
			return
		}

		_, err = testVerifyTransactionStep(tw, rawTxWithErr.RawTx)
		if err != nil {
			return
		}

		_, err = testSubmitTransactionStep(tw, rawTxWithErr.RawTx)
		if err != nil {
			return
		}
	}

}

func TestSummary_Token(t *testing.T) {
	//tm := testInitWalletManager()
	walletID := "W3TuDqe8VShgyPcg2dw4FRrNQbmxxiGPTJ"
	accountID := "3D58HdM35ZJrMgAzzRduGy6mPVqc8yeGNFm5kNBU16tZYkf84N9C4uppHtJWfw6bEXMtkFgXTxPnw3kN9m7QhiX2"
	summaryAddress := "GUZ3K7iLmyuUnV4sGMZDG5E5GXd9dPanhHGAmxNnKVjfn9JvkMSPmm5PWXpmXjjhE2uzAWVaoPApqsLcqi9iawwinuoDp68vWC9Z4FPjpDMV4jVWL1kbVWLEAL6tzcN3gbv"

	contract := openwallet.SmartContract{
		Address:  "AIPP",
		Symbol:   "SERO",
		Name:     "AIPP",
		Token:    "AIPP",
		Decimals: 18,
	}

	testGetAssetsAccountBalance(tw, walletID, accountID)

	testGetAssetsAccountTokenBalance(tw, walletID, accountID, contract)

	rawTxArray, err := testCreateSummaryTransactionStep(tw, walletID, accountID,
		summaryAddress, "", "", "",
		0, 100, &contract, nil)
	if err != nil {
		log.Errorf("CreateSummaryTransaction failed, unexpected error: %v", err)
		return
	}

	//执行汇总交易
	for _, rawTxWithErr := range rawTxArray {

		if rawTxWithErr.Error != nil {
			log.Error(rawTxWithErr.Error.Error())
			continue
		}

		_, err = testSignTransactionStep(tw, rawTxWithErr.RawTx)
		if err != nil {
			return
		}

		_, err = testVerifyTransactionStep(tw, rawTxWithErr.RawTx)
		if err != nil {
			return
		}

		_, err = testSubmitTransactionStep(tw, rawTxWithErr.RawTx)
		if err != nil {
			return
		}
	}

}
