module github.com/blocktree/sero-adapter

go 1.12

require (
	github.com/asdine/storm v2.1.2+incompatible
	github.com/astaxie/beego v1.11.1
	github.com/blocktree/go-openw-sdk v1.3.4
	github.com/blocktree/go-owcdrivers v1.1.11 // indirect
	github.com/blocktree/go-owcrypt v1.0.3
	github.com/blocktree/openwallet v1.4.10
	github.com/bndr/gotabulate v1.1.2
	github.com/google/uuid v1.1.1
	github.com/imroc/req v0.2.3
	github.com/mr-tron/base58 v1.1.1
	github.com/sero-cash/go-sero v0.0.0-20190905034124-a9a295a8f2ca
	github.com/shopspring/decimal v0.0.0-20180709203117-cd690d0c9e24
	github.com/tidwall/gjson v1.2.1
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	go.etcd.io/bbolt v1.3.2
)

replace github.com/blocktree/openwallet => ../../openwallet
