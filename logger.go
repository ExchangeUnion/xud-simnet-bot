package main

import (
	"os"

	"github.com/ExchangeUnion/xud-tests/xudclient"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("")

func initLogger(logFile string) error {
	logging.SetFormatter(logging.MustStringFormatter(
		"%{time:2006/01/02 15:04:05} [%{level}] %{message}",
	))

	file, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)

	if err != nil {
		return err
	}

	logging.SetBackend(
		logging.NewLogBackend(os.Stdout, "", 0),
		logging.NewLogBackend(file, "", 0),
	)

	xudclient.UseLogger(*log)

	return nil
}