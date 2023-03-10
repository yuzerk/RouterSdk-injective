# RouterSDK-injective

## resources

1) document

<https://docs.injective.network/>

2) apiUrl

testnet (tendermint rpc): <https://testnet.tm.injective.network:443>

3) explorer:

<https://testnet.explorer.injective.network/>

4) faucet:

<https://inj.supply/>

## tools

1) getChainId

```shell
go run ./tools/getStubChainID/main.go -p testnet
```

```text
mainnet: 1019511453253
testnet: 1019511453254
```

2) publicKeyToAddress

```shell
go run ./tools/publicKeyToAddress/main.go -p 0x0468438a94627b0de2b6a7c9af99136ef7e607f7944b749c3534bb27a89e742d583b1c8b3aecfae45dea2ac58730aa6ba654c73c435d44755e5cd1500c8f4d036b -prefix inj
```

addr: inj10yyn2er9k5cs9qn55l7t23yxxk7egecpyvg2hh

3) createDenom

4) mintToken

5) sendToken

## router config setting

1) chainConfig

routerContract: mpc address
extra: format is `inj:inj`

2) tokenConfig

for meta coin,

```text
tokenAddress: inj
decimals: 6
```

for other tokens,

```text
tokenAddress: factory/{creator}/{subdenom}
decimals: 6 (maybe other value)
```

## sdk rpc test

1) start chain support program

```shell
./build/bin/injective-chain-support -c config.toml
```

2) test rpc calling

the rpc api is still under developing,
please reference the source code file `server/rpcapi.go` for the api methods
and the arguments and response type of each api method.

for example,

- call `GetLatestBlockNumber`

```shell
curl -sS -X POST -H "Content-Type:application/json" --data '{"jsonrpc":"2.0", "method":"bridge.GetLatestBlockNumber", "params":[], "id":1}' http://127.0.0.1:12556
```

```json
{
  "jsonrpc": "2.0",
  "result": 9038879,
  "id": 1
}
```

- call `GetTransactionStatus`

```shell
curl -sS -X POST -H "Content-Type:application/json" --data '{"jsonrpc":"2.0", "method":"bridge.GetTransactionStatus", "params":["768F66E059D1D2BFF5FD9F4A440DA8F32FAC4D64B2F45A32D66EEC69080DB103"], "id":1}' http://127.0.0.1:12556
```

```json
{
  "jsonrpc": "2.0",
  "result": {
    "confirmations": 999,
    "block_height": 9037827
  },
  "id": 1
}
```
