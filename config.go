package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/ExchangeUnion/xud-tests/build"
	"github.com/ExchangeUnion/xud-tests/channels"
	"github.com/ExchangeUnion/xud-tests/database"
	"github.com/ExchangeUnion/xud-tests/discord"
	"github.com/ExchangeUnion/xud-tests/xudrpc"
	"github.com/jessevdk/go-flags"
	"os"
)

type helpOptions struct {
	ShowHelp    bool `short:"h" long:"help" description:"Display this help message"`
	ShowVersion bool `short:"v" long:"version" description:"Display version and exit"`
}

type config struct {
	ConfigFile string `short:"c" long:"configfile" description:"Path to configuration file"`
	LogFile    string `short:"l" long:"logfile" description:"Path to the log file"`

	Xud     *xudrpc.Xud      `group:"XUD Options"`
	Discord *discord.Discord `group:"Discord Options"`

	Database       *database.Database       `group:"Database options"`
	ChannelManager *channels.ChannelManager `group:"Channel Manager Options"`

	// This option is only parsed in the TOML config file
	Channels []*channels.Channel

	Help *helpOptions `group:"Help Options"`
}

func loadConfig() *config {
	cfg := config{
		LogFile:    "./xud-tests.log",
		ConfigFile: "./xud-tests.toml",

		Database: &database.Database{
			FileName: "./xud-tests.json",
		},
	}

	parser := flags.NewParser(&cfg, flags.IgnoreUnknown)
	_, err := parser.Parse()

	if cfg.Help.ShowHelp {
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	if cfg.Help.ShowVersion {
		fmt.Println(build.GetVersion())
		os.Exit(0)
	}

	if err != nil {
		printCouldNotParseCli(err)
	}

	if cfg.ConfigFile != "" {
		_, err := toml.DecodeFile(cfg.ConfigFile, &cfg)

		if err != nil {
			fmt.Println("Could not read config file: " + err.Error())
		}
	}

	_, err = flags.Parse(&cfg)

	if err != nil {
		printCouldNotParseCli(err)
	}

	return &cfg
}

func printCouldNotParseCli(err error) {
	printFatal("Could not parse CLI arguments: %s", err)
}
