package server

import (
	"time"

	"github.com/anyswap/CrossChain-Router/v3/common"
	"github.com/anyswap/CrossChain-Router/v3/log"
)

// StatInfo statistics info
type StatInfo struct {
	SuccCount  uint64
	TotalCount uint64
}

var (
	// key is session token, start date timestamp
	stats = make(map[string]map[uint64]*StatInfo)
	// minimum interval to print rpc call stats
	latestPrintStatsTimestamp uint64
)

const (
	secondsPerDay      = 86400
	printStatsInterval = 28800
)

func getStatMap(token string) map[uint64]*StatInfo {
	return stats[token]
}

func getOrCreateStatMap(token string) map[uint64]*StatInfo {
	statMap, ok := stats[token]
	if !ok {
		statMap = make(map[uint64]*StatInfo)
		stats[token] = statMap
	}
	return statMap
}

func getDateStartTimestamp(timestamp uint64) uint64 {
	return timestamp - (timestamp % secondsPerDay)
}

func getOrCreateCurrStat(token string) *StatInfo {
	statMap := getOrCreateStatMap(token)
	currTime := getDateStartTimestamp(uint64(time.Now().Unix()))
	stat, ok := statMap[currTime]
	if !ok {
		stat = &StatInfo{}
		statMap[currTime] = stat
	}
	return stat
}

func statTotalCalls(token string) {
	stat := getOrCreateCurrStat(token)
	stat.TotalCount++
	log.Debug("update rpc call stats", "token", token, "total", stat.TotalCount)

	now := uint64(common.Now())
	if now-latestPrintStatsTimestamp > printStatsInterval {
		log.Infof("print rpc call stats. %v", common.ToJSONString(stat, false))
		latestPrintStatsTimestamp = now
	}
}

func statSucessCalls(token string) {
	stat := getOrCreateCurrStat(token)
	stat.SuccCount++
	log.Debug("update rpc call stats", "token", token, "succ", stat.SuccCount)
}
