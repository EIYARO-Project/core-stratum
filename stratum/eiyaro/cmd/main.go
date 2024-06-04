package main

import (
	"context"
	ey "corepool/stratum/eiyaro"
	"math/big"
	"strconv"
	"time"

	pb "corepool/common/format/generated"
	"corepool/common/logger"
	"corepool/common/rpc/hostprovider"
	"corepool/common/rpc/http"
	"corepool/common/service"
	"corepool/common/vars"
	ss "corepool/stratum"
)

func main() {
	vars.Init()

	stratumId := vars.GetInt("stratum.id", 0)
	service := service.New("stratum_eiyaro"+"."+strconv.Itoa(stratumId), service.NewConfig(vars.GetString("mode", "")))

	maxConn := vars.GetInt("stratum.max_conn", 32768)
	// init connection controller
	connCtl := ss.NewConnCtl(
		vars.GetDuration("stratum.default_ban_period", 20*time.Minute),
		pb.CoinType_EY,
		vars.GetBool("ip.ban_enable", false),
		vars.GetInt("ip.max_throughput", 131072),
		vars.GetInt("ip.max_connection", 1000),
		vars.GetFloat64("ip.throughput_ratio", 1.2),
		vars.GetFloat64("ip.connection_ratio", 1.2),
		vars.GetStringSlice("ip.white_list", []string{}))
	// init server global state
	state, err := ss.InitServerState(context.Background(), connCtl, stratumId, uint(maxConn))
	if err != nil {
		logger.Error("can't create server state")
		return
	}

	// configuration node & verifier
	node := vars.GetString("node.name", "btmc_testnet")
	nodeUrl := vars.GetString("node.url", "http://127.0.0.1:9888")
	hostprovider.InitStaticProvider(map[string][]string{node: {nodeUrl}})
	http.Init(time.Second)

	syncer, err := ey.NewBtmcNodeSyncer(node, nodeUrl)
	if err != nil {
		logger.Error("can't create node syncer", "error", err)
		return
	}

	verifier, err := ey.NewBtmcVerifier(state)
	if err != nil {
		logger.Error("can't create verifier", "error", err)
		return
	}

	// create btmSessionData obj
	dataBuilder := ey.NewBtmcSessionDataBuilder(uint64(state.GetId()), maxConn)

	// create diffAdjust
	diffAdjust := ss.NewDiffAdjust(big.NewInt(vars.GetInt64("session.diff", 500000)))

	// start server
	if err := ss.NewServer(
		vars.GetInt("stratum.port", 8118),
		maxConn,
		state,
		syncer,
		vars.GetDuration("node.sync_interval", 100*time.Millisecond), // sync interval
		verifier,
		vars.GetDuration("session.timeout", 5*time.Minute),
		vars.GetDuration("session.sched_interval", 0),
		dataBuilder,
		diffAdjust,
		ey.NewBtmDecoder(),
	); err != nil {
		logger.Error("can't create server", "error", err)
		return
	}

	service.Run(":" + vars.GetString("service.port", "8082"))
}
