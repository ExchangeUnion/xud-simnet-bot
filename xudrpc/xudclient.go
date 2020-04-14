package xudrpc

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"strconv"
)

type Xud struct {
	Host string `long:"xud.host" default:"127.0.0.1" description:"XUD gRPC service host"`
	Port int    `long:"xud.port" default:"28886" description:"XUD gRPC service port"`

	Certificate string `long:"xud.certificatepath" description:"Path to the certificate of the XUD gRPC interface"`

	ctx    context.Context
	client XudClient
}

func (xud *Xud) Init() error {
	creds, err := credentials.NewClientTLSFromFile(xud.Certificate, "")

	if err != nil {
		return err
	}

	con, err := grpc.Dial(xud.Host+":"+strconv.Itoa(xud.Port), grpc.WithTransportCredentials(creds))

	if err != nil {
		return err
	}

	if xud.ctx == nil {
		xud.ctx = context.Background()
	}

	xud.client = NewXudClient(con)
	return nil
}

func (xud *Xud) GetInfo() (*GetInfoResponse, error) {
	return xud.client.GetInfo(xud.ctx, &GetInfoRequest{})
}

func (xud *Xud) ListPeers() (*ListPeersResponse, error) {
	return xud.client.ListPeers(xud.ctx, &ListPeersRequest{})
}

func (xud *Xud) OpenChannel(request *OpenChannelRequest) (*OpenChannelResponse, error) {
	return xud.client.OpenChannel(xud.ctx, request)
}
