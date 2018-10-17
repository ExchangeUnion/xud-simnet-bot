package slackclient

import (
	"github.com/nlopes/slack"
)

// Slack represents a Slack Client
type Slack struct {
	Token   string `long:"slack.token" description:"Slack OAuth token"`
	Channel string `long:"slack.channel" description:"Slack channel to which messages should be sent"`
	Prefix  string `long:"slack.prefix" description:"Prefix for every message"`

	api *slack.Client
}

// InitSlack initializes a new Slack client
func (slackClient *Slack) InitSlack() {
	slackClient.api = slack.New(slackClient.Token)
}

// SendMessage sends a message to the configured channel
func (slackClient *Slack) SendMessage(message string) {
	if slackClient.Prefix != "" {
		message = slackClient.Prefix + ": " + message
	}

	_, _, _, err := slackClient.api.SendMessage(slackClient.Channel, slack.MsgOptionText(message, false))

	if err != nil {
		log.Warning("Could not send message (%v) to Slack: %v", message, err)
	}
}
