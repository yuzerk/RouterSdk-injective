# RouterSDK-injective

1) chainId

mainnet: 1019511453253
testnet: 1019511453254

2) apiUrl

testnet (tendermint rpc): https://testnet.tm.injective.network:443

3) explorer:

https://testnet.explorer.injective.network/

4) faucet:

https://inj.supply/

## tools

1) getChainId

```shell
go run ./tools/getStubChainID/main.go -p testnet
```

testnet: 1019511453254

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

        tokenAddress: inj
        decimals: 6

for other tokens,

        tokenAddress: factory/{creator}/{subdenom}
        decimals: 6 (maybe other value)
