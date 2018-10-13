package channels

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ExchangeUnion/xud-tests/lndclient"
	"github.com/lightningnetwork/lnd/lnrpc"
)

// TODO: handle cases in which a remote has multiple channels
// TODO: try not to open a new channel if there is already a pending channel

const newChannelAmt = 10000000

// channelCloseTimeout defines after how many seconds a channel times out and should be closed
const channelCloseTimeout = time.Duration(2 * 24 * time.Hour)

// inactiveTimes is a map between the public key of a node and the last time it was seen
var inactiveTimes = make(map[string]time.Time)
var inactiveTimesLock = sync.RWMutex{}

var nodeName string

// InitChannelManager initializes a new channel manager
func InitChannelManager(lnd *lndclient.Lnd, name string) {
	nodeName = name

	handleChannels(lnd)

	ticker := time.NewTicker(30 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				handleChannels(lnd)
				break
			}
		}
	}()
}

func handleChannels(lnd *lndclient.Lnd) {
	channels, err := lnd.ListChannels()

	if err != nil {
		logCouldNotConnect(err)
	}

	channelsMap := getChannelsMap(channels.Channels)

	openNewChannels(lnd, channelsMap)
	closeTimedOutChannels(lnd, channelsMap)
}

func openNewChannels(lnd *lndclient.Lnd, channels map[string]*lnrpc.Channel) {
	peers, err := lnd.ListPeers()

	if err != nil {
		logCouldNotConnect(err)
	}

	inactiveTimesLock.RLock()

	for _, peer := range peers.Peers {
		_, hasChannel := channels[peer.PubKey]

		if !hasChannel {
			log.Debug("Opening new %v channel channel to: %v", nodeName, peer.PubKey)

			_, err := lnd.OpenChannel(lnrpc.OpenChannelRequest{
				NodePubkeyString:   peer.PubKey,
				LocalFundingAmount: newChannelAmt,
				PushSat:            newChannelAmt / 2,
			})

			if err != nil {
				logCouldNotConnect(err)
			}
		}
	}

	inactiveTimesLock.RUnlock()
}

func closeTimedOutChannels(lnd *lndclient.Lnd, channels map[string]*lnrpc.Channel) {
	now := time.Now()

	for _, channel := range channels {
		inactiveTimesLock.RLock()
		lastSeen, isInMap := inactiveTimes[channel.RemotePubkey]
		inactiveTimesLock.RUnlock()

		inactiveTimesLock.Lock()
		if channel.Active {
			if isInMap {
				delete(inactiveTimes, channel.RemotePubkey)
			}
		} else {
			if isInMap {
				if now.Sub(lastSeen) > channelCloseTimeout {
					log.Debug("Closing %v channel channel to: %v", nodeName, channel.RemotePubkey)

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

		inactiveTimesLock.Unlock()
	}
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
