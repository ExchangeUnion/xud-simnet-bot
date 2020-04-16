package channels

import (
	"github.com/ExchangeUnion/xud-simnet-bot/database"
	"github.com/ExchangeUnion/xud-simnet-bot/discord"
	"github.com/ExchangeUnion/xud-simnet-bot/xudrpc"
	"github.com/google/logger"
	"math"
	"time"
)

type ChannelManager struct {
	Interval int `long:"manager.interval" default:"20" description:"Interval in seconds at which new channels should be opened"`

	channels []*Channel

	xud      *xudrpc.Xud
	discord  *discord.Discord
	database *database.Database
}

// This struct has a dualistic nature:
// 1. Lightning channel creations
//    "Currency", "Amount" and "PushAmount" need to be set
//
// 2. Ethereum token faucet
//    "Currency", "TokenAddress", "Amount" need to be set
//
//    If the "TokenAddress" is set or the "Currency" equals "ETH"
//    no channels will be created for that currency but the faucet
//    will send tokens on request
type Channel struct {
	// Symbol of the currency
	Currency string
	// Address of the Ethereum token
	TokenAddress string
	// Capacity of the channel or amount of token that should be sent
	Amount float64
	// Amount that should be pushed to the other side in case of a channel creation
	PushAmount float64
}

var decimals = math.Pow(10, 8)

func (manager *ChannelManager) Init(channels []*Channel, xud *xudrpc.Xud, discord *discord.Discord, database *database.Database) {
	logger.Info("Initializing channel manager")

	for i := len(channels) - 1; i >= 0; i-- {
		entry := channels[i]

		if entry.TokenAddress != "" || entry.Currency == "ETH" {
			channels = channels[:i+copy(channels[i:], channels[i+1:])]
		}
	}

	manager.channels = channels

	manager.xud = xud
	manager.discord = discord
	manager.database = database

	ticker := time.NewTicker(time.Duration(manager.Interval) * time.Second)

	manager.openChannels()

	for range ticker.C {
		manager.openChannels()
	}
}

func (manager *ChannelManager) openChannels() {
	peers, err := manager.xud.ListPeers()

	if err != nil {
		message := "Could not get XUD peers: " + err.Error()

		logger.Warning(message)
		_ = manager.discord.SendMessage(message)
		return
	}

	for _, peer := range peers.Peers {
		channelsOpened := manager.database.GetChannelsOpened(peer.NodePubKey)

		if len(channelsOpened) == len(manager.channels) {
			continue
		}

		nodeInfo := "**" + peer.Alias + "** (`" + peer.NodePubKey + "`)"

		for _, channel := range manager.channels {
			channelOpenedAlready := false

			for _, channelOpened := range channelsOpened {
				if channelOpened == channel.Currency {
					channelOpenedAlready = true
					break
				}
			}

			if channelOpenedAlready {
				continue
			}

			message := "Opening " + channel.Currency + " channel to " + nodeInfo

			logger.Info(message)
			_ = manager.discord.SendMessage(message)

			_, err := manager.xud.OpenChannel(&xudrpc.OpenChannelRequest{
				Amount:         coinsToSatoshis(channel.Amount),
				PushAmount:     coinsToSatoshis(channel.PushAmount),
				Currency:       channel.Currency,
				NodeIdentifier: peer.NodePubKey,
			})

			if err != nil {
				// Ignore common LND error and retry opening the channel next iteration
				if err.Error() == "rpc error: code = Code(102) desc = Synchronizing blockchain" {
					continue
				}

				message = "Could not open " + channel.Currency + " channel to " + nodeInfo + ": " + err.Error()

				logger.Warning(message)
				_ = manager.discord.SendMessage(message)
				continue
			}

			manager.database.AddChannelsOpened(peer.NodePubKey, channel.Currency)
		}

	}
}

func coinsToSatoshis(coins float64) int64 {
	return int64(math.Round(coins * decimals))
}
