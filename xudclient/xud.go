package xudclient

import (
	"context"
	"strconv"

	"github.com/ExchangeUnion/xud-tests/xudrpc"
	"google.golang.org/grpc"
)

// Xud represents a XUD client
type Xud struct {
	GrpcHost string `long:"xud.host" default:"localhost" description:"XUD gRPC service host"`
	GrpcPort int    `long:"xud.port" default:"8886" description:"XUD gRPC service port"`

	ctx    context.Context
	client xudrpc.XudClient
}

// Connect to a XUD node
func (xud *Xud) Connect() error {
	uri := xud.GrpcHost + ":" + strconv.Itoa(xud.GrpcPort)

	con, err := grpc.Dial(uri, grpc.WithInsecure())

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
