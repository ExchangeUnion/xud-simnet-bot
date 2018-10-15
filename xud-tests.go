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

	if !cfg.DisableTrading {
		log.Info("Starting trading bot")

		xud := cfg.Xud

		err := xud.Init()

		if err == nil {
			trading.InitTradingBot(&wg, xud)
		} else {
			printErrorAndExit("Could not read required files for XUD:", err)
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

	wg.Wait()
	log.Info("All services died")
}

func initChannelManager(lnd *lndclient.Lnd, isBtc bool) {
	nodeName := "lndbtc"

	if !isBtc {
		nodeName = "lndltc"
	}

	log.Info("Starting channel manager for %v", nodeName)

	err := lnd.Init()

	if err == nil {
		channels.InitChannelManager(&wg, lnd, nodeName)
	} else {
		printErrorAndExit("Could not read required files for", nodeName, ":", err)
	}
}

func printErrorAndExit(messages ...interface{}) {
	fmt.Println(messages...)
	os.Exit(1)
}
