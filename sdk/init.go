package sdk

import (
	"math/big"
	"strings"

	tokenfactoryTypes "github.com/InjectiveLabs/sdk-go/chain/tokenfactory/types"
	chainTypes "github.com/InjectiveLabs/sdk-go/chain/types"
	"github.com/anyswap/CrossChain-Router/v3/log"
	"github.com/anyswap/CrossChain-Router/v3/tokens"
	cosmosClient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptoTypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	authTx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	mainnetNetWork = "mainnet"
	testnetNetWork = "testnet"
	devnetNetWork  = "devnet"
)

func NewClientContext() cosmosClient.Context {
	amino := codec.NewLegacyAmino()

	interfaceRegistry := codecTypes.NewInterfaceRegistry()
	interfaceRegistry.RegisterImplementations((*cryptoTypes.PubKey)(nil), &secp256k1.PubKey{})

	authtypes.RegisterInterfaces(interfaceRegistry)
	bankTypes.RegisterInterfaces(interfaceRegistry)
	sdktx.RegisterInterfaces(interfaceRegistry)
	tokenfactoryTypes.RegisterInterfaces(interfaceRegistry)
	chainTypes.RegisterInterfaces(interfaceRegistry)

	protoCodec := codec.NewProtoCodec(interfaceRegistry)
	txConfig := authTx.NewTxConfig(protoCodec, authTx.DefaultSignModes)

	return cosmosClient.Context{}.
		WithCodec(protoCodec).
		WithInterfaceRegistry(interfaceRegistry).
		WithTxConfig(txConfig).
		WithLegacyAmino(amino)
}

// GetStubChainID get stub chainID
func GetStubChainID(chainName, network string) *big.Int {
	chainName = strings.ToUpper(chainName)
	stubChainID := new(big.Int).SetBytes([]byte(chainName))
	switch network {
	case mainnetNetWork:
	case testnetNetWork:
		stubChainID.Add(stubChainID, big.NewInt(1))
	case devnetNetWork:
		stubChainID.Add(stubChainID, big.NewInt(2))
	default:
		log.Fatalf("unknown network %v", network)
	}
	stubChainID.Mod(stubChainID, tokens.StubChainIDBase)
	stubChainID.Add(stubChainID, tokens.StubChainIDBase)
	return stubChainID
}
