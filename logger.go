package main

import (
	"fmt"
	"log"
	"os"

	"github.com/google/logger"
)

func initLogger(logPath string) {
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)

	if err != nil {
		printFatal("Could not open log file: %s", err)
	}

	logger.Init("xud-simnet-bot", true, true, file)
	logger.SetFlags(log.LstdFlags)

	logger.Info("Initialized logger")
}

func printFatal(format string, a ...interface{}) {
	fmt.Printf(format+"\n", a...)
	os.Exit(1)
}

func logConfig(cfg *config) {
	logger.Info("Loaded config: " + stringify(cfg))
}
