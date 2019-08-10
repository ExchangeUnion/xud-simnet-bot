package raidenchannels

import (
	"fmt"
	"math"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/ExchangeUnion/xud-tests/xudrpc"

	"github.com/ExchangeUnion/xud-tests/ethclient"

	"github.com/ExchangeUnion/xud-tests/xudclient"

	"github.com/ExchangeUnion/xud-tests/discordclient"
	"github.com/ExchangeUnion/xud-tests/raidenclient"
)

type token struct {
	address       string
	channelAmount float64
}

var channelTokens = []token{
	// WETH token
	{
		address:       "0x9F50cEA29307d7D91c5176Af42f3aB74f0190dD3",
		channelAmount: 1e21,
	},
	// DAI token
	{
		address:       "0x76671A2831Dc0aF53B09537dea57F1E22899655d",
		channelAmount: 1 * 1e23,
	},
}

// channelCloseTimeout defines after how many seconds a channel times out and should be closed
const channelCloseTimeout = time.Duration(2 * 24 * time.Hour)

var inactiveTimes = make(map[string]time.Time)

var raidenChannelsMap = make(map[string]map[string]bool)

// InitChannelManager initializes a new Raiden channel manager
func InitChannelManager(
	wg *sync.WaitGroup,
	xud *xudclient.Xud,
	raiden *raidenclient.Raiden,
	eth *ethclient.Ethereum,
	discord *discordclient.Discord,
	dataDir string,
	enableBalancing bool) {

	wg.Add(1)

	dataPath := path.Join(dataDir, "raiden.dat")

	readInactiveTimes(dataPath)

	initRaidenChannelsMap(raiden)

	secondTicker := time.NewTicker(30 * time.Second)
	dailyTicker := time.NewTicker(24 * time.Hour)

	go func() {
		defer wg.Done()

		openChannels(xud, raiden, eth, discord, dataPath)

		if enableBalancing {
			balanceChannels(raiden, discord)
		}

		for {
			select {
			case <-secondTicker.C:
				openChannels(xud, raiden, eth, discord, dataPath)

				if enableBalancing {
					balanceChannels(raiden, discord)
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

		log.Debug("Initialized token: " + token.address)
	}
}

func openChannels(xud *xudclient.Xud, raiden *raidenclient.Raiden, eth *ethclient.Ethereum, discord *discordclient.Discord, dataPath string) {
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
				hasChannel := channelMap[peer.RaidenAddress]

				if !hasChannel {
					sendEther(eth, discord, peer.RaidenAddress)
					openChannel(raiden, eth, discord, token, peer.RaidenAddress)
				} else {
					log.Debug(peer.RaidenAddress + " already has a " + token.address + " channel. Skipping")
				}
			}
		}
	}

	go updateInactiveTimes(peers.Peers, raiden, discord, dataPath)
}

func updateInactiveTimes(peers []*xudrpc.Peer, raiden *raidenclient.Raiden, discord *discordclient.Discord, dataPath string) {
	/*log.Debug("Checking for inactive Raiden channels")

	now := time.Now()

	// Remove peers that are active from the map
	for _, peer := range peers {
		inactiveTimes[peer.RaidenAddress] = now
	}

	channels, err := raiden.ListChannels("")

	if err != nil {
		log.Error("Could not query Raiden channels")
	}

	for _, channel := range channels {
		lastSeen, isInMap := inactiveTimes[channel.PartnerAddress]

		if channel.State == "opened" {
			if isInMap {
				if now.Sub(lastSeen) > channelCloseTimeout {
					delete(inactiveTimes, channel.PartnerAddress)

					for _, token := range channelTokens {
						_, err := raiden.CloseChannel(channel.PartnerAddress, token.address)

						raidenChannelsMap[token.address][channel.PartnerAddress] = false

						log.Debug("About to close channel " + token.address + "/" + channel.PartnerAddress)

						sendMessage(
							discord,
							"Closed "+token.address+" channel to "+channel.PartnerAddress,
							"Could not close "+token.address+" channel to "+channel.PartnerAddress+": "+fmt.Sprint(err),
							err,
						)
					}
				}
			}
		}
	}

	saveInactiveTimes(dataPath)*/
}

func sendEther(eth *ethclient.Ethereum, discord *discordclient.Discord, partnerAddress string) {
	/*balance, err := eth.EthBalance(partnerAddress)

	if err != nil {
		message := "Could not query Ether balance of " + partnerAddress + " : " + err.Error()

		log.Warning(message)
		discord.SendMessage(message)
		return
	}

	// If the Ether balance of the other side is 0, send 1 Ether
	if balance.Cmp(big.NewInt(0)) == 0 {
		err := eth.SendEth(partnerAddress, big.NewInt(1000000000000000000))

		sendMesssage(
			discord,
			"Sent Ether to "+partnerAddress,
			"Could not send Ether to "+partnerAddress+": "+fmt.Sprint(err),
			err,
		)

		if err != nil {
			return
		}
	} else {
		etherBalance := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1000000000000000000))
		log.Debug("Not sending Ether to " + partnerAddress + " because it has a balance of: " + etherBalance.String())
	}*/
}

func openChannel(raiden *raidenclient.Raiden, eth *ethclient.Ethereum, discord *discordclient.Discord, token token, partnerAddress string) {
	raidenChannelsMap[token.address][partnerAddress] = true

	go func() {
		sendMessage(
			discord,
			"About to open "+token.address+" channel to "+partnerAddress,
			"",
			nil,
		)

		_, err := raiden.OpenChannel(partnerAddress, token.address, token.channelAmount, eth.SettleTimeout)

		sendMessage(
			discord,
			"Opened "+token.address+" channel to "+partnerAddress,
			"Could not open "+token.address+" channel to "+partnerAddress+": "+fmt.Sprint(err),
			err,
		)

		if err != nil {
			return
		}

		paymentAmount := math.Round(token.channelAmount / 2)

		log.Info("Sending " + strconv.FormatFloat(paymentAmount, 'f', -1, 64) + " " + token.address + " to " + partnerAddress)

		_, err = raiden.SendPayment(partnerAddress, token.address, paymentAmount)

		sendMessage(
			discord,
			"Sent half of "+token.address+" channel capacity to "+partnerAddress,
			"Could send half of "+token.address+" to "+partnerAddress+": "+fmt.Sprint(err),
			err,
		)
	}()
}

func balanceChannels(raiden *raidenclient.Raiden, discord *discordclient.Discord) {
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
				discord.SendMessage(message)
			}
		}
	}
}

func sendMessage(discord *discordclient.Discord, message string, errorMessage string, err error) {
	if err == nil {
		log.Info(message)
		discord.SendMessage(message)
	} else {
		log.Warning(errorMessage)
		discord.SendMessage(errorMessage)
	}
}
