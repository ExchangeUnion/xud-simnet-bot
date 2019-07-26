package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/ExchangeUnion/xud-tests/lndchannels"
	"github.com/ExchangeUnion/xud-tests/lndclient"
	"github.com/ExchangeUnion/xud-tests/raidenchannels"
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

	cfg.Discord.Init()
	err := cfg.Xud.Init()

	if !cfg.DisableTrading {
		log.Infof("Starting trading bot with %v mode", cfg.TradingMode)

		if err == nil {
			trading.InitTradingBot(&wg, cfg.Xud, cfg.TradingMode)
		} else {
			printErrorAndExit("Could not read required files for XUD:", err)
		}
	}

	if !cfg.DisableChannelManager {
		if !xudCfg.LndConfigs.Btc.Disable {
			initChannelManager(xudCfg.LndConfigs.Btc, true)
		}

		if !xudCfg.LndConfigs.Ltc.Disable {
			initChannelManager(xudCfg.LndConfigs.Ltc, false)
		}

		if !cfg.Ethereum.Disable {
			if !xudCfg.Raiden.Disable {
				log.Info("Starting Raiden channel manager")

				err := cfg.Ethereum.Init()

				if err != nil {
					log.Error("Could not start Ethereum client: %v", err)
					return
				}

				xudCfg.Raiden.Init()

				raidenchannels.InitChannelManager(
					&wg,
					cfg.Xud,
					xudCfg.Raiden,
					cfg.Ethereum,
					cfg.Discord,
					cfg.DataDir,
					cfg.EnableBalancing,
				)
			}
		}
	}

	cfg.Discord.SendMessage("Started bot")

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
		lndchannels.InitChannelManager(&wg, lnd, cfg.Discord, cfg.DataDir, nodeName)
	} else {
		printErrorAndExit("Could not read required files for", nodeName, ":", err)
	}
}

func printErrorAndExit(messages ...interface{}) {
	fmt.Println(messages...)
	os.Exit(1)
}
