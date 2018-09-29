package main

import (
	"fmt"
	"os"
)

func main() {
	if err := initConfig(); err != nil {
		printError("Could not initialize config:", err)
	}

	if err := initLogger(cfg.LogPath); err != nil {
		printError("Could not initialize logger:", err)
	}

	xud := cfg.Xud

	if err := xud.Connect(); err != nil {
		printError("Could not connect to XUD:", err)
	}

	info, _ := xud.GetInfo()

	log.Info("Conntected to XUD node %v version %v", info.NodePubKey, info.Version)
}

func printError(messages ...interface{}) {
	fmt.Println(messages...)
	os.Exit(1)
}
