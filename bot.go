package main

import (
	"github.com/ExchangeUnion/xud-simnet-bot/xudrpc"
	"github.com/google/logger"
)

func main() {
	cfg := loadConfig()
	initLogger(cfg.LogFile)
	logConfig(cfg)

	info := initXud(cfg)
	initDiscord(cfg, info)

	cfg.Database.Init()

	cfg.ChannelManager.Init(cfg.Channels, cfg.Xud, cfg.Discord, cfg.Database)

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
	checkError("Discord", err, false)

	err = cfg.Discord.SendMessage("Started xud-simnet-bot with XUD node: **" + info.Alias + "** (`" + info.NodePubKey + "`)")
	checkError("Discord", err, false)

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
