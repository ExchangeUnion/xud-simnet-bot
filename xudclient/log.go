package xudclient

import "github.com/op/go-logging"

var log logging.Logger

// UseLogger tells the "xudclient" which logger to use
func UseLogger(logger logging.Logger) {
	log = logger
}
