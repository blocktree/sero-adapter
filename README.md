# sero-adapter

sero-adapter适配了openwallet.AssetsAdapter接口，给应用提供了底层的区块链协议支持。


## 使用方法

1. 下载最新版的openw-sero和SERO.ini

https://github.com/blocktree/sero-adapter/releases

openw-sero与SERO.ini放在同一个目录中。

2. 下载gero

https://github.com/sero-cash/go-sero/releases/tag/v1.0.0-rc11

3. 本地启动gero

打开终端，cd到geropkg目录下，执行命令：

```shell

//导入c++库目录到环境变量
$ export DYLD_LIBRARY_PATH="./czero/lib/"
$ export LD_LIBRARY_PATH="./czero/lib/"

//启动本地gero节点，这个节点不同步区块数据，只是提供密码算法
$ bin/gero --mineMode --datadir ~/geroData --nodiscover --rpc --rpcport 8545 --rpcapi local,sero --rpcaddr 127.0.0.1 --rpccorsdomain "*" --exchangeValueStr

//后台运行gero
$ nohup bin/gero --mineMode --datadir ~/geroData --nodiscover --rpc --rpcport 8545 --rpcapi local,sero --rpcaddr 127.0.0.1 --rpccorsdomain "*" --exchangeValueStr >> gero.log 2>&1 &

```

4. 执行openw-sero创建钱包

openw-sero是openw-cli的分支，操作命令完全一致。