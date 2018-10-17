package channels

import (
	"encoding/gob"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ExchangeUnion/xud-tests/slackclient"

	"github.com/ExchangeUnion/xud-tests/lndclient"
	"github.com/lightningnetwork/lnd/lnrpc"
)

// TODO: handle cases in which a remote has multiple channels

const newChannelAmt = 10000000

// channelCloseTimeout defines after how many seconds a channel times out and should be closed
const channelCloseTimeout = time.Duration(2 * 24 * time.Hour)

// InitChannelManager initializes a new channel manager
func InitChannelManager(wg *sync.WaitGroup, lnd *lndclient.Lnd, slack *slackclient.Slack, dataDir string, nodeName string) {
	// inactiveTimes is a map between the public key of a node and the last time it was seen
	inactiveTimes := make(map[string]time.Time)

	// Path to the latest copy of inactiveTimes on the disk
	dataPath := path.Join(dataDir, nodeName+".dat")

	readInactiveTimes(dataPath, &inactiveTimes)

	wg.Add(1)

	ticker := time.NewTicker(30 * time.Second)

	go func() {
		defer wg.Done()

		handleChannels(lnd, nodeName, slack, inactiveTimes, dataPath)

		for {
			select {
			case <-ticker.C:
				handleChannels(lnd, nodeName, slack, inactiveTimes, dataPath)
				break
			}
		}
	}()
}

func handleChannels(lnd *lndclient.Lnd, nodeName string, slack *slackclient.Slack, inactiveTimes map[string]time.Time, dataPath string) {
	channels, err := lnd.ListChannels()

	if err != nil {
		logCouldNotConnect(nodeName, err)
		return
	}

	channelsMap := getChannelsMap(channels.Channels)

	openNewChannels(lnd, nodeName, slack, channelsMap)
	closeTimedOutChannels(lnd, nodeName, slack, inactiveTimes, channelsMap)

	saveInactiveTimes(dataPath, inactiveTimes)
}

func openNewChannels(lnd *lndclient.Lnd, nodeName string, slack *slackclient.Slack, channels map[string]*lnrpc.Channel) {
	peers, err := lnd.ListPeers()

	if err != nil {
		logCouldNotConnect(nodeName, err)
		return
	}

	pendingChannels, err := lnd.PendingChannels()

	if err != nil {
		logCouldNotConnect(nodeName, err)
		return
	}

	pendingOpen := pendingChannels.GetPendingOpenChannels()

	for _, peer := range peers.Peers {
		_, hasChannel := channels[peer.PubKey]

		if !hasChannel && !pendingOpenChannelsContainsPeer(pendingOpen, peer.PubKey) {
			message := "Opening new " + nodeName + " channel to: " + peer.PubKey
			log.Info(message)
			slack.SendMessage(message)

			_, err := lnd.OpenChannel(lnrpc.OpenChannelRequest{
				NodePubkeyString:   peer.PubKey,
				LocalFundingAmount: newChannelAmt,
				PushSat:            newChannelAmt / 2,
			})

			if err != nil {
				logCouldNotConnect(nodeName, err)
			}
		}
	}
}

func closeTimedOutChannels(lnd *lndclient.Lnd, nodeName string, slack *slackclient.Slack, inactiveTimes map[string]time.Time, channels map[string]*lnrpc.Channel) {
	now := time.Now()

	for _, channel := range channels {
		lastSeen, isInMap := inactiveTimes[channel.RemotePubkey]

		if channel.Active {
			if isInMap {
				delete(inactiveTimes, channel.RemotePubkey)
			}
		} else {
			if isInMap {
				if now.Sub(lastSeen) > channelCloseTimeout {
					message := "Closing " + nodeName + " channel to: " + channel.RemotePubkey
					log.Info(message)
					slack.SendMessage(message)

					lnd.CloseChannel(lnrpc.CloseChannelRequest{
						ChannelPoint: getChannelPoint(channel.ChannelPoint),
						Force:        true,
					})

					delete(inactiveTimes, channel.RemotePubkey)
				}
			} else {
				inactiveTimes[channel.RemotePubkey] = now
			}
		}
	}
}

func saveInactiveTimes(dataPath string, data map[string]time.Time) {
	file, err := os.OpenFile(dataPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	defer file.Close()

	if err != nil {
		log.Warning("Could not write channel data to disk: %v", err)
		return
	}

	encoder := gob.NewEncoder(file)
	encoder.Encode(data)
}

func readInactiveTimes(dataPath string, data *map[string]time.Time) {
	if _, err := os.Stat(dataPath); err != nil {
		// File does not exist
		return
	}

	file, err := os.Open(dataPath)
	defer file.Close()

	if err != nil {
		log.Warning("Could not read channel data from disk: %v", err)
		return
	}

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(data)

	if err != nil {
		log.Warning("Could not parse channel data from disk: %v", err)
	}
}

func pendingOpenChannelsContainsPeer(pendingChannels []*lnrpc.PendingChannelsResponse_PendingOpenChannel, peerPubKey string) bool {
	for _, channel := range pendingChannels {
		if channel.Channel.RemoteNodePub == peerPubKey {
			return true
		}
	}

	return false
}

func getChannelsMap(channels []*lnrpc.Channel) map[string]*lnrpc.Channel {
	channelsMap := make(map[string]*lnrpc.Channel)

	for _, channel := range channels {
		channelsMap[channel.RemotePubkey] = channel
	}

	return channelsMap
}

func getChannelPoint(channelPoint string) *lnrpc.ChannelPoint {
	split := strings.Split(channelPoint, ":")
	output, _ := strconv.Atoi(split[1])

	return &lnrpc.ChannelPoint{
		FundingTxid: &lnrpc.ChannelPoint_FundingTxidStr{
			FundingTxidStr: split[0],
		},
		OutputIndex: uint32(output),
	}
}
