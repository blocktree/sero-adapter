package openwtester

import (
	"fmt"
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openw"
	"github.com/blocktree/openwallet/openwallet"
	"testing"
)

var (
	tw *openw.WalletManager
)

func init() {
	tw = testInitWalletManager()
}

func testInitWalletManager() *openw.WalletManager {
	log.SetLogFuncCall(true)
	tc := openw.NewConfig()

	tc.ConfigDir = configFilePath
	tc.EnableBlockScan = false
	tc.SupportAssets = []string{
		"SERO",
	}
	return openw.NewWalletManager(tc)
	//tm.Init()
}

func TestWalletManager_CreateWallet(t *testing.T) {
	tm := testInitWalletManager()
	w := &openwallet.Wallet{Alias: "HELLO SERO", IsTrust: true, Password: "12345678"}
	nw, key, err := tm.CreateWallet(testApp, w)
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("wallet:", nw)
	log.Info("key:", key)

}

func TestWalletManager_GetWalletInfo(t *testing.T) {

	tm := testInitWalletManager()

	wallet, err := tm.GetWalletInfo(testApp, "W3TuDqe8VShgyPcg2dw4FRrNQbmxxiGPTJ")
	if err != nil {
		log.Error("unexpected error:", err)
		return
	}
	log.Info("wallet:", wallet)
}

func TestWalletManager_GetWalletList(t *testing.T) {

	tm := testInitWalletManager()

	list, err := tm.GetWalletList(testApp, 0, 10000000)
	if err != nil {
		log.Error("unexpected error:", err)
		return
	}
	for i, w := range list {
		log.Info("wallet[", i, "] :", w)
	}
	log.Info("wallet count:", len(list))

	tm.CloseDB(testApp)
}

func TestWalletManager_CreateAssetsAccount(t *testing.T) {

	tm := testInitWalletManager()

	walletID := "W3TuDqe8VShgyPcg2dw4FRrNQbmxxiGPTJ"
	account := &openwallet.AssetsAccount{Alias: "mainnetSERO", WalletID: walletID, Required: 1, Symbol: "SERO", IsTrust: true}
	account, address, err := SERO_CreateAssetsAccount(testApp, walletID, "12345678", account, tm)
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("account:", account)
	log.Info("address:", address)

	tm.CloseDB(testApp)
}

func TestWalletManager_GetAssetsAccountList(t *testing.T) {

	tm := testInitWalletManager()

	walletID := "W3TuDqe8VShgyPcg2dw4FRrNQbmxxiGPTJ"
	list, err := tm.GetAssetsAccountList(testApp, walletID, 0, 10000000)
	if err != nil {
		log.Error("unexpected error:", err)
		return
	}
	for i, w := range list {
		log.Infof("account[%d] : %+v", i, w)
	}
	log.Info("account count:", len(list))

	tm.CloseDB(testApp)

}

func TestWalletManager_CreateAddress(t *testing.T) {

	tm := testInitWalletManager()

	walletID := "WKFkmvsSFz5mC1cAX3edJC2e6hH6ow3X9E"
	//accountID := "4gbLYX9shEoABGKaZrbTmAfHRXPKkVK6wudFEp7miNFZ7F9ZCL6t38Nr6tSr8GDS11tNZn7iwghsbt2qs6P1bkje"
	accountID := "3D58HdM35ZJrMgAzzRduGy6mPVqc8yeGNFm5kNBU16tZYkf84N9C4uppHtJWfw6bEXMtkFgXTxPnw3kN9m7QhiX2"
	address, err := tm.CreateAddress(testApp, walletID, accountID, 300)
	if err != nil {
		log.Error(err)
		return
	}

	for _, addr := range address {
		log.Info(addr.Address)
	}


	tm.CloseDB(testApp)
}

func TestWalletManager_GetAddressList(t *testing.T) {

	tm := testInitWalletManager()

	walletID := "W3TuDqe8VShgyPcg2dw4FRrNQbmxxiGPTJ"
	accountID := "3D58HdM35ZJrMgAzzRduGy6mPVqc8yeGNFm5kNBU16tZYkf84N9C4uppHtJWfw6bEXMtkFgXTxPnw3kN9m7QhiX2"
	list, err := tm.GetAddressList(testApp, walletID, accountID, 0, -1, false)
	if err != nil {
		log.Error("unexpected error:", err)
		return
	}
	for _, w := range list {
		fmt.Println(w.Address)
		//log.Info("account[", i, "] :", w.AccountID)
		//log.Info("address[", i, "] :", w.PublicKey)
	}
	log.Info("address count:", len(list))

	tm.CloseDB(testApp)
}


