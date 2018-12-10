package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/ExchangeUnion/xud-tests/channels"
	"github.com/ExchangeUnion/xud-tests/lndclient"
	"github.com/ExchangeUnion/xud-tests/trading"
)

var wg sync.WaitGroup

func main() {
	if err := initConfig(); err != nil {
		printErrorAndExit("Could not initialize config:", err)
	}

	if err := initLogger(cfg.LogPath); err != nil {
		printErrorAndExit("Could not initialize logger:", err)
	}

	cfg.Slack.InitSlack()

	if !cfg.DisableTrading {
		log.Infof("Starting trading bot with %v mode", cfg.TradingMode)

		xud := cfg.Xud

		err := xud.Init()

		if err == nil {
			trading.InitTradingBot(&wg, xud, cfg.TradingMode)
		} else {
			printErrorAndExit("Could not read required files for XUD:", err)
		}
	}

	if !cfg.DisableChannelManager {
		if !xudCfg.LndBtc.Disable {
			initChannelManager(xudCfg.LndBtc, true)
		}

		if !xudCfg.LndLtc.Disable {
			initChannelManager(xudCfg.LndLtc, false)
		}
	}

	wg.Wait()
	log.Warning("All services died")
}

func initChannelManager(lnd *lndclient.Lnd, isBtc bool) {
	nodeName := "lndbtc"

	if !isBtc {
		nodeName = "lndltc"
	}

	log.Info("Starting channel manager for %v", nodeName)

	err := lnd.Init()

	if err == nil {
		channels.InitChannelManager(&wg, lnd, cfg.Slack, cfg.DataDir, nodeName)
	} else {
		printErrorAndExit("Could not read required files for", nodeName, ":", err)
	}
}

func printErrorAndExit(messages ...interface{}) {
	fmt.Println(messages...)
	os.Exit(1)
}
