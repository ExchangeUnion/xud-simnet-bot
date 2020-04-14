package channels

import (
	"github.com/ExchangeUnion/xud-tests/database"
	"github.com/ExchangeUnion/xud-tests/discord"
	"github.com/ExchangeUnion/xud-tests/xudrpc"
	"github.com/google/logger"
	"time"
)

type ChannelManager struct {
	Interval int `long:"manager.interval" default:"20" description:"Interval in seconds at which new channels should be opened"`

	channels []*Channel

	xud      *xudrpc.Xud
	discord  *discord.Discord
	database *database.Database
}

type Channel struct {
	Currency   string
	Amount     int64
	PushAmount int64
}

func (manager *ChannelManager) Init(channels []*Channel, xud *xudrpc.Xud, discord *discord.Discord, database *database.Database) {
	logger.Info("Initializing channel manager")

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
				Amount:         channel.Amount,
				PushAmount:     channel.PushAmount,
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
