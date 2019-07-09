package ethclient

import (
	"context"
	"crypto/ecdsa"
	"io/ioutil"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	geth "github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/crypto/sha3"
)

var sendLock = &sync.Mutex{}

// 1 gwei is enough for our simnet
var gasPrice = big.NewInt(1000000000)

var ethTransferGasLimit = uint64(21000)
var erc20TransferGasLimit = uint64(50000)

// Ethereum represents an Ethereum client
type Ethereum struct {
	RPCHost        string `long:"eth.rpchost" description:"Host of the RPC interface of an Ethereum client"`
	PrivateKeyPath string `long:"eth.privkey" description:"Path to the private key of a Ethereum address"`

	client        *geth.Client
	chainIDSigner types.EIP155Signer

	privateKey *ecdsa.PrivateKey
	address    common.Address
}

// Init initializes a new Ethereum client
func (eth *Ethereum) Init() error {
	// Initialize the RPC client
	client, err := geth.Dial(eth.RPCHost)

	if err != nil {
		return err
	}

	eth.client = client

	chainID, err := client.NetworkID(context.Background())

	if err != nil {
		return err
	}

	eth.chainIDSigner = types.NewEIP155Signer(chainID)

	// Load the private key from the file
	rawPrivateKey, err := ioutil.ReadFile(eth.PrivateKeyPath)

	if err != nil {
		return err
	}

	stringPrivateKey := strings.TrimSuffix(string(rawPrivateKey), "\n")
	privateKey, err := crypto.HexToECDSA(stringPrivateKey)

	if err != nil {
		return err
	}

	eth.privateKey = privateKey

	// Get the address of the private key
	publicKey := privateKey.Public()
	publicKeyEcdsa := publicKey.(*ecdsa.PublicKey)

	eth.address = crypto.PubkeyToAddress(*publicKeyEcdsa)

	return nil
}

// SendEth sends a specific amount of Ether to a given address
func (eth *Ethereum) SendEth(address string, amount *big.Int) error {
	sendLock.Lock()

	nonce, err := eth.client.PendingNonceAt(context.Background(), eth.address)

	if err != nil {
		sendLock.Unlock()
		return err
	}

	toAddress := common.HexToAddress(address)

	tx := types.NewTransaction(nonce, toAddress, amount, ethTransferGasLimit, gasPrice, nil)
	tx, err = types.SignTx(tx, eth.chainIDSigner, eth.privateKey)

	if err != nil {
		sendLock.Unlock()
		return err
	}

	err = eth.client.SendTransaction(context.Background(), tx)

	sendLock.Unlock()

	return err
}

// SendToken send a specific amount of a token to a given address
func (eth *Ethereum) SendToken(token string, recipient string, amount string) (*types.Transaction, error) {
	sendLock.Lock()

	nonce, err := eth.client.PendingNonceAt(context.Background(), eth.address)

	if err != nil {
		sendLock.Unlock()
		return nil, err
	}

	tokenAddress := common.HexToAddress(token)
	recipientAddress := common.HexToAddress(recipient)

	transferFunctionSignature := []byte("transfer(address,uint256)")

	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFunctionSignature)

	methodID := hash.Sum(nil)[:4]
	paddedAddress := common.LeftPadBytes(recipientAddress.Bytes(), 32)

	tokenAmount := new(big.Int)
	tokenAmount.SetString(amount, 10)

	paddedAmount := common.LeftPadBytes(tokenAmount.Bytes(), 32)

	var data []byte

	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	tx := types.NewTransaction(nonce, tokenAddress, big.NewInt(0), erc20TransferGasLimit, gasPrice, data)
	tx, err = types.SignTx(tx, eth.chainIDSigner, eth.privateKey)

	if err != nil {
		sendLock.Unlock()
		return nil, err
	}

	err = eth.client.SendTransaction(context.Background(), tx)

	sendLock.Unlock()

	return nil, err
}
