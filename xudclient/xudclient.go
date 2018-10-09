package xudclient

import (
	"context"
	"errors"
	"io"
	"strconv"

	"google.golang.org/grpc/credentials"

	"github.com/ExchangeUnion/xud-tests/xudrpc"
	"google.golang.org/grpc"
)

// Xud represents a XUD client
type Xud struct {
	GrpcHost        string `long:"xud.host" default:"localhost" description:"XUD gRPC service host"`
	GrpcPort        int    `long:"xud.port" default:"8886" description:"XUD gRPC service port"`
	GrpcCertificate string `long:"xud.certificatepath" description:"Path to the certificate of XUD gRPC"`

	ctx    context.Context
	client xudrpc.XudClient
}

// OrderRemoved is a callback that allows clients to get notified about order removals
type OrderRemoved func(order xudrpc.OrderRemoval)

// Connect to a XUD node
func (xud *Xud) Connect() error {
	creds, err := credentials.NewClientTLSFromFile(xud.GrpcCertificate, "")

	if err != nil {
		return err
	}

	uri := xud.GrpcHost + ":" + strconv.Itoa(xud.GrpcPort)
	con, err := grpc.Dial(uri, grpc.WithTransportCredentials(creds))

	if err != nil {
		return err
	}

	if xud.ctx == nil {
		xud.ctx = context.Background()
	}

	xud.client = xudrpc.NewXudClient(con)

	return err
}

// GetInfo gets general information about the XUD node
func (xud *Xud) GetInfo() (*xudrpc.GetInfoResponse, error) {
	return xud.client.GetInfo(xud.ctx, &xudrpc.GetInfoRequest{})
}

// PlaceOrderSync places a new order in XUD
func (xud *Xud) PlaceOrderSync(request xudrpc.PlaceOrderRequest) (*xudrpc.PlaceOrderResponse, error) {
	return xud.client.PlaceOrderSync(xud.ctx, &request)
}

// SubscribeRemovedOrders notifies the client via a callback about removed orders
func (xud *Xud) SubscribeRemovedOrders(callback OrderRemoved) error {
	stream, streamErr := xud.client.SubscribeRemovedOrders(xud.ctx, &xudrpc.SubscribeRemovedOrdersRequest{})

	if streamErr != nil {
		return streamErr
	}

	wait := make(chan struct{})

	go func() {
		for {
			orderRemoval, err := stream.Recv()

			if err != nil {
				if err == io.EOF {
					err = errors.New("lost connection to XUD")
				}

				streamErr = err
				close(wait)
				return
			}

			callback(*orderRemoval)
		}
	}()

	<-wait

	return streamErr
}
