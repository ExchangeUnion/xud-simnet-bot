package lndclient

import (
	"context"
	"encoding/hex"
	"io/ioutil"
	"strconv"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc"
)

// Lnd represents a LND client
type Lnd struct {
	Disable bool

	Host string
	Port int

	Certificate string `toml:"certpath"`

	DisableMacaroons bool   `toml:"nomacaroons"`
	Macaroon         string `toml:"macaroonpath"`

	ctx    context.Context
	client lnrpc.LightningClient
}

// ChannelCloseUpdate is a callback that allows clients to get notified about channel closing events
type ChannelCloseUpdate func(update lnrpc.CloseStatusUpdate)

// Init to the LND node
func (lnd *Lnd) Init() error {
	creds, err := credentials.NewClientTLSFromFile(lnd.Certificate, "")

	if err != nil {
		return err
	}

	uri := lnd.Host + ":" + strconv.Itoa(lnd.Port)
	con, err := grpc.Dial(uri, grpc.WithTransportCredentials(creds))

	if err != nil {
		return err
	}

	if lnd.ctx == nil {
		lnd.ctx = context.Background()

		if !lnd.Disable && lnd.Macaroon != "" {
			macaroon, err := getMacaroon(lnd.Macaroon)

			if err == nil {
				lnd.ctx = metadata.NewOutgoingContext(lnd.ctx, macaroon)
			} else {
				return err
			}
		}
	}

	lnd.client = lnrpc.NewLightningClient(con)

	return nil
}

func getMacaroon(macaroonPath string) (macaroon metadata.MD, err error) {
	data, err := ioutil.ReadFile(macaroonPath)

	if err == nil {
		macaroon = metadata.Pairs("macaroon", hex.EncodeToString(data))
	}

	return macaroon, err
}

// GetInfo returns general information about the LND node
func (lnd *Lnd) GetInfo() (*lnrpc.GetInfoResponse, error) {
	return lnd.client.GetInfo(lnd.ctx, &lnrpc.GetInfoRequest{})
}

// ListPeers gets a list of all currently active peers
func (lnd *Lnd) ListPeers() (*lnrpc.ListPeersResponse, error) {
	return lnd.client.ListPeers(lnd.ctx, &lnrpc.ListPeersRequest{})
}

// ListChannels gets alist of all open channels of the node
func (lnd *Lnd) ListChannels() (*lnrpc.ListChannelsResponse, error) {
	return lnd.client.ListChannels(lnd.ctx, &lnrpc.ListChannelsRequest{})
}

// OpenChannel opens a new channel
func (lnd *Lnd) OpenChannel(request lnrpc.OpenChannelRequest) (*lnrpc.ChannelPoint, error) {
	return lnd.client.OpenChannelSync(lnd.ctx, &request)
}

// CloseChannel attempts to close a channel
func (lnd *Lnd) CloseChannel(request lnrpc.CloseChannelRequest) error {
	_, err := lnd.client.CloseChannel(lnd.ctx, &request)

	return err
}
