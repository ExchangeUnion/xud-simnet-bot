package lndclient

import (
	"context"
	"errors"
	"io"
	"strconv"

	"google.golang.org/grpc/credentials"

	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc"
)

// Lnd represents a LND client
type Lnd struct {
	Host string
	Port int

	Certificate string

	Macaroon string

	ctx    context.Context
	client lnrpc.LightningClient
}

// ChannelCloseUpdate is a callback that allows clients to get notified about channel closing events
type ChannelCloseUpdate func(update lnrpc.CloseStatusUpdate)

// Connect to the LND node
func (lnd *Lnd) Connect() error {
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
	}

	lnd.client = lnrpc.NewLightningClient(con)

	return nil
}

// ListPeers gets a list of all currently active peers
func (lnd *Lnd) ListPeers() (*lnrpc.ListPeersResponse, error) {
	return lnd.client.ListPeers(lnd.ctx, &lnrpc.ListPeersRequest{})
}

// OpenChannel opens a new channel
func (lnd *Lnd) OpenChannel(request lnrpc.OpenChannelRequest) (*lnrpc.ChannelPoint, error) {
	return lnd.client.OpenChannelSync(lnd.ctx, &request)
}

// CloseChannel attempts to close a channel
func (lnd *Lnd) CloseChannel(request lnrpc.CloseChannelRequest, callback ChannelCloseUpdate) error {
	stream, streamErr := lnd.client.CloseChannel(lnd.ctx, &request)

	if streamErr != nil {
		return streamErr
	}

	wait := make(chan struct{})

	go func() {
		for {
			update, err := stream.Recv()

			if err != nil {
				if err == io.EOF {
					err = errors.New("lost connection to LND")
				}

				streamErr = err
				close(wait)
				return
			}

			callback(*update)
		}
	}()

	<-wait

	return streamErr
}
