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

// Xud represents a XUD Client
type Xud struct {
	Host string `long:"xud.host" default:"localhost" description:"XUD gRPC service host"`
	Port int    `long:"xud.port" default:"8886" description:"XUD gRPC service port"`

	Certificate string `long:"xud.certificatepath" description:"Path to the certificate of the XUD gRPC interface"`

	Config string `long:"xud.configpath" description:"Path to the config file of XUD"`

	Ctx    context.Context
	Client xudrpc.XudClient
}

// OrderRemoved is a callback that allows clients to get notified about order removals
type OrderRemoved func(order xudrpc.OrderUpdate)

// OrderAdded is a callback that allows clients to get notified about added orders
type OrderAdded func(order xudrpc.OrderUpdate)

// Init to a XUD node
func (xud *Xud) Init() error {
	creds, err := credentials.NewClientTLSFromFile(xud.Certificate, "")

	if err != nil {
		return err
	}

	uri := xud.Host + ":" + strconv.Itoa(xud.Port)
	con, err := grpc.Dial(uri, grpc.WithTransportCredentials(creds))

	if err != nil {
		return err
	}

	if xud.Ctx == nil {
		xud.Ctx = context.Background()
	}

	xud.Client = xudrpc.NewXudClient(con)

	return nil
}

// GetInfo gets general information about the XUD node
func (xud *Xud) GetInfo() (*xudrpc.GetInfoResponse, error) {
	return xud.Client.GetInfo(xud.Ctx, &xudrpc.GetInfoRequest{})
}

// PlaceOrderSync places a new order in XUD
func (xud *Xud) PlaceOrderSync(request xudrpc.PlaceOrderRequest) (*xudrpc.PlaceOrderResponse, error) {
	return xud.Client.PlaceOrderSync(xud.Ctx, &request)
}

// SubscribeRemovedOrders notifies the Client via a callback about removed orders
func (xud *Xud) SubscribeRemovedOrders(callback OrderRemoved) error {
	stream, streamErr := xud.Client.SubscribeOrders(xud.Ctx, &xudrpc.SubscribeOrdersRequest{})

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

// SubscribeAddedOrders notifies the Client via a callback about removed orders
func (xud *Xud) SubscribeAddedOrders(callback OrderAdded) error {
	stream, streamErr := xud.Client.SubscribeOrders(xud.Ctx, &xudrpc.SubscribeOrdersRequest{})

	if streamErr != nil {
		return streamErr
	}

	//wait := make(chan struct{})

	go func() {
		for {
			order, err := stream.Recv()

			if err != nil {
				if err == io.EOF {
					err = errors.New("lost connection to XUD")
				}

				streamErr = err
				//close(wait)
				return
			}

			callback(*order)
		}
	}()

	//<-wait

	return streamErr
}
