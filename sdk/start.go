package sdk

import (
	"time"

	"github.com/anyswap/CrossChain-Router/v3/cmd/utils"
	"github.com/anyswap/CrossChain-Router/v3/common"
	"github.com/anyswap/CrossChain-Router/v3/log"
	"github.com/anyswap/CrossChain-Router/v3/router"
	"github.com/anyswap/CrossChain-Router/v3/tools"
	"github.com/anyswap/RouterSDK-injective/config"
)

var (
	// BridgeInstance b instance
	BridgeInstance *Bridge

	// BridgeInited provide full rpc when inited
	BridgeInited bool

	adjustInterval = 60 // seconds

	initedRouterContracts = make(map[string]bool)
)

// StartEndpoint start endpoint
func StartEndpoint() {
	b := NewCrossChainBridge()

	cfg := config.GetServerConfig()
	b.SetGatewayConfig(cfg.GatewayConfig)
	b.AdjustGatewayOrder()
	b.InitAfterConfig()

	chainID := cfg.ChainID
	latestBlock, err := b.GetLatestBlockNumber()
	if err != nil {
		log.Warn("get lastest block number failed", "chainID", chainID, "err", err)
	} else {
		log.Infof("[%5v] lastest block number is %v", chainID, latestBlock)
	}

	BridgeInstance = b
	go b.adjustGateway()
}

// InitAfterLoad init after load
func InitAfterLoad() {
	b := BridgeInstance

	cfg := config.GetServerConfig()
	chainID := cfg.ChainID
	wbr := router.GetBridgeByChainID(chainID)
	if wbr == nil {
		log.Fatal("bridge is not init", "chainID", chainID)
	}

	chainCfg := wbr.GetChainConfig()
	b.SetChainConfig(chainCfg)
	log.Info("init chain config success", "chainID", chainID, "chainCfg", common.ToJSONString(chainCfg, false))

	routerContract := chainCfg.RouterContract
	if routerContract != "" {
		if err := b.InitRouterInfo(routerContract, chainCfg.RouterVersion); err == nil {
			initedRouterContracts[routerContract] = true
		}
	}

	for _, tokenID := range router.AllTokenIDs {
		tokenAddr := router.GetCachedMultichainToken(tokenID, chainID)
		if tokenAddr == "" {
			continue
		}
		tokenCfg := wbr.GetTokenConfig(tokenAddr)
		b.SetTokenConfig(tokenAddr, tokenCfg)
		log.Info("init token config success", "chainID", chainID, "tokenCfg", common.ToJSONString(tokenCfg, false))
		routerContract = tokenCfg.RouterContract
		if routerContract != "" && !initedRouterContracts[routerContract] {
			if err := b.InitRouterInfo(routerContract, tokenCfg.RouterVersion); err == nil {
				initedRouterContracts[routerContract] = true
			}
		}
	}

	BridgeInited = true
	log.Info("init after load finished", "chainID", chainID, "chainName", chainCfg.BlockChain)
}

// AdjustGatewayOrder adjust gateway order once
func (b *Bridge) AdjustGatewayOrder() {
	chainID := config.GetServerConfig().ChainID
	// use block number as weight
	var weightedAPIs tools.WeightedStringSlice
	gateway := b.GetGatewayConfig()
	if gateway == nil {
		return
	}
	var maxHeight uint64
	length := len(gateway.APIAddress)
	for i := length; i > 0; i-- { // query in reverse order
		if utils.IsCleanuping() {
			return
		}
		apiAddress := gateway.APIAddress[i-1]
		height, _ := b.GetLatestBlockNumberOf(apiAddress)
		weightedAPIs = weightedAPIs.Add(apiAddress, height)
		if height > maxHeight {
			maxHeight = height
		}
	}
	if length == 0 { // update for bridges only use grpc apis
		maxHeight, _ = b.GetLatestBlockNumber()
	}
	if maxHeight > 0 {
		log.Info("update latest block number", "chainID", chainID, "height", maxHeight)
	}
	if len(weightedAPIs) > 0 {
		weightedAPIs.Reverse() // reverse as iter in reverse order in the above
		weightedAPIs = weightedAPIs.Sort()
		gateway.APIAddress = weightedAPIs.GetStrings()
		gateway.WeightedAPIs = weightedAPIs
	}
}

func (b *Bridge) adjustGateway() {
	for adjustCount := 0; ; adjustCount++ {
		for i := 0; i < adjustInterval; i++ {
			if utils.IsCleanuping() {
				return
			}
			time.Sleep(1 * time.Second)
		}

		b.AdjustGatewayOrder()

		if adjustCount%3 == 0 && b.GetGatewayConfig().WeightedAPIs.Len() > 0 {
			log.Info("adjust gateways", "result", b.GetGatewayConfig().WeightedAPIs)
		}
	}
}
