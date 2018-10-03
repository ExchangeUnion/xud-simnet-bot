package main

import (
	"fmt"
	"os"

	"github.com/ExchangeUnion/xud-tests/trading"
)

func main() {
	if err := initConfig(); err != nil {
		printError("Could not initialize config:", err)
	}

	if err := initLogger(cfg.LogPath); err != nil {
		printError("Could not initialize logger:", err)
	}

	xud := cfg.Xud

	err := xud.Connect()
	info, err := xud.GetInfo()

	if err != nil {
		log.Error("Could not connect to XUD: %v", err)
		os.Exit(1)
	}

	log.Info("Conntected to XUD node %v version %v", info.NodePubKey, info.Version)

	trading.InitTradingBot(xud)

	select {}
}

func printError(messages ...interface{}) {
	fmt.Println(messages...)
	os.Exit(1)
}
