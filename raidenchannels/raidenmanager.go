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
	channelAmount float64
}

var channelTokens = []token{
	// WETH token
	{
		address:       "0x9F50cEA29307d7D91c5176Af42f3aB74f0190dD3",
		channelAmount: 10e21,
	},
	// DAI token
	{
		address:       "0x76671A2831Dc0aF53B09537dea57F1E22899655d",
		channelAmount: 3.25 * 10e23,
	},
}

var raidenChannelsMap = make(map[string]map[string]bool)

// InitChannelManager initializes a new Raiden channel manager
func InitChannelManager(
	wg *sync.WaitGroup,
	xud *xudclient.Xud,
	raiden *raidenclient.Raiden,
	eth *ethclient.Ethereum,
	slack *slackclient.Slack,
	enableBalancing bool) {

	wg.Add(1)

	initRaidenChannelsMap(raiden)

	secondTicker := time.NewTicker(30 * time.Second)
	dailyTicker := time.NewTicker(24 * time.Hour)

	go func() {
		defer wg.Done()

		openChannels(xud, raiden, eth, slack)

		if enableBalancing {
			balanceChannels(raiden, slack)
		}

		for {
			select {
			case <-secondTicker.C:
				openChannels(xud, raiden, eth, slack)

				if enableBalancing {
					balanceChannels(raiden, slack)
				}
				break

			case <-dailyTicker.C:
				log.Info("Resetting Raiden channel map")
				raidenChannelsMap = make(map[string]map[string]bool)
				initRaidenChannelsMap(raiden)
				break
			}
		}
	}()
}

func initRaidenChannelsMap(raiden *raidenclient.Raiden) {
	log.Debug("Querying and indexing existing Raiden channels")

	for _, token := range channelTokens {
		raidenChannelsMap[token.address] = make(map[string]bool)

		channels, err := raiden.ListChannels(token.address)

		if err != nil {
			log.Error("Could not query channels of Raiden: %v", err.Error())
			return
		}

		for _, channel := range channels {
			raidenChannelsMap[token.address][channel.PartnerAddress] = true
		}
	}
}

func openChannels(xud *xudclient.Xud, raiden *raidenclient.Raiden, eth *ethclient.Ethereum, slack *slackclient.Slack) {
	log.Debug("Checking XUD for new Raiden partner addresses")

	peers, err := xud.ListPeers()

	if err != nil {
		log.Error("Could not query XUD peers: %v", err.Error())
		return
	}

	for _, token := range channelTokens {
		channelMap := raidenChannelsMap[token.address]

		for _, peer := range peers.Peers {
			if peer.RaidenAddress != "" {
				_, hasChannel := channelMap[peer.RaidenAddress]

				if !hasChannel {
					sendEther(eth, slack, peer.RaidenAddress)
					openChannel(raiden, slack, token, peer.RaidenAddress)
				} else {
					log.Debug(peer.RaidenAddress + " already has a " + token.address + " channel. Skipping")
				}
			}
		}
	}
}

func sendEther(eth *ethclient.Ethereum, slack *slackclient.Slack, partnerAddress string) {
	balance, err := eth.EthBalance(partnerAddress)

	if err != nil {
		message := "Could not query Ether balance of " + partnerAddress + " : " + err.Error()

		log.Warning(message)
		slack.SendMessage(message)
		return
	}

	// If the Ether balance of the other side is 0, send 1 Ether
	if balance.Cmp(big.NewInt(0)) == 0 {
		err := eth.SendEth(partnerAddress, big.NewInt(1000000000000000000))

		sendMesssage(
			slack,
			"Sent Ether to "+partnerAddress,
			"Could not send Ether to "+partnerAddress+": "+err.Error(),
			err,
		)

		if err != nil {
			return
		}
	} else {
		etherBalance := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1000000000000000000))
		log.Debug("Not sending Ether to " + partnerAddress + " because it has a balance of: " + etherBalance.String())
	}
}

func openChannel(raiden *raidenclient.Raiden, slack *slackclient.Slack, token token, partnerAddress string) {
	raidenChannelsMap[token.address][partnerAddress] = true

	go func() {
		_, err := raiden.OpenChannel(partnerAddress, token.address, token.channelAmount, 500)

		sendMesssage(
			slack,
			"Opened "+token.address+" channel to "+partnerAddress,
			"Could not open "+token.address+" channel to "+partnerAddress+": "+err.Error(),
			err,
		)

		if err != nil {
			return
		}

		_, err = raiden.SendPayment(partnerAddress, token.address, token.channelAmount/2)

		sendMesssage(
			slack,
			"Sent half of "+token.address+"channel capacity to "+partnerAddress,
			"Could send half of "+token.address+" to "+partnerAddress+": "+err.Error(),
			err,
		)
	}()
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
			if channel.Balance/token.channelAmount > float64(0.8) {
				_, err := raiden.SendPayment(channel.PartnerAddress, channel.TokenAddress, channel.Balance-(token.channelAmount/float64(2)))

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

func sendMesssage(slack *slackclient.Slack, message string, errorMessage string, err error) {
	if err == nil {
		log.Info(message)
		slack.SendMessage(message)
	} else {
		log.Warning(errorMessage)
		slack.SendMessage(errorMessage)
	}
}
