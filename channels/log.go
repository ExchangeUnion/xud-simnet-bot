package channels

import "github.com/op/go-logging"

var log logging.Logger

// UseLogger tells the "channels" package which logger to use
func UseLogger(logger logging.Logger) {
	log = logger
}

func logCouldNotConnect(err error) {
	log.Warning("Could not connect to %v: %v", nodeName, err)
}
