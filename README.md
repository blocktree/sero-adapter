# sero-adapter

sero-adapter适配了openwallet.AssetsAdapter接口，给应用提供了底层的区块链协议支持。

## 部署gero local节点

### linux

1. 安装gero

```shell

# 环境准备：依次运行如下命令
$ apt update
$ apt install wget curl rsync libgmp-dev lrzsz -y

# 创建目录
$ mkdir -p /data/sero/geroData

$ cd /data/sero

# 下载二进制文件
$ wget https://sero-media-1256272584.cos.ap-shanghai.myqcloud.com/gero/v1.0.0-rc11/gero-v1.0.0-rc11-linux-amd64-v4.tar.gz

# 添加环境变量
$ vim ~/.bashrc

# 之后按字母i后将光标放到最后一行，粘贴如下内容
$ export DYLD_LIBRARY_PATH="/data/sero/geropkg/czero/lib/"
$ export LD_LIBRARY_PATH="/data/sero/geropkg/czero/lib/"

# 执行命令, 让其生效
$ source ~/.bashrc

# 检验生效的效果如下，执行命令
$ echo $LD_LIBRARY_PATH
# 输出如下内容： /data/sero/geropkg/czero/lib/

#  部署依赖库libgomp.so.1.0.0
# 首先检查执行命令
$ ls -l /usr/lib/x86_64-linux-gnu/libgomp.so.1  

# 如果现实如下信息，则无需再部署此依赖库
lrwxrwxrwx 1 root root 16 Feb  6  2018 /usr/lib/x86_64-linux-gnu/libgomp.so.1 -> libgomp.so.1.0.0

# 如果没显示上面的信息，则执行如下操作：
$ wget https://github.com/blocktree/sero-adapter/releases/download/v1.0.12/libgomp.so.1.0.0.zip
$ unzip libgomp.so.1.0.0.zip
$ rsync -avz /data/sero/libgomp.so.1.0.0  /usr/lib/x86_64-linux-gnu/
$ ln -s /usr/lib/x86_64-linux-gnu/libgomp.so.1.0.0  /usr/lib/x86_64-linux-gnu/libgomp.so.1

# 本地启动gero
nohup /data/sero/geropkg/bin/gero --mineMode --datadir "/data/sero/geroData" --nodiscover --rpc --rpcport 8545 --rpcapi "local,sero" --rpcaddr 127.0.0.1 --rpccorsdomain "*" --exchangeValueStr >> /data/sero/geroData/run.log 2>&1 &

```

2. 本地启动gero

```shell

# 本地启动gero
nohup /data/sero/geropkg/bin/gero --mineMode --datadir "/data/sero/geroData" --nodiscover --rpc --rpcport 8545 --rpcapi "local,sero" --rpcaddr 127.0.0.1 --rpccorsdomain "*" --exchangeValueStr >> /data/sero/geroData/run.log 2>&1 &

```

### mac os

1. 下载gero

https://sero-media-1256272584.cos.ap-shanghai.myqcloud.com/gero/v1.0.0-rc11/gero-v1.0.0-rc11-darwin-amd64.tar.gz

2. 本地启动gero

打开终端，cd到geropkg目录下，执行命令：

```shell

# 导入c++库目录到环境变量
$ export DYLD_LIBRARY_PATH="./czero/lib/"
$ export LD_LIBRARY_PATH="./czero/lib/"

# 启动本地gero节点，这个节点不同步区块数据，只是提供密码算法
$ bin/gero --mineMode --datadir ~/geroData --nodiscover --rpc --rpcport 8545 --rpcapi local,sero --rpcaddr 127.0.0.1 --rpccorsdomain "*" --exchangeValueStr

```

## 使用方法

1. 下载最新版的openw-sero

https://github.com/blocktree/sero-adapter/releases

2. 创建SERO.ini配置文件，与openw-sero放在同一个目录中。

SERO.ini内容：

```ini

# Remote Server RPC api url
serverAPI = "http://127.0.0.1:8545"
# fix gas for transaction
fixGas = 25000
# Cache data file directory, default = "", current directory: ./data
dataDir = ""

```

3. openw-sero是openw-cli的分支，操作命令完全一致。

4. 注意事项

openw-sero支持SERO主链币和代币的转账和汇总。
由于代币汇总需要SERO手续费支持，所以我们可以把创建SERO账户时的第一个默认地址作为手续费地址。
每次汇总账户的代币时候，程序就会取一个足够做SERO手续费的utxo组建交易，多出的SERO找零到汇总地址。
如果SERO不足够发起代币汇总，你就可以转账SERO给账户的默认地址，即手续费地址。