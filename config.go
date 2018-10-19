package main

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/ExchangeUnion/xud-tests/lndclient"
	"github.com/ExchangeUnion/xud-tests/slackclient"
	"github.com/ExchangeUnion/xud-tests/xudclient"
	"github.com/jessevdk/go-flags"
)

// xud-tests config types
type helpOptions struct {
	ShowHelp    bool `long:"help" short:"h" description:"Show help"`
	ShowVersion bool `long:"version" short:"v" description:"Show version number"`
}

type config struct {
	DataDir    string `long:"datadir" short:"d" description:"Data directory for xud-tests"`
	ConfigPath string `long:"configpath" description:"Path to the config file"`
	LogPath    string `long:"logpath" description:"Path to the log file"`

	DisableTrading        bool `long:"disabletrading" description:"Whether to disable the trading bot"`
	DisableChannelManager bool `long:"disablechannelmanager" description:"Whether to disable the channel manager"`

	Xud *xudclient.Xud `group:"XUD"`

	Slack *slackclient.Slack `group:"Slack"`

	Help *helpOptions `group:"Help Options"`
}

// XUD config types
type xudConfig struct {
	LndBtc *lndclient.Lnd `toml:"lndbtc"`
	LndLtc *lndclient.Lnd `toml:"lndltc"`
}

var cfg = config{}
var xudCfg = xudConfig{}

func initConfig() error {
	// Ignore unknown flags when parsing command line arguments the first time
	// so that the "unknown flag" error doesn't show up twice
	parser := flags.NewParser(&cfg, flags.IgnoreUnknown)
	parser.Parse()

	if cfg.Help.ShowHelp {
		parser.Usage = "[options]"

		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	if cfg.Help.ShowVersion {
		printVersion()
		os.Exit(0)
	}

	if cfg.ConfigPath == "" {
		getXudTestsDataDir()
		updateDefaultPaths()
	}

	if err := flags.IniParse(cfg.ConfigPath, &cfg); err != nil {
		fmt.Println("Could not find config file:", err)
	}

	// Parse flags again to override config
	flags.Parse(&cfg)

	getXudTestsDataDir()
	updateDefaultPaths()

	if _, err := os.Stat(cfg.DataDir); os.IsNotExist(err) {
		err = os.Mkdir(cfg.DataDir, 0700)

		if err != nil {
			return err
		}
	}

	// Parse XUD config for information about how to connect to the LNDs
	_, err := toml.DecodeFile(cfg.Xud.Config, &xudCfg)

	if err == nil {
		if !xudCfg.LndBtc.Disable {
			setXudLndDefaultValues(xudCfg.LndBtc, true)
		}

		if !xudCfg.LndLtc.Disable {
			setXudLndDefaultValues(xudCfg.LndLtc, false)
		}
	} else {
		fmt.Println("Could not parse config file of XUD:", err)
	}

	return nil
}

func getXudTestsDataDir() {
	if cfg.DataDir == "" {
		cfg.DataDir = getDataDir(applicationName)
	}
}

func setXudLndDefaultValues(lnd *lndclient.Lnd, isBtc bool) {
	dataDir := getDataDir("lnd")

	if lnd.Certificate == "" {
		lnd.Certificate = path.Join(dataDir, "tls.cert")
	}

	if !lnd.DisableMacaroons && lnd.Macaroon == "" {
		// We are using simnet on our test nodes and therfore
		// I assumed that this bot is going to be used on simnet
		// if not explicitly told otherwise
		macaroonDir := path.Join(dataDir, "data", "chain")

		if isBtc {
			macaroonDir = path.Join(macaroonDir, "bitcoin")
		} else {
			macaroonDir = path.Join(macaroonDir, "litecoin")
		}

		lnd.Macaroon = path.Join(macaroonDir, "admin.macaroon")
	}
}

func updateDefaultPaths() {
	// xud-tests paths
	configPath := path.Join(cfg.DataDir, applicationName+".conf")
	logPath := path.Join(cfg.DataDir, applicationName+".logs")

	if cfg.ConfigPath == "" {
		cfg.ConfigPath = configPath
	}

	if cfg.LogPath == "" {
		cfg.LogPath = logPath
	}

	// XUD paths
	xudDir := getDataDir("xud")

	if cfg.Xud.Certificate == "" {
		cfg.Xud.Certificate = path.Join(xudDir, "tls.cert")
	}

	if cfg.Xud.Config == "" {
		cfg.Xud.Config = path.Join(xudDir, "xud.conf")
	}
}

func getDataDir(application string) (dir string) {
	homeDir := getHomeDir()

	switch runtime.GOOS {
	case "darwin":
		dir = path.Join(homeDir, application)
		break

	case "windows":
		dir = path.Join(homeDir, strings.Title(application))
		break

	default:
		dir = path.Join(homeDir, "."+application)
		break
	}

	return cleanPath(dir)
}

func getHomeDir() (dir string) {
	usr, _ := user.Current()

	switch runtime.GOOS {
	case "darwin":
		dir = path.Join(usr.HomeDir, "Library", "Application Support")
		break

	case "windows":
		dir = path.Join(usr.HomeDir, "AppData", "Local")
		break

	default:
		dir = usr.HomeDir
		break
	}

	return cleanPath(dir)
}

func cleanPath(path string) string {
	path = filepath.Clean(os.ExpandEnv(path))
	return strings.Replace(path, "\\", "/", -1)
}
