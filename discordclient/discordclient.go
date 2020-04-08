package discordclient

import (
	"github.com/bwmarrin/discordgo"
)

// Discord represents a Discord client
type Discord struct {
	Token     string `long:"discord.token" description:"Discord authentication token"`
	ChannelID string `long:"discord.channelid" description:"ID of the channel to which messages should be sent"`
	Prefix    string `long:"discord.prefix" description:"Prefix for every message"`

	api *discordgo.Session
}

// Init initializes a new Discord client
func (discord *Discord) Init() (err error) {
	discord.api, err = discordgo.New("Bot " + discord.Token)

	if err != nil {
		return err
	}

	err = discord.api.Open()

	return err
}

// SendMessage sends a message to the Discord channel
func (discord *Discord) SendMessage(message string) {
	if discord.Prefix != "" {
		message = discord.Prefix + ": " + message
	}

	_, err := discord.api.ChannelMessageSend(discord.ChannelID, message)

	if err != nil {
		log.Warning("Could not send (%v) to Discord: %v", message, err.Error())
	}
}
