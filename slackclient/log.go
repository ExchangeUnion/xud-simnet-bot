package slackclient

import "github.com/op/go-logging"

var log logging.Logger

// UseLogger tells the "slackclient" package which logger to use
func UseLogger(logger logging.Logger) {
	log = logger
}
