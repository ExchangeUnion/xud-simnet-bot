package main

import (
	"fmt"
	"os"
)

func main() {
	err := initConfig()

	if err != nil {
		os.Exit(1)
	}

	tx, err := cfg.Ethereum.SendToken(cfg.TokenAddress, cfg.RecipientAddress, cfg.Amount)

	if err == nil {
		fmt.Println("Sent transaction: ", tx.Hash().String())
	} else {
		fmt.Println("Could not send tokens: ", err)
	}
}
