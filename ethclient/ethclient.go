package ethclient

import (
	"context"
	"crypto/ecdsa"
	"io/ioutil"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	geth "github.com/ethereum/go-ethereum/ethclient"
)

// 1 gwei is enough for our simnet
var gasPrice = big.NewInt(1000000000)

var ethTransferGasLimit = uint64(21000)

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

// SendEth sends a specific amount of Ethereum to a given address
func (eth *Ethereum) SendEth(address string, amount *big.Int) error {
	nonce, err := eth.client.PendingNonceAt(context.Background(), eth.address)

	if err != nil {
		return err
	}

	toAddress := common.HexToAddress(address)

	tx := types.NewTransaction(nonce, toAddress, amount, ethTransferGasLimit, gasPrice, nil)
	tx, err = types.SignTx(tx, eth.chainIDSigner, eth.privateKey)

	if err != nil {
		return err
	}

	err = eth.client.SendTransaction(context.Background(), tx)

	return err
}
