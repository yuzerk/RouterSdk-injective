package server

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"

	"github.com/anyswap/CrossChain-Router/v3/log"
	"github.com/anyswap/CrossChain-Router/v3/tokens"
	wrapper "github.com/anyswap/CrossChain-Router/v3/tokens/wrapper/impl"
	routersdk "github.com/anyswap/RouterSDK-injective/sdk"
)

var (
	errBridgeNotInited   = errors.New("bridge not inited")
	errWrongNumberOfArgs = errors.New("wrong number of args")
	errWrongArgs         = errors.New("wrong args")
)

// ChainSupportAPI rpc api handler
type ChainSupportAPI struct{}

// RPCNullArgs null args
type RPCNullArgs struct{}

func convertToArgument(dst interface{}, src interface{}) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &dst)
	if err != nil {
		log.Warn("convert argument failed", "src", src, "dst", dst, "err", err)
	}
	return err
}

// GetServerInfo api
func (s *ChainSupportAPI) GetServerInfo(r *http.Request, args *RPCNullArgs, result *GetServerInfoResult) error {
	*result = *getServerInfo()
	return nil
}

// GetVersionInfo api
func (s *ChainSupportAPI) GetVersionInfo(r *http.Request, args *RPCNullArgs, result *string) error {
	*result = getVersionInfo()
	return nil
}

// GetStatInfo api
func (s *ChainSupportAPI) GetStatInfo(r *http.Request, sessToken *string, result *map[uint64]*StatInfo) error {
	*result = getStatMap(*sessToken)
	return nil
}

// RegisterSwap register swap.
// used in `RegisterRouterSwap` server rpc.
func (b *ChainSupportAPI) RegisterSwap(r *http.Request, args *[]interface{}, result *wrapper.RegisterSwapResult) error {
	if !routersdk.BridgeInited {
		return errBridgeNotInited
	}
	if len(*args) != 2 {
		return errWrongNumberOfArgs
	}
	txhash, ok := (*args)[0].(string)
	if !ok {
		return errWrongArgs
	}
	var registerArgs tokens.RegisterArgs
	err := convertToArgument(&registerArgs, (*args)[1])
	if err != nil {
		return err
	}
	txinfos, errs := routersdk.BridgeInstance.RegisterSwap(txhash, &registerArgs)
	*result = wrapper.RegisterSwapResult{
		SwapTxInfos: txinfos,
		Errs:        errs,
	}
	return nil
}

// VerifyTransaction verify swap tx is valid and success on chain with needed confirmations.
func (b *ChainSupportAPI) VerifyTransaction(r *http.Request, args *[]interface{}, result *tokens.SwapTxInfo) error {
	if !routersdk.BridgeInited {
		return errBridgeNotInited
	}
	if len(*args) != 2 {
		return errWrongNumberOfArgs
	}
	txhash, ok := (*args)[0].(string)
	if !ok {
		return errWrongArgs
	}
	var verifyArgs tokens.VerifyArgs
	err := convertToArgument(&verifyArgs, (*args)[1])
	if err != nil {
		return err
	}
	txinfo, err := routersdk.BridgeInstance.VerifyTransaction(txhash, &verifyArgs)
	if err != nil {
		return err
	}
	*result = *txinfo
	return nil
}

// BuildRawTransaction build tx with specified args.
func (b *ChainSupportAPI) BuildRawTransaction(r *http.Request, args *[]interface{}, result *interface{}) error {
	if !routersdk.BridgeInited {
		return errBridgeNotInited
	}
	if len(*args) != 1 {
		return errWrongNumberOfArgs
	}
	var buildArgs tokens.BuildTxArgs
	err := convertToArgument(&buildArgs, (*args)[0])
	if err != nil {
		return err
	}
	rawTx, err := routersdk.BridgeInstance.BuildRawTransaction(&buildArgs)
	if err != nil {
		return err
	}
	*result = rawTx
	return nil
}

func (b *ChainSupportAPI) VerifyMsgHash(r *http.Request, args *[]interface{}, result *bool) error {
	if !routersdk.BridgeInited {
		return errBridgeNotInited
	}
	if len(*args) != 2 {
		return errWrongNumberOfArgs
	}
	var rawTx routersdk.BuildRawTx
	err := convertToArgument(&rawTx, (*args)[0])
	if err != nil {
		return err
	}
	var msgHash []string
	err = convertToArgument(&msgHash, (*args)[1])
	if err != nil {
		return err
	}
	err = routersdk.BridgeInstance.VerifyMsgHash(&rawTx, msgHash)
	if err != nil {
		return err
	}
	*result = true
	return nil
}

type SignTxResult struct {
	SignedTx interface{} `json:"signedTx"`
	TxHash   string      `json:"txhash"`
}

// MPCSignTransaction mpc sign tx.
func (b *ChainSupportAPI) MPCSignTransaction(r *http.Request, args *[]interface{}, result *SignTxResult) error {
	if !routersdk.BridgeInited {
		return errBridgeNotInited
	}
	if len(*args) != 2 {
		return errWrongNumberOfArgs
	}
	var rawTx routersdk.BuildRawTx
	err := convertToArgument(&rawTx, (*args)[0])
	if err != nil {
		return err
	}
	var buildArgs tokens.BuildTxArgs
	err = convertToArgument(&buildArgs, (*args)[1])
	if err != nil {
		return err
	}
	signedTx, txHash, err := routersdk.BridgeInstance.MPCSignTransaction(&rawTx, &buildArgs)
	if err != nil {
		return err
	}
	*result = SignTxResult{
		SignedTx: signedTx,
		TxHash:   txHash,
	}
	return nil
}

// SendTransaction send signed raw tx.
func (b *ChainSupportAPI) SendTransaction(r *http.Request, args *string, result *string) error {
	if !routersdk.BridgeInited {
		return errBridgeNotInited
	}
	encodeTx := *args
	txBytes, err := base64.StdEncoding.DecodeString(encodeTx)
	if err != nil {
		return err
	}
	txhash, err := routersdk.BridgeInstance.SendTransaction(txBytes)
	if err != nil {
		return err
	}
	*result = txhash
	return nil
}

// GetTransaction get tx by hash.
func (b *ChainSupportAPI) GetTransaction(r *http.Request, args *[]string, result *interface{}) error {
	if len(*args) != 1 {
		return errWrongNumberOfArgs
	}
	txhash := (*args)[0]
	tx, err := routersdk.BridgeInstance.GetTransaction(txhash)
	if err != nil {
		return err
	}
	*result = tx
	return nil
}

// GetTransactionStatus get tx status by hash.
// get tx related infos like block height, confirmations, receipts etc.
// These infos is used to verify tx is acceptable.
// you can extend `TxStatus` if fields in it is not enough to do the checking.
func (b *ChainSupportAPI) GetTransactionStatus(r *http.Request, args *[]string, result *tokens.TxStatus) error {
	if len(*args) != 1 {
		return errWrongNumberOfArgs
	}
	txhash := (*args)[0]
	txStatus, err := routersdk.BridgeInstance.GetTransactionStatus(txhash)
	if err != nil {
		return err
	}
	*result = *txStatus
	return nil
}

// GetLatestBlockNumber get latest block number through gateway urls.
// used in `GetRouterSwap` server rpc.
func (b *ChainSupportAPI) GetLatestBlockNumber(r *http.Request, args *RPCNullArgs, result *uint64) error {
	blockNumber, err := routersdk.BridgeInstance.GetLatestBlockNumber()
	if err != nil {
		return err
	}
	*result = blockNumber
	return nil
}

// GetBalance get balance is used for checking budgets to prevent DOS attacking
func (b *ChainSupportAPI) GetBalance(r *http.Request, args *[]string, result *big.Int) error {
	if len(*args) != 1 {
		return errWrongNumberOfArgs
	}
	address := (*args)[0]
	denom := routersdk.BridgeInstance.Denom
	balance, err := routersdk.BridgeInstance.GetDenomBalance(address, denom)
	if err != nil {
		return err
	}
	result.Set(balance.BigInt())
	return nil
}

// IsValidAddress check if given `address` is valid on this chain.
// prevent swap to an invalid `bind` address which will make assets loss.
func (b *ChainSupportAPI) IsValidAddress(r *http.Request, args *[]string, result *bool) error {
	if len(*args) != 1 {
		return errWrongNumberOfArgs
	}
	address := (*args)[0]
	*result = routersdk.BridgeInstance.IsValidAddress(address)
	return nil
}

// PublicKeyToAddress public key to address
func (b *ChainSupportAPI) PublicKeyToAddress(r *http.Request, args *[]string, result *string) error {
	if len(*args) != 1 {
		return errWrongNumberOfArgs
	}
	pk := (*args)[0]
	address, err := routersdk.BridgeInstance.PublicKeyToAddress(pk)
	if err != nil {
		return err
	}
	*result = address
	return nil
}

// GetMPCAddress get mpc address of router contract
func (b *ChainSupportAPI) GetMPCAddress(r *http.Request, args *[]string, result *string) error {
	if len(*args) != 1 {
		return errWrongNumberOfArgs
	}
	*result = (*args)[0]
	return nil
}

// GetPoolNonce get pool nonce
func (b *ChainSupportAPI) GetPoolNonce(r *http.Request, args *[]string, result *uint64) error {
	if len(*args) != 2 {
		return errWrongNumberOfArgs
	}
	address := (*args)[0]
	height := (*args)[1]
	nonce, err := routersdk.BridgeInstance.GetPoolNonce(address, height)
	if err != nil {
		return err
	}
	*result = nonce
	return nil
}
