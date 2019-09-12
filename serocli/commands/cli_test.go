package commands

import (
	"github.com/blocktree/go-openw-cli/openwcli"
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/owtp"
	"path/filepath"
	"testing"
)


func init() {
	owtp.Debug = false
}

func getTestOpenwCLI() *openwcli.CLI {

	confFile := filepath.Join("conf", "test.ini")
	//confFile := filepath.Join("conf", "prod.ini")

	config, err := openwcli.LoadConfig(confFile)
	if err != nil {
		log.Error("unexpected error: ", err)
		return nil
	}

	//加载SERO配置
	err = LoadSEROConfig()
	if err != nil {
		log.Error("unexpected error: ", err)
		return nil
	}

	cli, err := openwcli.NewCLI(config)
	if err != nil {
		log.Error("unexpected error: ", err)
		return nil
	}
	err = cli.SetSignRawTransactionFunc(SERO_SignRawTransaction)
	if err != nil {
		log.Error("unexpected error: ", err)
		return nil
	}
	globalCLI = cli

	return cli

}


func TestCLI_CreateWalletOnServer(t *testing.T) {
	cli := getTestOpenwCLI()
	if cli == nil {
		return
	}
	_, err := cli.CreateWalletOnServer("test-sero-wallet", "12345678")
	if err != nil {
		log.Error("CreateWalletOnServer error:", err)
		return
	}
}

func TestCLI_GetWalletsOnServer(t *testing.T) {
	cli := getTestOpenwCLI()
	if cli == nil {
		return
	}
	wallets, err := cli.GetWalletsOnServer()
	if err != nil {
		log.Error("GetWalletsOnServer error:", err)
		return
	}
	for i, w := range wallets {
		log.Info("wallet[", i, "]:", w)
	}
}

func TestCLI_CreateAccountOnServer(t *testing.T) {
	cli := getTestOpenwCLI()
	if cli == nil {
		return
	}
	//walletID := "WFXjcUJXsDB3uBxa8fbr8xqThUeVNtzMbs"
	walletID := "WFXtudgu9Q5ktpcfDPC8gVEbHF1t1QWiVV"
	wallet, err := cli.GetWalletByWalletID(walletID)
	if err != nil {
		log.Error("GetWalletByWalletID error:", err)
		return
	}

	if wallet != nil {
		_, _, err = SERO_CreateAccountOnServer(cli, "mainnetSERO_1", "12345678", "SERO", wallet)
		if err != nil {
			log.Error("CreateAccountOnServer error:", err)
			return
		}
	}
}

func TestCLI_GetAccountsOnServer(t *testing.T) {
	cli := getTestOpenwCLI()
	if cli == nil {
		return
	}

	//walletID := "WFXjcUJXsDB3uBxa8fbr8xqThUeVNtzMbs"
	walletID := "WFXtudgu9Q5ktpcfDPC8gVEbHF1t1QWiVV"
	accounts, err := cli.GetAccountsOnServer(walletID)
	if err != nil {
		log.Error("GetAccountsOnServer error:", err)
		return
	}
	for i, w := range accounts {
		log.Info("account[", i, "]:", w)
	}
}