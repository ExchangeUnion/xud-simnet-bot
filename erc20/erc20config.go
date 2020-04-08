package main

import (
	"github.com/ExchangeUnion/xud-tests/ethclient"
	"github.com/jessevdk/go-flags"
)

type config struct {
	TokenAddress     string `long:"token" description:"Address of the ERC20 token"`
	RecipientAddress string `long:"recipient" description:"Address to which the tokens should be sent"`
	Amount           string `long:"amount" description:"Amount of tokens that should be sent in the smallest denomation of the token"`

	Ethereum *ethclient.Ethereum `group:"ETH"`
}

var cfg = config{}

func initConfig() error {
	parser := flags.NewParser(&cfg, flags.Default)
	_, err := parser.Parse()

	return err
}
