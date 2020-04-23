package faucet

import (
	"encoding/json"
	"github.com/ExchangeUnion/xud-simnet-bot/channels"
	"github.com/ExchangeUnion/xud-simnet-bot/discord"
	"github.com/google/logger"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"strings"
)

type Faucet struct {
	Port int `long:"faucet.port" description:"Port to which the HTTP server of the faucet will listen"`

	channels []channels.Channel

	eth     *Ethereum
	discord *discord.Discord
}

type faucetRequest struct {
	Address string `json:"address"`
}

type faucetResponse struct {
	TokensSent map[string]string `json:"tokensSent"`
}

type errorResponse struct {
	Error string `json:"error"`
}

var decimals = big.NewFloat(math.Pow(10, 18))

func (faucet *Faucet) Start(channels []channels.Channel, eth *Ethereum, discord *discord.Discord) {
	logger.Info("Starting faucet at port: " + strconv.Itoa(faucet.Port))

	var channelNames []string

	for _, channel := range channels {
		channelNames = append(channelNames, channel.Currency)
	}

	logger.Info("Faucet currencies: " + strings.Join(channelNames, ", "))


	faucet.channels = channels

	faucet.eth = eth
	faucet.discord = discord

	http.HandleFunc("/faucet", func(writer http.ResponseWriter, request *http.Request) {
		decoder := json.NewDecoder(request.Body)

		var resultBody faucetRequest
		err := decoder.Decode(&resultBody)

		if err != nil {
			writeResponse(writer, 400, errorResponse{
				Error: "could not parse request: " + err.Error(),
			})
			return
		}

		if resultBody.Address == "" {
			writeResponse(writer, 400, errorResponse{
				"no address was provided",
			})
			return
		}

		response, err := faucet.sendTokens(resultBody.Address)

		if err != nil {
			writeResponse(writer, 400, errorResponse{
				Error: "could not send tokens: " + err.Error(),
			})
			_ = discord.SendMessage("Could not send tokens: " + err.Error())

			return
		}

		writeResponse(writer, 200, response)

		_ = discord.SendMessage("Sent tokens to `" + resultBody.Address + "`")
	})

	err := http.ListenAndServe("0.0.0.0:"+strconv.Itoa(faucet.Port), nil)

	if err != nil {
		logger.Fatal("Could not start faucet: " + err.Error())
	}
}

func (faucet *Faucet) sendTokens(address string) (response faucetResponse, err error) {
	response.TokensSent = map[string]string{}

	for _, channel := range faucet.channels {
		amount := big.NewFloat(0.0)
		amount = amount.Mul(decimals, big.NewFloat(channel.Amount))

		stringAmount := big.NewInt(0)
		stringAmount, _ = amount.Int(stringAmount)

		response.TokensSent[channel.Currency] = stringAmount.String()

		if channel.TokenAddress != "" {
			err = faucet.eth.SendToken(channel.TokenAddress, address, stringAmount.String())
		} else if channel.Currency == "ETH" {
			err = faucet.eth.SendEther(address, stringAmount)
		}

		if err != nil {
			return response, err
		}
	}

	return response, err
}

func writeResponse(writer http.ResponseWriter, status int, data interface{}) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)

	_ = json.NewEncoder(writer).Encode(data)
}
