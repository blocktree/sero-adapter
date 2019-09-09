package sero_addrdec

import (
	"fmt"
	"github.com/blocktree/openwallet/hdkeystore"
	"github.com/blocktree/openwallet/openwallet"
	"github.com/blocktree/sero-adapter/client"
	"github.com/mr-tron/base58"
	"github.com/sero-cash/go-sero/common/hexutil"
)

var (
	Default = &AddressDecoderV2{}
)

//AddressDecoderV2
type AddressDecoderV2 struct {
	*openwallet.AddressDecoderV2Base
	IsTestNet bool
	Client *client.Client
}

// AddressDecode decode address
func (dec *AddressDecoderV2) AddressDecode(pubKey string, opts ...interface{}) ([]byte, error) {

	return base58.Decode(pubKey)
}

// AddressEncode encode address
func (dec *AddressDecoderV2) AddressEncode(hash []byte, opts ...interface{}) (string, error) {
	pkrAddress := base58.Encode(hash[:])
	return pkrAddress, nil
}

func (dec *AddressDecoderV2) CustomCreateAddress(account *openwallet.AssetsAccount, newIndex uint64) (*openwallet.Address, error) {

	if dec.Client == nil {
		return nil, fmt.Errorf("sero client is nil")
	}

	rnd, err := hdkeystore.GenerateSeed(32)
	if err != nil {
		return nil, err
	}

	address, err := dec.Client.LocalPk2Pkr(account.PublicKey, hexutil.Encode(rnd))
	if err != nil {
		return nil, err
	}

	//publickey, err := base58.Decode(account.PublicKey)
	//if err != nil {
	//	result.Success = false
	//	result.Err = err
	//	return result
	//}
	//
	//var pk keys.Uint512
	//copy(pk[:], publickey[:])
	//rnd := keys.RandUint256()
	//pkr := keys.Addr2PKr(&pk, &rnd) //收款码
	//address := base58.Encode(pkr[:])

	newAddr := &openwallet.Address{
		AccountID: account.AccountID,
		Symbol:    account.Symbol,
		Index:     newIndex,
		Address:   address,
		Balance:   "0",
		WatchOnly: false,
		PublicKey: account.PublicKey,
		Alias:     "",
		Tag:       "",
		HDPath:    "",
		IsChange:  false,
	}

	return newAddr, nil
}

// SupportCustomCreateAddressFunction 支持创建地址实现
func (dec *AddressDecoderV2) SupportCustomCreateAddressFunction() bool {
	return true
}