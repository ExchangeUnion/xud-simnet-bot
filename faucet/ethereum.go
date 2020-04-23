package faucet

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/logger"
	"golang.org/x/crypto/sha3"
	"io/ioutil"
	"math/big"
	"sync"
)

var sendLock = &sync.Mutex{}

// 1 gwei is enough for our simnet
var gasPrice = big.NewInt(1000000000)

var ethTransferGasLimit = uint64(21000)
var erc20TransferGasLimit = uint64(50000)

type Ethereum struct {
	RPCHost      string `long:"eth.rpcuri" description:"URI of the RPC interface of an Ethereum client"`
	KeystorePath string `long:"eth.keystore" description:"Path to the keystore of the Ethereum address"`
	Password     string `long:"eth.password" description:"Password of the keystore"`

	chainID *big.Int
	client  *ethclient.Client

	ctx context.Context

	keystore *keystore.KeyStore
	account  accounts.Account
}

func (eth *Ethereum) Init() error {
	var err error
	eth.client, err = ethclient.Dial(eth.RPCHost)

	if err != nil {
		return err
	}

	if eth.ctx == nil {
		eth.ctx = context.Background()
	}

	eth.chainID, err = eth.client.NetworkID(eth.ctx)

	if err != nil {
		return err
	}

	rawKeyStore, err := ioutil.ReadFile(eth.KeystorePath)

	if err != nil {
		return err
	}

	eth.keystore = keystore.NewKeyStore("./tmpKeyStore", keystore.StandardScryptN, keystore.StandardScryptP)
	eth.account, err = eth.keystore.Import(rawKeyStore, eth.Password, eth.Password)

	if err != nil {
		return err
	}

	err = eth.keystore.Unlock(eth.account, eth.Password)

	if err != nil {
		return err
	}

	logger.Info("Initialized Ethereum client with address: " + eth.account.Address.String())

	return nil
}

func (eth *Ethereum) SendEther(address string, amount *big.Int) error {
	sendLock.Lock()
	defer sendLock.Unlock()

	nonce, err := eth.client.PendingNonceAt(eth.ctx, eth.account.Address)

	if err != nil {
		return err
	}

	recipient := common.HexToAddress(address)

	transaction := types.NewTransaction(nonce, recipient, amount, ethTransferGasLimit, gasPrice, nil)
	transaction, err = eth.keystore.SignTx(eth.account, transaction, eth.chainID)

	if err != nil {
		return err
	}

	return eth.client.SendTransaction(eth.ctx, transaction)
}

func (eth *Ethereum) SendToken(token string, address string, amount string) error {
	sendLock.Lock()
	defer sendLock.Unlock()

	nonce, err := eth.client.PendingNonceAt(eth.ctx, eth.account.Address)

	if err != nil {
		return err
	}

	tokenAddress := common.HexToAddress(token)
	recipient := common.HexToAddress(address)

	transferFunctionSignature := []byte("transfer(address,uint256)")

	hash := sha3.NewLegacyKeccak256()
	_, err = hash.Write(transferFunctionSignature)

	if err != nil {
		return err
	}

	tokenAmount := new(big.Int)
	tokenAmount.SetString(amount, 10)

	var data []byte

	data = append(data, hash.Sum(nil)[:4]...)
	data = append(data, common.LeftPadBytes(recipient.Bytes(), 32)...)
	data = append(data, common.LeftPadBytes(tokenAmount.Bytes(), 32)...)

	transaction := types.NewTransaction(nonce, tokenAddress, big.NewInt(0), erc20TransferGasLimit, gasPrice, data)
	transaction, err = eth.keystore.SignTx(eth.account, transaction, eth.chainID)

	if err != nil {
		return err
	}

	return eth.client.SendTransaction(eth.ctx, transaction)
}
