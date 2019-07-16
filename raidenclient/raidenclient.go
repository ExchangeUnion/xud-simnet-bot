package raidenclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
)

// Raiden represents a client
type Raiden struct {
	Disable bool

	Host string
	Port int

	endpoint string
}

// Channel contains information about a Raiden channel and is used a response in multiple calls
type Channel struct {
	TokenNetworkIdentifer string  `json:"token_network_identifier"`
	ChannelIdentifier     uint    `json:"channel_identifier"`
	PartnerAddress        string  `json:"partner_address"`
	TokenAddress          string  `json:"token_address"`
	Balance               float64 `json:"balance"`
	TotalDeposit          float64 `json:"total_deposit"`
	State                 string  `json:"state"`
	SettleTimeout         uint    `json:"settle_timeout"`
	RevealTimeout         uint    `json:"reveal_timeout"`
}

// SendPaymentResponse is the reponse of the "SendPayment" call of Raiden
type SendPaymentResponse struct {
	InitiatorAddress string  `json:"initiator_address"`
	TargetAddress    string  `json:"target_address"`
	TokenAddress     string  `json:"token_address"`
	Amount           float64 `json:"amount"`
	Identifier       uint64  `json:"identifier"`
}

// RaidenError allow to parse errors the Raiden API returns
type RaidenError struct {
	Errors string `json:"errors"`
}

var openLock = &sync.Mutex{}
var closeLock = &sync.Mutex{}

// Init the Raiden node
func (raiden *Raiden) Init() {
	raiden.endpoint = "http://" + raiden.Host + ":" + strconv.Itoa(raiden.Port) + "/api/v1/"
}

// ListChannels of either all tokens or the one specified
func (raiden *Raiden) ListChannels(tokenAddress string) ([]Channel, error) {
	var response []Channel

	endpoint := "channels"

	if tokenAddress != "" {
		endpoint += "/" + tokenAddress
	}

	responseBody, err := raiden.makeHTTPRequest(
		http.MethodGet,
		endpoint,
		nil,
	)

	if err != nil {
		return response, err
	}

	err = json.Unmarshal(responseBody, &response)
	err = handleResponse(responseBody, err)

	return response, err
}

// ListTokens lists all registered tokens
func (raiden *Raiden) ListTokens() ([]string, error) {
	var response []string

	responseBody, err := raiden.makeHTTPRequest(
		http.MethodGet,
		"tokens",
		nil,
	)

	if err != nil {
		return response, err
	}

	err = json.Unmarshal(responseBody, &response)
	err = handleResponse(responseBody, err)

	return response, err
}

// SendPayment send coins to the target address
func (raiden *Raiden) SendPayment(targetAddress string, tokenAddress string, amount float64) (SendPaymentResponse, error) {
	var response SendPaymentResponse

	responseBody, err := raiden.makeHTTPRequest(
		http.MethodPost,
		"payments/"+tokenAddress+"/"+targetAddress,
		map[string]interface{}{
			"amount": amount,
		},
	)

	if err != nil {
		return response, err
	}

	err = json.Unmarshal(responseBody, &response)
	err = handleResponse(responseBody, err)

	return response, err
}

// OpenChannel opens a new channel
func (raiden *Raiden) OpenChannel(partnerAddress string, tokenAddress string, totalDeposit float64, settleTimeout uint64) (Channel, error) {
	openLock.Lock()
	defer openLock.Unlock()

	var response Channel

	responseBody, err := raiden.makeHTTPRequest(
		http.MethodPut,
		"channels",
		map[string]interface{}{
			"partner_address": partnerAddress,
			"token_address":   tokenAddress,
			"total_deposit":   totalDeposit,
			"settle_timeout":  settleTimeout,
		},
	)

	if err != nil {
		return response, err
	}

	err = json.Unmarshal(responseBody, &response)
	err = handleResponse(responseBody, err)

	return response, err
}

// CloseChannel closes a channel
func (raiden *Raiden) CloseChannel(partnerAddress string, tokenAddress string) (Channel, error) {
	closeLock.Lock()
	defer closeLock.Unlock()

	var response Channel

	responseBody, err := raiden.makeHTTPRequest(
		http.MethodPatch,
		"channels/"+tokenAddress+partnerAddress,
		map[string]interface{}{
			"state": "closed",
		},
	)

	if err != nil {
		return response, err
	}

	err = json.Unmarshal(responseBody, &response)
	err = handleResponse(responseBody, err)

	return response, err
}

func (raiden *Raiden) makeHTTPRequest(method string, endpoint string, requestBody map[string]interface{}) ([]byte, error) {
	httpClient := &http.Client{
		Timeout: 0,
	}

	url := raiden.endpoint + endpoint

	var response *http.Response
	var err error

	if method == http.MethodGet {
		response, err = httpClient.Get(url)
	} else {
		jsonRequestBody, _ := json.Marshal(requestBody)

		request, err := http.NewRequest(
			method,
			url,
			bytes.NewBuffer(jsonRequestBody),
		)

		if err != nil {
			return nil, err
		}

		response, err = httpClient.Do(request)
	}

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	return ioutil.ReadAll(response.Body)
}

func handleResponse(response []byte, err error) error {
	// If the parsing of the response fails -> return the whole body of the response as error
	if err != nil {
		return errors.New(string(response))
	}

	return nil
}
