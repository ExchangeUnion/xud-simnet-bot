package ethclient

import (
	"context"
	"io/ioutil"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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
	RPCHost      string `long:"eth.rpchost" description:"Host of the RPC interface of an Ethereum client"`
	KeystorePath string `long:"eth.keystore" description:"Path to the keystore of the Ethereum address"`
	Password     string `long:"eth.password" description:"Password of the keystore"`

	client  *geth.Client
	chainID *big.Int

	keystore *keystore.KeyStore
	account  accounts.Account
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

	eth.chainID = chainID

	// Load the keystore from the file
	rawKeyStore, err := ioutil.ReadFile(eth.KeystorePath)

	if err != nil {
		return err
	}

	keystore := keystore.NewKeyStore("./tmp", keystore.StandardScryptN, keystore.StandardScryptP)
	account, err := keystore.Import(rawKeyStore, eth.Password, eth.Password)

	keystore.Unlock(account, eth.Password)

	if err != nil {
		return err
	}

	eth.keystore = keystore
	eth.account = account

	return nil
}

// SendEth sends a specific amount of Ether to a given address
func (eth *Ethereum) SendEth(address string, amount *big.Int) error {
	sendLock.Lock()

	nonce, err := eth.client.PendingNonceAt(context.Background(), eth.account.Address)

	if err != nil {
		sendLock.Unlock()
		return err
	}

	toAddress := common.HexToAddress(address)

	tx := types.NewTransaction(nonce, toAddress, amount, ethTransferGasLimit, gasPrice, nil)
	tx, err = eth.keystore.SignTx(eth.account, tx, eth.chainID)

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

	nonce, err := eth.client.PendingNonceAt(context.Background(), eth.account.Address)

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
	tx, err = eth.keystore.SignTx(eth.account, tx, eth.chainID)

	if err != nil {
		sendLock.Unlock()
		return nil, err
	}

	err = eth.client.SendTransaction(context.Background(), tx)

	sendLock.Unlock()

	return nil, err
}
