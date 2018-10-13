package main

import (
	"fmt"
	"os"

	"github.com/ExchangeUnion/xud-tests/channels"
	"github.com/ExchangeUnion/xud-tests/lndclient"
	"github.com/ExchangeUnion/xud-tests/trading"
)

func main() {
	if err := initConfig(); err != nil {
		printError("Could not initialize config:", err)
	}

	if err := initLogger(cfg.LogPath); err != nil {
		printError("Could not initialize logger:", err)
	}

	if !cfg.DisableTrading {
		log.Info("Starting trading bot")

		xud := cfg.Xud

		err := xud.Connect()
		info, err := xud.GetInfo()

		if err == nil {
			log.Info("Conntected to XUD node %v version %v", info.NodePubKey, info.Version)
			trading.InitTradingBot(xud)
		} else {
			log.Error("Could not connect to XUD: %v", err)
		}
	}

	if !cfg.DisableChannelManager {
		if !xudCfg.LndBtc.Disable {
			initChannelManager(xudCfg.LndBtc, true)
		}

		if !xudCfg.LndLtc.Disable {
			initChannelManager(xudCfg.LndBtc, false)
		}
	}

	select {}
}

func initChannelManager(lnd *lndclient.Lnd, isBtc bool) {
	nodeName := "lndbtc"

	if !isBtc {
		nodeName = "lndltc"
	}

	log.Info("Starting channel manager for %v", nodeName)

	err := lnd.Connect()
	info, err := lnd.GetInfo()

	if err == nil {
		log.Info("Connected to %v node %v version %v", nodeName, info.IdentityPubkey, info.Version)
		channels.InitChannelManager(lnd, nodeName)
	} else {
		log.Error("Could not connect to %v: %v", nodeName, err)
	}
}

func printError(messages ...interface{}) {
	fmt.Println(messages...)
	os.Exit(1)
}
