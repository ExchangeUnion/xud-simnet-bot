package main

import (
	"github.com/ExchangeUnion/xud-simnet-bot/xudrpc"
	"github.com/google/logger"
	"sync"
)

func main() {
	cfg := loadConfig()
	initLogger(cfg.LogFile)
	logConfig(cfg)

	var wg sync.WaitGroup
	wg.Add(2)

	info := initXud(cfg)
	initDiscord(cfg, info)

	go func() {
		cfg.Database.Init()

		cfg.ChannelManager.Init(cfg.Channels, cfg.Xud, cfg.Discord, cfg.Database)
		wg.Done()
	}()

	go func() {
		err := cfg.Ethereum.Init()

		checkError("Ethereum", err, true)

		cfg.Faucet.Start(cfg.Channels, cfg.Ethereum, cfg.Discord)
		wg.Done()
	}()

	wg.Wait()
	logger.Info("Shutting down")
}

func initXud(cfg *config) *xudrpc.GetInfoResponse {
	logger.Info("Initializing XUD client")

	err := cfg.Xud.Init()
	checkError("XUD", err, true)

	info, err := cfg.Xud.GetInfo()
	checkError("XUD", err, true)

	logger.Info("Initialized XUD client: " + stringify(info))

	return info
}

func initDiscord(cfg *config, info *xudrpc.GetInfoResponse) {
	logger.Info("Initializing Discord client")

	err := cfg.Discord.Init()
	checkError("Discord", err, true)

	err = cfg.Discord.SendMessage("Started xud-simnet-bot with XUD node: **" + info.Alias + "** (`" + info.NodePubKey + "`)")
	checkError("Discord", err, true)

	logger.Info("Initialized Discord client")
}

func checkError(service string, err error, fatal bool) {
	if err != nil {
		message := "Could not initialize " + service + ": " + err.Error()

		if fatal {
			logger.Fatal(message)
		} else {
			logger.Warning(message)
		}
	}
}
