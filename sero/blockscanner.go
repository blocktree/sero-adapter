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
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openwallet"
	"github.com/mr-tron/base58"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"strings"
	"time"
)

const (
	maxExtractingSize = 10 // thread count
)

//EOSBlockScanner EOS block scanner
type SEROBlockScanner struct {
	*openwallet.BlockScannerBase

	CurrentBlockHeight   uint64         //当前区块高度
	extractingCH         chan struct{}  //扫描工作令牌
	wm                   *WalletManager //钱包管理者
	IsScanMemPool        bool           //是否扫描交易池
	RescanLastBlockCount uint64         //重扫上N个区块数量
}

type ExtractOutput map[string][]*openwallet.TxOutPut
type ExtractInput map[string][]*openwallet.TxInput
type ExtractData map[string]*openwallet.TxExtractData

//ExtractResult extract result
type ExtractResult struct {
	extractData map[string]ExtractData
	TxID        string
	BlockHash   string
	BlockHeight uint64
	BlockTime   int64
	Success     bool
}

//SaveResult result
type SaveResult struct {
	TxID        string
	BlockHeight uint64
	Success     bool
}

// NewSEROBlockScanner create a block scanner
func NewSEROBlockScanner(wm *WalletManager) *SEROBlockScanner {
	bs := SEROBlockScanner{
		BlockScannerBase: openwallet.NewBlockScannerBase(),
	}

	bs.extractingCH = make(chan struct{}, maxExtractingSize)
	bs.wm = wm
	bs.IsScanMemPool = true
	bs.RescanLastBlockCount = 0

	// set task
	bs.SetTask(bs.ScanBlockTask)

	return &bs
}

//ScanBlockTask scan block task
func (bs *SEROBlockScanner) ScanBlockTask() {

	//获取本地区块高度
	blockHeader, err := bs.GetScannedBlockHeader()
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not get new block height; unexpected error: %v", err)
		return
	}

	currentHeight := blockHeader.Height
	currentHash := blockHeader.Hash

	for {

		/*
			1. 读取当前最大高度。
			2. 读取当前已扫描高度。
			3. 当前高度+1，扫描区块。
			4. 读取getBlocksInfo。
			5. 逐个提取交易单，把getBlocksInfo相关的密文output拿出来。
			6. 通过output地址查找账户id。
			7. 通过账户id解密output，得到utxo，utxo以root作为id储存到数据库，每个nil作为id与root关联。
			8. 通过from地址查找账户id，合计output的数量，构建一个input对象。
			9. 手续费构建一个input对象。
			10. 提取input，output，transaction结果。
			11. 所有交易单都提取完后，读取Nils数组，查找数据库中的nil存在的记录，查找nil关联root的utxo，并删除。
		*/

		if !bs.Scanning {
			// stop scan
			return
		}

		maxBlockHeight, err := bs.wm.GetBlockHeight()
		if err != nil {
			bs.wm.Log.Errorf("get chain info failed, err=%v", err)
			break
		}

		bs.wm.Log.Info("current block height:", currentHeight, " maxBlockHeight:", maxBlockHeight)
		if currentHeight == maxBlockHeight {
			bs.wm.Log.Std.Info("block scanner has scanned full chain data. Current height %d", maxBlockHeight)
			break
		}

		// next block
		currentHeight = currentHeight + 1

		bs.wm.Log.Std.Info("block scanner scanning height: %d ...", currentHeight)
		block, err := bs.wm.GetBlockByNumber(currentHeight)

		if err != nil {
			bs.wm.Log.Std.Info("block scanner can not get new block data by rpc; unexpected error: %v", err)
			break
		}

		isFork := false

		if currentHash != block.ParentHash {
			bs.wm.Log.Std.Info("block has been fork on height: %d.", currentHeight)
			bs.wm.Log.Std.Info("block height: %d local hash = %s ", currentHeight-1, currentHash)
			bs.wm.Log.Std.Info("block height: %d mainnet hash = %s ", currentHeight-1, block.ParentHash)
			bs.wm.Log.Std.Info("delete recharge records on block height: %d.", currentHeight-1)

			//查询本地分叉的区块
			forkBlock, _ := bs.GetLocalBlock(currentHeight - 1)

			//删除上一区块链的所有充值记录
			//bs.DeleteRechargesByHeight(currentHeight - 1)
			//删除上一区块的未扫记录
			bs.DeleteUnscanRecord(currentHeight - 1)
			//删除上一区块的未花记录
			bs.DeleteUnspentByHeight(currentHeight - 1)
			currentHeight = currentHeight - 2 //倒退2个区块重新扫描
			if currentHeight <= 0 {
				currentHeight = 1
			}

			localBlock, err := bs.GetLocalBlock(currentHeight)
			if err != nil {
				bs.wm.Log.Std.Error("block scanner can not get local block; unexpected error: %v", err)

				localBlock, err = bs.wm.GetBlockByNumber(currentHeight)
				if err != nil {
					bs.wm.Log.Std.Error("block scanner can not get prev block; unexpected error: %v", err)
					break
				}

			}

			//重置当前区块的hash
			currentHash = localBlock.BlockHash

			bs.wm.Log.Std.Info("rescan block on height: %d, hash: %s .", currentHeight, currentHash)

			//重新记录一个新扫描起点
			bs.SaveLocalBlockHead(localBlock.BlockNumber, localBlock.BlockHash)

			isFork = true

			if forkBlock != nil {

				//通知分叉区块给观测者，异步处理
				bs.newBlockNotify(forkBlock, isFork)
			}

		} else {

			currentHash = block.BlockHash
			err := bs.BatchExtractTransactions(block)
			if err != nil {
				bs.wm.Log.Std.Error("block scanner ran BatchExtractTransactions occured unexpected error: %v", err)
				break
			}

			//保存本地新高度
			bs.SaveLocalBlockHead(currentHeight, currentHash)
			bs.SaveLocalBlock(block)

			isFork = false

			//通知新区块给观测者，异步处理
			bs.newBlockNotify(block, isFork)
		}
	}
}

//newBlockNotify 获得新区块后，通知给观测者
func (bs *SEROBlockScanner) newBlockNotify(block *BlockData, isFork bool) {
	header := block.BlockHeader(bs.wm.Symbol())
	header.Fork = isFork
	bs.NewBlockNotify(header)
}

// BatchExtractTransactions 批量提取交易单
func (bs *SEROBlockScanner) BatchExtractTransactions(block *BlockData) error {

	var (
		quit       = make(chan struct{})
		done       = 0 //完成标记
		failed     = 0
		shouldDone = len(block.transactions) //需要完成的总数
	)

	if len(block.transactions) == 0 {
		return nil
	}

	//查询该高度的utxo和作废码信息
	blockInfo, err := bs.wm.GetBlocksInfo(block.BlockNumber)
	if err != nil {
		return err
	}

	block.blockInfo = blockInfo

	//先作废已使用的utxo
	for _, nilKey := range block.blockInfo.Nils {
		nilErr := bs.DeleteUnspent(nilKey)
		if nilErr != nil {
			return nilErr
		}
	}

	bs.wm.Log.Std.Info("block scanner ready extract transactions total: %d ", len(block.transactions))

	//生产通道
	producer := make(chan ExtractResult)
	defer close(producer)

	//消费通道
	worker := make(chan ExtractResult)
	defer close(worker)

	//保存工作
	saveWork := func(height uint64, result chan ExtractResult) {
		//回收创建的地址
		for gets := range result {

			if gets.Success {
				notifyErr := bs.newExtractDataNotify(height, gets.extractData)
				if notifyErr != nil {
					failed++ //标记保存失败数
					bs.wm.Log.Std.Info("newExtractDataNotify unexpected error: %v", notifyErr)
				}
			} else {
				//记录未扫区块
				unscanRecord := NewUnscanRecord(height, "", "")
				bs.SaveUnscanRecord(unscanRecord)
				failed++ //标记保存失败数
			}
			//累计完成的线程数
			done++
			if done == shouldDone {
				close(quit) //关闭通道，等于给通道传入nil
			}
		}
	}

	//提取工作
	extractWork := func(eblock *BlockData, eProducer chan ExtractResult) {
		for _, txid := range block.transactions {
			bs.extractingCH <- struct{}{}
			//shouldDone++
			go func(mblock *BlockData, mTxid string, end chan struct{}, mProducer chan<- ExtractResult) {

				//导出提出的交易
				mProducer <- bs.ExtractTransaction(mblock, mTxid, bs.ScanTargetFunc)
				//释放
				<-end

			}(eblock, txid, bs.extractingCH, eProducer)
		}
	}

	/*	开启导出的线程	*/

	//独立线程运行消费
	go saveWork(block.BlockNumber, worker)

	//独立线程运行生产
	go extractWork(block, producer)

	//以下使用生产消费模式
	bs.extractRuntime(producer, worker, quit)

	if failed > 0 {
		return fmt.Errorf("block scanner saveWork failed")
	}

	return nil
}

//extractRuntime 提取运行时
func (bs *SEROBlockScanner) extractRuntime(producer chan ExtractResult, worker chan ExtractResult, quit chan struct{}) {

	var (
		values = make([]ExtractResult, 0)
	)

	for {
		var activeWorker chan<- ExtractResult
		var activeValue ExtractResult
		//当数据队列有数据时，释放顶部，传输给消费者
		if len(values) > 0 {
			activeWorker = worker
			activeValue = values[0]
		}
		select {
		//生成者不断生成数据，插入到数据队列尾部
		case pa := <-producer:
			values = append(values, pa)
		case <-quit:
			//退出
			return
		case activeWorker <- activeValue:
			values = values[1:]
		}
	}
	//return
}

// ExtractTransaction 提取交易单
func (bs *SEROBlockScanner) ExtractTransaction(block *BlockData, txid string, scanTargetFunc openwallet.BlockScanTargetFunc) ExtractResult {
	var (
		result = ExtractResult{
			BlockHash:   block.BlockHash,
			BlockHeight: block.BlockNumber,
			TxID:        txid,
			extractData: make(map[string]ExtractData),
			BlockTime:   int64(block.Timestamp),
		}
	)

	//bs.wm.Log.Std.Debug("block scanner scanning tx: %s ...", txid)
	trx, err := bs.wm.GetTransactionByHash(txid)
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not extract transaction data; unexpected error: %v", err)
		result.Success = false
		return result
	}

	bs.extractTransaction(block, trx, &result, scanTargetFunc)

	return result

}

//ExtractTransactionData 提取交易单
func (bs *SEROBlockScanner) extractTransaction(block *BlockData, trx *gjson.Result, result *ExtractResult, scanTargetFunc openwallet.BlockScanTargetFunc) {

	var (
		success = true
	)

	//手续费
	fees, _ := decimal.NewFromString(trx.Get("Tx.Fee.Value").String())
	fees = fees.Shift(-bs.wm.Decimal())

	//先提取output，因为要先解析output，才能知道哪些代币交易
	tokenExtractOutput, isTokenTrasfer, err := bs.extractTxOutput(block, trx, scanTargetFunc)
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not extract transaction data; unexpected error: %v", err)
		result.Success = false
		return
	}

	for token, sourceExtractOutput := range tokenExtractOutput {

		tokenFees := decimal.Zero
		if token == bs.wm.Symbol() {
			tokenFees = fees
		}

		for sourceKey, extractOutput := range sourceExtractOutput {
			var (
				to     = make([]string, 0)
				txType = uint64(0)
				coin   openwallet.Coin
			)
			for _, output := range extractOutput {
				if isTokenTrasfer && token == bs.wm.Symbol() {
					output.TxType = 1
					txType = 1
				}
				coin = output.Coin
				to = append(to, output.Address+":"+output.Amount)
			}

			tx := &openwallet.Transaction{
				From:        []string{},
				To:          to,
				Fees:        tokenFees.String(),
				Coin:        coin,
				BlockHash:   block.BlockHash,
				BlockHeight: block.BlockNumber,
				TxID:        trx.Get("Hash").String(),
				Decimal:     bs.wm.Decimal(),
				ConfirmTime: int64(block.Timestamp),
				Status:      openwallet.TxStatusSuccess,
				TxType:      txType,
			}
			wxID := openwallet.GenTransactionWxID(tx)
			tx.WxID = wxID

			sourceKeyExtractData := result.extractData[token]
			if sourceKeyExtractData == nil {
				sourceKeyExtractData = make(ExtractData)
			}

			extractData := sourceKeyExtractData[sourceKey]
			if extractData == nil {
				extractData = &openwallet.TxExtractData{}
			}

			extractData.TxOutputs = extractOutput
			extractData.Transaction = tx

			sourceKeyExtractData[sourceKey] = extractData
			result.extractData[token] = sourceKeyExtractData
		}
	}

	//提取input部分记录
	tokenExtractInput, err := bs.extractTxInput(block, trx, scanTargetFunc)
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not extract transaction data; unexpected error: %v", err)
		result.Success = false
		return
	}

	for token, sourceExtractInput := range tokenExtractInput {

		tokenFees := decimal.Zero
		if token == bs.wm.Symbol() {
			tokenFees = fees
		}

		for sourceKey, extractInput := range sourceExtractInput {
			var (
				from   = make([]string, 0)
				txType = uint64(0)
				coin   openwallet.Coin
			)
			for _, input := range extractInput {
				if isTokenTrasfer && token == bs.wm.Symbol() {
					input.TxType = 1
					txType = 1
				}
				coin = input.Coin
				if !(token == bs.wm.Symbol() && input.Index == 0) {
					from = append(from, input.Address+":"+input.Amount)
				}

			}

			sourceKeyExtractData := result.extractData[token]
			if sourceKeyExtractData == nil {
				sourceKeyExtractData = make(ExtractData)
			}

			extractData := sourceKeyExtractData[sourceKey]
			if extractData == nil {
				extractData = &openwallet.TxExtractData{}
			}

			extractData.TxInputs = extractInput
			if extractData.Transaction == nil {
				extractData.Transaction = &openwallet.Transaction{
					From:        from,
					To:          []string{},
					Fees:        tokenFees.String(),
					Coin:        coin,
					BlockHash:   block.BlockHash,
					BlockHeight: block.BlockNumber,
					TxID:        trx.Get("Hash").String(),
					Decimal:     bs.wm.Decimal(),
					ConfirmTime: int64(block.Timestamp),
					Status:      openwallet.TxStatusSuccess,
					TxType:      txType,
				}
			} else {
				extractData.Transaction.From = from
			}

			sourceKeyExtractData[sourceKey] = extractData
			result.extractData[token] = sourceKeyExtractData
		}
	}

	result.Success = success
}

//ExtractTxInput 提取交易单输入部分
func (bs *SEROBlockScanner) extractTxInput(block *BlockData, trx *gjson.Result, scanTargetFunc openwallet.BlockScanTargetFunc) (map[string]ExtractInput, error) {

	var (
		tokenExtractInput = make(map[string]ExtractInput)
	)

	createAt := time.Now().Unix()

	//手续费
	fees, _ := decimal.NewFromString(trx.Get("Tx.Fee.Value").String())
	fees = fees.Shift(-bs.wm.Decimal())
	txid := trx.Get("Hash").String()

	hex, _ := hexutil.Decode(trx.Get("Tx.From").String())
	address := base58.Encode(hex)
	sourceKey, ok := scanTargetFunc(openwallet.ScanTarget{
		Address:          address,
		Symbol:           bs.wm.Symbol(),
		BalanceModelType: openwallet.BalanceModelTypeAddress})
	if ok {
		bs.wm.Log.Infof("scanTargetFunc found: %s", sourceKey)
		//组装一个SERO手续费作为输入
		feesInput := openwallet.TxInput{}
		feesInput.Coin = openwallet.Coin{
			Symbol:     bs.wm.Symbol(),
			IsContract: false,
		}
		feesInput.TxID = txid
		feesInput.Amount = fees.String()
		feesInput.Address = address
		feesInput.Index = 0
		feesInput.Sid = openwallet.GenTxOutPutSID(txid, bs.wm.Symbol(), "", 0)
		feesInput.CreateAt = createAt
		feesInput.BlockHeight = block.BlockNumber
		feesInput.BlockHash = block.BlockHash

		sourceKeyExtractInput := tokenExtractInput[bs.wm.Symbol()]
		if sourceKeyExtractInput == nil {
			sourceKeyExtractInput = make(ExtractInput)
		}

		extractInput := sourceKeyExtractInput[sourceKey]
		if extractInput == nil {
			extractInput = make([]*openwallet.TxInput, 0)
		}

		extractInput = append(extractInput, &feesInput)

		sourceKeyExtractInput[sourceKey] = extractInput
		tokenExtractInput[bs.wm.Symbol()] = sourceKeyExtractInput

		if bs.WalletDAI != nil {
			//因为输入的金额匿名，需要查钱包系统的数据库中的交易单
			txsDB, _ := bs.WalletDAI.GetTransactionByTxID(txid, bs.wm.Symbol())
			if txsDB != nil {
				bs.wm.Log.Infof("GetTransactionByTxID found: %s", txid)
				for _, txDB := range txsDB {
					currency := ""
					if txDB.Coin.IsContract {
						currency = txDB.Coin.Contract.Address
					} else {
						currency = txDB.Coin.Symbol
					}

					for i, f := range txDB.From {
						index := uint64(i + 1)
						fv := strings.Split(f, ":")
						if len(fv) == 2 {
							input := openwallet.TxInput{}
							input.Coin = txDB.Coin
							input.TxID = txDB.TxID
							input.Amount = fv[1]
							input.Address = fv[0]
							input.Index = index
							input.Sid = openwallet.GenTxOutPutSID(txid, bs.wm.Symbol(), txDB.Coin.ContractID, index)
							input.CreateAt = createAt
							input.BlockHeight = block.BlockNumber
							input.BlockHash = block.BlockHash

							sourceKeyExtractInput := tokenExtractInput[currency]
							if sourceKeyExtractInput == nil {
								sourceKeyExtractInput = make(ExtractInput)
							}

							extractInput := sourceKeyExtractInput[sourceKey]
							if extractInput == nil {
								extractInput = make([]*openwallet.TxInput, 0)
							}

							extractInput = append(extractInput, &input)

							sourceKeyExtractInput[sourceKey] = extractInput
							tokenExtractInput[currency] = sourceKeyExtractInput
						}
					}
				}
			}
		}

	}

	return tokenExtractInput, nil
}

//ExtractTxInput 提取交易单输出部分
func (bs *SEROBlockScanner) extractTxOutput(block *BlockData, trx *gjson.Result, scanTargetFunc openwallet.BlockScanTargetFunc) (map[string]ExtractOutput, bool, error) {

	var (
		tokenExtractOutput = make(map[string]ExtractOutput)
		isTokenTrasfer     = false
	)

	txid := trx.Get("Hash").String()
	vout := block.GetOutputInfoByTxID(txid)

	//bs.wm.Log.Debug("vout:", vout.Array())
	createAt := time.Now().Unix()
	address := ""
	for i, out := range vout {

		if out.State.OS.Out_Z != nil {
			addr, err := hexutil.Decode(out.State.OS.Out_Z.PKr)
			if err != nil {
				return nil, isTokenTrasfer, err
			}
			address = base58.Encode(addr)
		} else if out.State.OS.Out_O != nil {
			addr, err := hexutil.Decode(out.State.OS.Out_O.Addr)
			if err != nil {
				return nil, isTokenTrasfer, err
			}
			address = base58.Encode(addr)
		} else {
			continue
		}

		sourceKey, ok := scanTargetFunc(openwallet.ScanTarget{
			Address:          address,
			Symbol:           bs.wm.Symbol(),
			BalanceModelType: openwallet.BalanceModelTypeAddress})
		if ok {
			bs.wm.Log.Infof("scanTargetFunc found: %s", sourceKey)
			tkBytes, err := base58.Decode(sourceKey)
			if err != nil {
				return nil, isTokenTrasfer, fmt.Errorf("base58 decode TK failed")
			}

			//通过accountID解密utxo
			var tdOut *TDOut
			if decOuts, _ := bs.wm.DecOut([]Out{out}, tkBytes); decOuts != nil {
				tdOut = &decOuts[0]
			}

			if tdOut == nil {
				return nil, isTokenTrasfer, fmt.Errorf("decode output failed")
			}

			currency, err := bs.wm.LocalIdToCurrency(tdOut.Asset.Tkn.Currency)
			if err != nil {
				return nil, isTokenTrasfer, err
			}

			amount, _ := decimal.NewFromString(tdOut.Asset.Tkn.Value)
			value := amount
			outPut := openwallet.TxOutPut{}
			contractId := openwallet.GenContractID(bs.wm.Symbol(), currency)
			//资产为主链币，计算精度
			if currency == bs.wm.Symbol() {
				amount = amount.Shift(-bs.wm.Decimal())
				outPut.Coin = openwallet.Coin{
					Symbol:     bs.wm.Symbol(),
					IsContract: false,
				}
			} else {
				isTokenTrasfer = true
				outPut.Coin = openwallet.Coin{
					Symbol:     bs.wm.Symbol(),
					IsContract: true,
					ContractID: contractId,
					Contract: openwallet.SmartContract{
						ContractID: contractId,
						Address:    currency,
						Symbol:     bs.wm.Symbol(),
					},
				}
			}
			outPut.TxID = txid
			outPut.Amount = amount.String()
			outPut.Address = address
			outPut.Index = uint64(i)
			outPut.Sid = openwallet.GenTxOutPutSID(txid, bs.wm.Symbol(), contractId, uint64(i))
			outPut.CreateAt = createAt
			outPut.BlockHeight = block.BlockNumber
			outPut.BlockHash = block.BlockHash

			sourceKeyExtractOutput := tokenExtractOutput[currency]
			if sourceKeyExtractOutput == nil {
				sourceKeyExtractOutput = make(ExtractOutput)
			}

			extractOutput := sourceKeyExtractOutput[sourceKey]
			if extractOutput == nil {
				extractOutput = make([]*openwallet.TxOutPut, 0)
			}

			extractOutput = append(extractOutput, &outPut)

			sourceKeyExtractOutput[sourceKey] = extractOutput
			tokenExtractOutput[currency] = sourceKeyExtractOutput

			//新增utxo
			utxo := &Unspent{
				Height:   block.BlockNumber,
				Root:     out.Root,
				Address:  address,
				TK:       sourceKey,
				Currency: currency,
				Value:    value.String(),
			}

			//保存新的utxo记录
			err = bs.SaveUnspent(utxo, tdOut.Nils)
			if err != nil {
				return nil, isTokenTrasfer, fmt.Errorf("save unspent failed")
			}
			bs.wm.Log.Infof("SaveUnspent successed ")
		}
	}

	return tokenExtractOutput, isTokenTrasfer, nil
}

//newExtractDataNotify 发送通知
func (bs *SEROBlockScanner) newExtractDataNotify(height uint64, tokenExtractData map[string]ExtractData) error {

	for o, _ := range bs.Observers {

		for _, extractData := range tokenExtractData {
			for key, data := range extractData {
				err := o.BlockExtractDataNotify(key, data)
				if err != nil {
					bs.wm.Log.Error("BlockExtractDataNotify unexpected error:", err)
					//记录未扫区块
					unscanRecord := NewUnscanRecord(height, "", "ExtractData Notify failed.")
					err = bs.SaveUnscanRecord(unscanRecord)
					if err != nil {
						bs.wm.Log.Std.Error("block height: %d, save unscan record failed. unexpected error: %v", height, err.Error())
					}

				}
			}
		}
	}

	return nil
}

//ScanBlock 扫描指定高度区块
func (bs *SEROBlockScanner) ScanBlock(height uint64) error {

	block, err := bs.scanBlock(height)
	if err != nil {
		return err
	}

	//通知新区块给观测者，异步处理
	bs.newBlockNotify(block, false)

	return nil
}

func (bs *SEROBlockScanner) scanBlock(height uint64) (*BlockData, error) {

	block, err := bs.wm.GetBlockByNumber(height)
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not get new block data; unexpected error: %v", err)

		//记录未扫区块
		unscanRecord := NewUnscanRecord(height, "", err.Error())
		bs.SaveUnscanRecord(unscanRecord)
		bs.wm.Log.Std.Info("block height: %d extract failed.", height)
		return nil, err
	}

	bs.wm.Log.Std.Info("block scanner scanning height: %d ...", block.BlockNumber)

	err = bs.BatchExtractTransactions(block)
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not extractRechargeRecords; unexpected error: %v", err)
	}

	//保存区块
	//bs.wm.SaveLocalBlock(block)

	return block, nil
}

//SetRescanBlockHeight 重置区块链扫描高度
func (bs *SEROBlockScanner) SetRescanBlockHeight(height uint64) error {
	height = height - 1
	if height < 0 {
		return fmt.Errorf("block height to rescan must greater than 0 ")
	}

	block, err := bs.wm.GetBlockByNumber(height)
	if err != nil {
		return err
	}

	bs.SaveLocalBlockHead(height, block.BlockHash)

	return nil
}

// GetGlobalMaxBlockHeight GetGlobalMaxBlockHeight
func (bs *SEROBlockScanner) GetGlobalMaxBlockHeight() uint64 {
	height, err := bs.wm.GetBlockHeight()
	if err != nil {
		bs.wm.Log.Std.Info("get global head block error;unexpected error:%v", err)
		return 0
	}
	return height
}

//GetScannedBlockHeight 获取已扫区块高度
func (bs *SEROBlockScanner) GetScannedBlockHeight() uint64 {
	height, _ := bs.GetLocalBlockHead()
	return height
}

//GetBalanceByAddress 查询地址余额
func (bs *SEROBlockScanner) GetBalanceByAddress(address ...string) ([]*openwallet.Balance, error) {

	addrBalanceArr := make([]*openwallet.Balance, 0)

	for _, addr := range address {
		//log.Infof("GetTokenBalanceByAddress: %s", addr)
		utxo, err := bs.wm.ListUnspentByAddress(addr, bs.wm.Symbol(), 0, -1)
		if err != nil {
			log.Debugf("ListUnspentByAddress failed, err: %v", err)
			//return nil, nil
		}

		obj := &openwallet.Balance{}
		tb := decimal.Zero
		for _, u := range utxo {

			b, _ := decimal.NewFromString(u.Value)
			tb = tb.Add(b.Shift(-bs.wm.Decimal()))

		}

		obj.Symbol = bs.wm.Symbol()
		obj.Address = addr
		obj.ConfirmBalance = tb.String()
		obj.Balance = obj.ConfirmBalance

		addrBalanceArr = append(addrBalanceArr, obj)
	}

	return addrBalanceArr, nil
}

func (bs *SEROBlockScanner) GetCurrentBlockHeader() (*openwallet.BlockHeader, error) {
	height, err := bs.wm.GetBlockHeight()
	if err != nil {
		return nil, err
	}

	block, err := bs.wm.GetBlockByNumber(height)
	if err != nil {
		return nil, err
	}

	return block.BlockHeader(bs.wm.Symbol()), nil
}

//GetScannedBlockHeader 获取当前扫描的区块头
func (bs *SEROBlockScanner) GetScannedBlockHeader() (*openwallet.BlockHeader, error) {

	var (
		blockHeight uint64 = 0
		hash        string
		err         error
	)

	blockHeight, hash = bs.GetLocalBlockHead()

	//如果本地没有记录，查询接口的高度
	if blockHeight == 0 {
		blockHeight, err = bs.wm.GetBlockHeight()
		if err != nil {

			return nil, err
		}

		//就上一个区块链为当前区块
		blockHeight = blockHeight - 1

		block, err := bs.wm.GetBlockByNumber(blockHeight)
		if err != nil {
			return nil, err
		}
		hash = block.BlockHash
	}

	return &openwallet.BlockHeader{Height: blockHeight, Hash: hash}, nil
}
