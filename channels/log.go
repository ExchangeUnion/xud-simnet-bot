package channels

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/op/go-logging"
)

var log logging.Logger

// UseLogger tells the "channels" package which logger to use
func UseLogger(logger logging.Logger) {
	log = logger
}

func logConnected(info lnrpc.GetInfoResponse) {
	log.Info("Connected to %v node %v version %v", nodeName, info.IdentityPubkey, info.Version)
}

func logCouldNotConnect(err error) {
	log.Warning("Could not connect to %v: %v", nodeName, err)
}
