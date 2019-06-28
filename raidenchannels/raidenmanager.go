package raidenchannels

import (
	"math/big"
	"sync"
	"time"

	"github.com/ExchangeUnion/xud-tests/ethclient"

	"github.com/ExchangeUnion/xud-tests/xudclient"

	"github.com/ExchangeUnion/xud-tests/raidenclient"
	"github.com/ExchangeUnion/xud-tests/slackclient"
)

// TODO: clean inactive channels

type token struct {
	address       string
	channelAmount uint64
}

var channelTokens = []token{
	{
		address:       "123",
		channelAmount: 123,
	},
	{
		address:       "987",
		channelAmount: 987,
	},
}

var raidenChannels = make(map[string]bool)

// InitChannelManager initializes a new Raiden channel manager
func InitChannelManager(
	wg *sync.WaitGroup,
	xud *xudclient.Xud,
	raiden *raidenclient.Raiden,
	eth *ethclient.Ethereum,
	slack *slackclient.Slack) {

	wg.Add(1)

	initRaidenChannelsMap(raiden)

	secondTicker := time.NewTicker(30 * time.Second)
	dailyTicker := time.NewTicker(24 * time.Hour)

	go func() {
		defer wg.Done()

		openChannels(xud, raiden, eth, slack)
		balanceChannels(raiden, slack)

		for {
			select {
			case <-secondTicker.C:
				openChannels(xud, raiden, eth, slack)
				balanceChannels(raiden, slack)
				break

			case <-dailyTicker.C:
				raidenChannels = make(map[string]bool)
				initRaidenChannelsMap(raiden)
				break
			}
		}
	}()
}

func initRaidenChannelsMap(raiden *raidenclient.Raiden) {
	log.Debug("Querying and indexing existing Raiden channels")

	// If the node has a channel for one token one can assume that the node has channels for the other tokens too
	token := channelTokens[0]
	channels, err := raiden.ListChannels(token.address)

	if err != nil {
		log.Error("Could not query channels of Raiden: %v", err.Error())
		return
	}

	for _, channel := range channels {
		raidenChannels[channel.PartnerAddress] = true
	}
}

func openChannels(xud *xudclient.Xud, raiden *raidenclient.Raiden, eth *ethclient.Ethereum, slack *slackclient.Slack) {
	log.Debug("Checking XUD for new Raiden partner addresses")

	peers, err := xud.ListPeers()

	if err != nil {
		log.Error("Could not query XUD peers: %v", err.Error())
		return
	}

	for _, peer := range peers.Peers {
		_, hasChannels := raidenChannels[peer.RaidenAddress]

		if !hasChannels {
			raidenChannels[peer.RaidenAddress] = true
			openChannel(raiden, eth, slack, peer.RaidenAddress)
		}
	}
}

func openChannel(raiden *raidenclient.Raiden, eth *ethclient.Ethereum, slack *slackclient.Slack, partnerAddress string) {
	for tokenIndex := range channelTokens {
		go func(token token) {
			err := eth.SendEth(partnerAddress, big.NewInt(100000000000000000))

			if err != nil {
				message := "Could not send ETH to " + partnerAddress + ": " + err.Error()

				log.Warning(message)
				slack.SendMessage(message)
			}

			// TODO: which value should be used for the settle timeout?
			_, err = raiden.OpenChannel(partnerAddress, token.address, token.channelAmount, 512)

			message := "Opened " + token.address + " channel to " + partnerAddress

			if err != nil {
				message = "Could not open Raiden channel: " + err.Error()
			}

			log.Info(message)
			slack.SendMessage(message)

		}(channelTokens[tokenIndex])
	}
}

func balanceChannels(raiden *raidenclient.Raiden, slack *slackclient.Slack) {
	log.Debug("Checking Raiden for channels that need to be rebalanced")

	for _, token := range channelTokens {
		channels, err := raiden.ListChannels(token.address)

		if err != nil {
			log.Warning("Could not query Raiden channels: %v", err.Error())
			return
		}

		for _, channel := range channels {
			// If more than 80% of the balance is on our side it is time to rebalance the channel
			if float64(channel.Balance)/float64(channel.TotalDeposit) > float64(0.8) {
				_, err := raiden.SendPayment(channel.PartnerAddress, channel.TokenAddress, channel.Balance-channel.TotalDeposit/2)

				message := "Rebalanced " + token.address + " channel with " + channel.PartnerAddress

				if err != nil {
					message = "Could not rebalance channel: " + err.Error()
				}

				log.Info(message)
				slack.SendMessage(message)
			}
		}
	}
}
