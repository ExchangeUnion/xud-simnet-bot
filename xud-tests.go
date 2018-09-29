package main

import (
	"fmt"
	"os"
)

func main() {
	err := initConfig()

	if err != nil {
		printError("Could not initialize config:", err)
	}

	err = initLogger(cfg.LogPath)

	if err != nil {
		printError("Could not initialize logger:", err)
	}
}

func printError(messages ...interface{}) {
	fmt.Println(messages...)
	os.Exit(1)
}
