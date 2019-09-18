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
	"encoding/json"
	"fmt"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/blocktree/openwallet/hdkeystore"
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openwallet"
	"github.com/blocktree/sero-adapter/client"
	"github.com/blocktree/sero-adapter/sero_addrdec"
	"github.com/mr-tron/base58"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"math/big"
)

type WalletManager struct {
	openwallet.AssetsAdapterBase
	Config          *WalletConfig                   // 节点配置
	Decoder         *sero_addrdec.AddressDecoderV2  //地址编码器
	TxDecoder       openwallet.TransactionDecoder   //交易单编码器
	Log             *log.OWLogger                   //日志工具
	ContractDecoder openwallet.SmartContractDecoder //智能合约解析器
	Blockscanner    *SEROBlockScanner               //区块扫描器
	WalletClient    *client.Client                  // 节点客户端
	unspentDB       *storm.DB                       //未花记录数据库
	blockChainDB    *storm.DB                       //区块链数据库
}

func NewWalletManager() *WalletManager {
	wm := WalletManager{}
	wm.Config = NewConfig(Symbol)
	wm.Blockscanner = NewSEROBlockScanner(&wm)
	wm.Decoder = &sero_addrdec.AddressDecoderV2{}
	wm.TxDecoder = NewTransactionDecoder(&wm)
	wm.Log = log.NewOWLogger(wm.Symbol())
	wm.ContractDecoder = NewContractDecoder(&wm)

	return &wm
}

// 创建钱包 CreateWallet
func (wm *WalletManager) CreateWallet(name, password, keydir string) (*hdkeystore.HDKey, string, error) {

	//随机生成keystore
	return hdkeystore.StoreHDKey(
		keydir,
		name,
		password,
		hdkeystore.StandardScryptN,
		hdkeystore.StandardScryptP,
	)
}

// 创建账户 CreateAccount
func (wm *WalletManager) CreateAccount(alias string, key *hdkeystore.HDKey, newAccIndex uint64) (*openwallet.AssetsAccount, error) {

	account := &openwallet.AssetsAccount{}

	account.Alias = alias
	account.Symbol = wm.Symbol()
	account.Required = 1
	// root/n' , 使用强化方案
	account.HDPath = fmt.Sprintf("%s/%d'", key.RootPath, newAccIndex)

	childKey, err := key.DerivedKeyWithPath(account.HDPath, wm.CurveType())
	if err != nil {
		return nil, err
	}

	//账户扩展得到的私钥
	owprv, err := childKey.GetPrivateKeyBytes()
	if err != nil {
		return nil, err
	}

	//把账户扩展得到的私钥作为sero的种子
	sk, err := wm.LocalSeed2Sk(hexutil.Encode(owprv[:])) //私钥
	if err != nil {
		return nil, err
	}

	tk, err := wm.LocalSk2Tk(sk) //跟踪公钥
	if err != nil {
		return nil, err
	}

	pk, err := wm.LocalTk2Pk(tk) //公钥
	if err != nil {
		return nil, err
	}

	account.AccountID = tk //跟踪公钥作为accountID
	account.PublicKey = pk //sero公钥作为账户公钥
	account.Index = newAccIndex
	account.AddressIndex = -1
	account.WalletID = key.KeyID

	return account, nil

}


// 创建账户 CreateRootAccount
func (wm *WalletManager) CreateRootAccount(alias string, key *hdkeystore.HDKey) (*openwallet.AssetsAccount, error) {

	account := &openwallet.AssetsAccount{}

	account.Alias = alias
	account.Symbol = wm.Symbol()
	account.Required = 1
	// root/n' , 使用强化方案
	account.HDPath = key.RootPath

	childKey, err := key.DerivedKeyWithPath(account.HDPath, wm.CurveType())
	if err != nil {
		return nil, err
	}

	//账户扩展得到的私钥
	owprv, err := childKey.GetPrivateKeyBytes()
	if err != nil {
		return nil, err
	}

	//把账户扩展得到的私钥作为sero的种子
	sk, err := wm.LocalSeed2Sk(hexutil.Encode(owprv[:])) //私钥
	if err != nil {
		return nil, err
	}

	tk, err := wm.LocalSk2Tk(sk) //跟踪公钥
	if err != nil {
		return nil, err
	}

	pk, err := wm.LocalTk2Pk(tk) //公钥
	if err != nil {
		return nil, err
	}

	account.AccountID = tk //跟踪公钥作为accountID
	account.PublicKey = pk //sero公钥作为账户公钥
	account.Index = 0
	account.AddressIndex = -1
	account.WalletID = key.KeyID

	return account, nil

}


func (wm *WalletManager) CreateAddress(account *openwallet.AssetsAccount, newIndex uint64) (*openwallet.Address, error) {
	return wm.Decoder.CustomCreateAddress(account, newIndex)
}

func (wm WalletManager) CreateFixAddress(account *openwallet.AssetsAccount, rnd []byte, newIndex uint64) (*openwallet.Address, error) {
	return wm.Decoder.CreateFixAddress(account, rnd, newIndex)
}

func (wm *WalletManager) GetBlocksInfo(height uint64) (*Block, error) {
	request := []interface{}{
		height,
		1,
	}

	result, err := wm.WalletClient.Call("flight_getBlocksInfo", request)
	if err != nil {
		return nil, err
	}

	var blocks []*Block
	err = json.Unmarshal([]byte(result.Raw), &blocks)
	if err != nil {
		return nil, err
	}

	if len(blocks) == 0 {
		return nil, fmt.Errorf("block info can not find")
	}

	return blocks[0], nil
}

func (wm *WalletManager) GetBlockByNumber(height uint64) (*BlockData, error) {

	request := []interface{}{
		//height,
		hexutil.EncodeUint64(height),
		false,
	}

	result, err := wm.WalletClient.Call("sero_getBlockByNumber", request)
	if err != nil {
		return nil, err
	}

	return NewBlock(result), nil
}

//GetBlockHeight 获取区块链高度
func (wm *WalletManager) GetBlockHeight() (uint64, error) {

	result, err := wm.WalletClient.Call("sero_blockNumber", nil)
	if err != nil {
		return 0, err
	}

	blockNum, err := hexutil.DecodeUint64(result.String())
	if err != nil {
		return 0, err
	}

	return blockNum, nil
}

func (wm *WalletManager) GetTransactionByHash(txid string) (*gjson.Result, error) {

	request := []interface{}{
		txid,
	}

	result, err := wm.WalletClient.Call("flight_getTx", request)
	if err != nil {
		return nil, err
	}

	//blockNum, err := hexutil.DecodeUint64(result.String())
	//if err != nil {
	//	return 0, err
	//}

	return result, nil
}

func (wm *WalletManager) GetOut(root string) (*Out, error) {

	request := []interface{}{
		root,
	}

	result, err := wm.WalletClient.Call("flight_getOut", request)
	if err != nil {
		return nil, err
	}

	var out Out
	err = json.Unmarshal([]byte(result.Raw), &out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

func (wm *WalletManager) DecOut(outs []Out, tkBytes []byte) ([]TDOut, error) {

	request := []interface{}{
		outs,
		hexutil.Encode(tkBytes),
	}

	result, err := wm.WalletClient.Call("local_decOut", request)
	if err != nil {
		return nil, err
	}

	var douts []TDOut
	err = json.Unmarshal([]byte(result.Raw), &douts)
	if err != nil {
		return nil, err
	}

	//tk := Uint512{}
	//copy(tk[:], tkBytes)
	//douts := flight.DecOut(&tk, outs)
	return douts, nil
}

// GasPrice 费率
func (wm *WalletManager) GasPrice() (*big.Int, error) {
	result, err := wm.WalletClient.Call("sero_gasPrice", nil)
	if err != nil {
		return nil, err
	}
	return hexutil.DecodeBig(result.String())
}

// ListUnspentByAddress 未花记录
func (wm *WalletManager) ListUnspentByAddress(address, currency string, offset, limit int) ([]*Unspent, error) {

	var (
		utxo []*Unspent
		err error
	)

	if limit > 0 {
		err = wm.unspentDB.Select(
			q.And(
				q.Eq("Address", address),
				q.Eq("Currency", currency),
			)).Limit(limit).Skip(offset).Find(&utxo)
	} else {
		err = wm.unspentDB.Select(
			q.And(
				q.Eq("Address", address),
				q.Eq("Currency", currency),
			)).Skip(offset).Find(&utxo)
	}

	if err != nil {
		return nil, err
	}

	return utxo, nil
}

// ListUnspent 未花记录
func (wm *WalletManager) ListUnspent(tk string, currency string, offset, limit int) ([]*Unspent, error) {

	var (
		utxo []*Unspent
		err error
	)

	if limit > 0 {
		err = wm.unspentDB.Select(
			q.And(
				q.Eq("TK", tk),
				q.Eq("Currency", currency),
			)).Limit(limit).Skip(offset).Find(&utxo)
	} else {
		err = wm.unspentDB.Select(
			q.And(
				q.Eq("TK", tk),
				q.Eq("Currency", currency),
			)).Skip(offset).Find(&utxo)
	}
	if err != nil {
		return nil, err
	}

	return utxo, nil
}

//EstimateFee 预估手续费
func (wm *WalletManager) EstimateFee(feeRate decimal.Decimal) (decimal.Decimal, decimal.Decimal, error) {

	var (
		fees = decimal.Zero
	)

	if feeRate.LessThanOrEqual(decimal.Zero) {
		gasPrice, err := wm.GasPrice()
		if err != nil {
			return fees, decimal.Zero, err
		}

		feeRate = decimal.NewFromBigInt(gasPrice, 0)
		feeRate = feeRate.Shift(-wm.Decimal())
	}

	//fees = gasPrice * fixGas
	fees = feeRate.Mul(decimal.New(wm.Config.FixGas, 0))

	return fees, feeRate, nil
}

// GenTxParam 构建交易
func (wm *WalletManager) GenTxParam(
	from, tk string,
	decimals int32, feesRate decimal.Decimal,
	usedUTXO []*Unspent,
	to []Out_O) (*gjson.Result, error) {

	/*
		"params": [{    //参数1：预组装交易结构
		        "From": "0xb8d01......143099",           //找零收款码PKr，需要与TK配套。
		        "Gas": 25000,                            //最大Gas消耗
		        "GasPrice": 1000000000,                  //Gas价格，最小为1Gta
		        "Ins": ["0x7b30c......0fbb122e"],        //输入UTXO的Root列表，需要确保是TK下面的UTXO。
		        "Outs": [{                               //输出
		            "Asset": {                             //资产对象
		                "Tkn": {                                     //Token对象（同质化通证）
		                    "Currency": "0x00000......005345524f",     //币种Id，去掉0之后就是ASCII字符串 "SERO"
		                    "Value": "500000000000000000"              //币种金额
		                }
		            },
		            "PKr": "0x3b78d......4603daa"          //输出的收款码PKr
		        }]
		    },
		  "tu1nEPY......As8Ht4z"   //参数2：跟踪秘钥TK
		  ]

	*/

	type Tkn struct {
		Currency string `json:"Currency"`
		Value    string `json:"Value"`
	}

	type Asset struct {
		Tkn Tkn `json:"Tkn"`
	}

	type Out struct {
		Asset Asset  `json:"Asset"`
		PKr   string `json:"PKr"`
	}

	ins := make([]string, 0)
	for _, u := range usedUTXO {
		ins = append(ins, u.Root)
	}

	outs := make([]interface{}, 0)
	for _, output := range to {
		pkr, _ := base58.Decode(output.Addr)

		currencyID, err := wm.LocalCurrencyToId(output.Asset.Tkn.Currency)
		if err != nil {
			return nil, err
		}

		out := Out{
			Asset: Asset{
				Tkn: Tkn{
					Currency: currencyID,
					Value:    output.Asset.Tkn.Value,
				},
			},
			PKr: hexutil.Encode(pkr),
		}
		outs = append(outs, out)
	}

	fromHex, _ := base58.Decode(from)
	gasPrice := feesRate.Shift(wm.Decimal()).IntPart()
	payload := map[string]interface{}{
		"From":     hexutil.Encode(fromHex),
		"Gas":      wm.Config.FixGas,
		"GasPrice": gasPrice,
		"Ins":      ins,
		"Outs":     outs,
	}

	request := []interface{}{
		payload,
		tk,
	}

	result, err := wm.WalletClient.Call("flight_genTxParam", request)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// SignTxWithSk 签名交易
func (wm *WalletManager) SignTxWithSk(txJSON string, seed []byte) (*gjson.Result, error) {

	sk, err := wm.LocalSeed2Sk(hexutil.Encode(seed))
	if err != nil {
		return nil, err
	}

	var txStruct map[string]interface{}
	err = json.Unmarshal([]byte(txJSON), &txStruct)
	if err != nil {
		return nil, err
	}

	request := []interface{}{
		txStruct,
		sk,
	}

	result, err := wm.WalletClient.Call("local_signTxWithSk", request)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// CommitTx 广播交易
func (wm *WalletManager) CommitTx(txJSON string) (string, error) {

	var txStruct map[string]interface{}
	err := json.Unmarshal([]byte(txJSON), &txStruct)
	if err != nil {
		return "", err
	}

	request := []interface{}{
		txStruct,
	}

	_, err = wm.WalletClient.Call("flight_commitTx", request)
	if err != nil {
		return "", err
	}

	txid := txStruct["Hash"].(string)

	return txid, nil
}

func (wm *WalletManager) LocalSeed2Sk(seed string) (string, error) {
	return wm.WalletClient.LocalSeed2Sk(seed)
}

func (wm *WalletManager) LocalSk2Tk(sk string) (string, error) {
	return wm.WalletClient.LocalSk2Tk(sk)
}

func (wm *WalletManager) LocalTk2Pk(tk string) (string, error) {
	return wm.WalletClient.LocalTk2Pk(tk)
}

func (wm *WalletManager) LocalPk2Pkr(pk, rnd string) (string, error) {
	return wm.WalletClient.LocalPk2Pkr(pk, rnd)
}

func (wm *WalletManager) LocalCurrencyToId(name string) (string, error) {

	request := []interface{}{
		name,
	}

	result, err := wm.WalletClient.Call("local_currencyToId", request)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

func (wm *WalletManager) LocalIdToCurrency(id string) (string, error) {

	request := []interface{}{
		id,
	}

	result, err := wm.WalletClient.Call("local_idToCurrency", request)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}
