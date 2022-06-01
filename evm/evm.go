package evm

import (
	"context"
	"fmt"
	"git.diabin.com/BlockChain/qitmeer-presstest/abi"
	"git.diabin.com/BlockChain/qitmeer-presstest/config"
	"git.diabin.com/BlockChain/qitmeer-presstest/presstest"
	"github.com/bCoder778/log"
	"github.com/ethereum/go-ethereum"
	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/tyler-smith/go-bip39"
	"math/big"
	"strings"
	"sync"
	"time"
)

const CHAN_ID = 8132

var Coin = big.NewInt(0).Mul(big.NewInt(1e9), big.NewInt(1e9))

type Account struct {
	Address string
	Key     string
}

type EVM struct {
	evmClients []*ethclient.Client
	tokens     []*abi.Token
	accounts   []*Account
	token      string
}

func NewEVM(evmConf *config.EVM) (*EVM, error) {
	var evmClients []*ethclient.Client
	var tokens []*abi.Token
	var accounts []*Account
	for _, rpc := range evmConf.Rpc {
		client, err := ethclient.Dial(rpc)
		if err != nil {
			return nil, err
		}
		evmClients = append(evmClients, client)
		token, err := abi.NewToken(common.HexToAddress(evmConf.Token), client)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	for i := 0; i < evmConf.Account; i++ {
		account, err := NewAccount(evmConf.Mnemonic, i)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return &EVM{
		evmClients: evmClients,
		tokens:     tokens,
		token:      evmConf.Token,
		accounts:   accounts,
	}, nil
}

func NewAccount(mnemonic string, index int) (*Account, error) {
	seed := bip39.NewSeed(mnemonic, "")
	wallet, err := hdwallet.NewFromSeed(seed)
	if err != nil {
		return nil, err
	}
	path := hdwallet.MustParseDerivationPath(fmt.Sprintf("m/44'/60'/0'/0/%d", index)) //最后一位是同一个助记词的地址id，从0开始，相同助记词可以生产无限个地址
	account, err := wallet.Derive(path, false)
	if err != nil {
		return nil, err
	}
	address := account.Address.Hex()
	privateKey, _ := wallet.PrivateKeyHex(account)
	return &Account{
		Address: address,
		Key:     privateKey,
	}, nil
}

func (e *EVM) ShareMoney(amount string) bool {
	nonce, err := e.evmClients[0].PendingNonceAt(context.Background(), common.HexToAddress(e.accounts[0].Address))
	if err != nil {
		return false
	}
	for i := 1; i < len(e.accounts); i++ {
		amount, ok := big.NewInt(0).SetString(amount, 10)
		if !ok {
			log.Errorf("error amount %s", amount)
			return false
		}
		amount = big.NewInt(0).Mul(Coin, amount)
		if hash, err := e.TransferToken(e.accounts[0].Address, e.accounts[0].Key, e.accounts[i].Address, e.token, amount, nonce); err != nil {
			log.Errorf("TransferToken %s", err.Error())
			continue
		} else {
			log.Infof("TransferToken, %s", hash)
		}
		nonce++
		hash, err := e.TransferMeer(e.accounts[0].Address, e.accounts[0].Key, e.accounts[i].Address, big.NewInt(0).Mul(big.NewInt(1000000), big.NewInt(1000000000)), nonce)
		if err != nil {
			log.Errorf("TransferMeer %s", err.Error())
			continue
		} else {
			log.Infof("TransferMeer, %s", hash)
		}
		nonce++
	}
	return true
}

func (e *EVM) PressEVM(count, tps int) interface{} {
	nodeCount := len(e.tokens)
	avgTps := tps / nodeCount
	avgCount := 0
	wg := &sync.WaitGroup{}
	startTime := time.Now().Unix()
	for i, token := range e.tokens {
		wg.Add(1)
		if i == 0 {
			avgCount += count/nodeCount + count%nodeCount
		} else {
			avgCount = count / nodeCount
		}
		go e.sendTransferByTps(e.evmClients[i], token, e.accounts[i], avgTps, avgCount, wg, i)
	}
	wg.Wait()
	useSec := time.Now().Unix() - startTime
	log.Infof("txCount %d, tps = %.2f, times = %d", count, float64(count)/float64(useSec), useSec)
	return &presstest.Result{
		Success: count,
		Error:   "",
		Data:    float64(count) / float64(useSec),
	}
}

type Tx struct {
	EthTx *types.Transaction
	From  string
	Index int
}

func (e *EVM) PressEVMV2(count, tps int) interface{} {
	txCh := make(chan *Tx, count)
	contract := common.HexToAddress(e.token)
	interval := 1000 / tps
	index := 0
	for i, account := range e.accounts {
		avgCount := count / len(e.accounts)
		if i == 0 {
			avgCount += count % len(e.accounts)
		}
		privKey, err := crypto.HexToECDSA(account.Key)
		if err != nil {
			return err
		}
		txOut, err := bind.NewKeyedTransactorWithChainID(privKey, big.NewInt(CHAN_ID))
		if err != nil {
			return err
		}
		txOut.From = common.HexToAddress(account.Address)
		txOut.NoSend = false
		txOut.Value = FloatToWei(0)

		nonce, err := e.evmClients[0].PendingNonceAt(context.Background(), common.HexToAddress(account.Address))
		if err != nil {
			return err
		}

		for j := 0; j < avgCount; j++ {
			tx, err := e.CreateTx(e.evmClients[0], txOut, &contract, txOut.From, big.NewInt(1), nonce)
			if err != nil {
				return err
			}
			txCh <- &Tx{
				EthTx: tx,
				From:  account.Address,
				Index: index,
			}
			nonce += 1
			index++
			log.Infof("CreateTx %d", index)
		}
	}
	wg := &sync.WaitGroup{}
	startTime := time.Now().Unix()
	for i, client := range e.evmClients {
		wg.Add(1)
		go e.sendTransfer(client, txCh, wg, interval, i)
	}
	wg.Wait()
	useSec := time.Now().Unix() - startTime
	log.Infof("txCount %d, tps = %.2f, times = %d", count, float64(count)/float64(useSec), useSec)
	return &presstest.Result{
		Success: count,
		Error:   "",
		Data:    float64(count) / float64(useSec),
	}
}

func (e *EVM) sendTransfer(client *ethclient.Client, txCh chan *Tx, wg *sync.WaitGroup, interval int, index int) {
	defer wg.Done()
	for {
		select {
		case tx := <-txCh:
			err := client.SendTransaction(context.Background(), tx.EthTx)
			if err != nil {
				log.Errorf("%d send %s address %s nonce %d, %s", index, tx.EthTx.Hash().String(), tx.From, tx.EthTx.Nonce(), err.Error())
			} else {
				log.Infof("send %d transfer %s", tx.Index, tx.EthTx.Hash().String())
			}
			if len(txCh) == 0 {
				return
			}
			time.Sleep(time.Duration(interval) * time.Millisecond)
		}
	}
}

func (e *EVM) sendTransferByTps(client *ethclient.Client, token *abi.Token, account *Account, avgTps, avgCount int, wg *sync.WaitGroup, index int) error {
	defer wg.Done()
	nonce, err := client.PendingNonceAt(context.Background(), common.HexToAddress(account.Address))
	if err != nil {
		return err
	}
	privKey, err := crypto.HexToECDSA(account.Key)
	if err != nil {
		return err
	}
	txOut, err := bind.NewKeyedTransactorWithChainID(privKey, big.NewInt(CHAN_ID))
	if err != nil {
		return err
	}
	txOut.From = common.HexToAddress(account.Address)
	txOut.NoSend = false
	txOut.Nonce = big.NewInt(int64(nonce))
	txOut.GasPrice = big.NewInt(DefaultGasPrice)
	txOut.Value = FloatToWei(0)
	interval := 1000 / avgTps

	for i := 0; i < avgCount; i++ {
		tx, err := token.Transfer(txOut, common.HexToAddress(account.Address), big.NewInt(1))
		if err != nil {
			log.Errorf("send transfer %s", err.Error())
			continue
		}
		log.Infof("%d send %d transfer %s", index, i, tx.Hash().String())
		txOut.Nonce = big.NewInt(0).Add(txOut.Nonce, big.NewInt(1))
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
	return nil
}

func (e *EVM) CreateTx(client *ethclient.Client, opts *bind.TransactOpts, contract *common.Address, to common.Address,
	amount *big.Int, nonce uint64) (*types.Transaction, error) {

	input, err := e.parseInput("transfer", to, amount)
	if err != nil {
		return nil, err
	}
	msg := ethereum.CallMsg{
		From:     opts.From,
		To:       contract,
		GasPrice: big.NewInt(DefaultGasPrice),
		Value:    big.NewInt(0),
		Data:     input,
	}
	//gasLimit, err := client.EstimateGas(context.Background(), msg)
	//if err != nil {
	//	return nil, err
	//}
	baseTx := &types.LegacyTx{
		To:       contract,
		Nonce:    nonce,
		GasPrice: msg.GasPrice,
		Gas:      30000,
		Value:    msg.Value,
		Data:     input,
	}
	signedTx, err := opts.Signer(opts.From, types.NewTx(baseTx))
	if err != nil {
		return nil, err
	}
	return signedTx, nil
}

func (e *EVM) parseInput(method string, params ...interface{}) ([]byte, error) {
	parsed, err := ethabi.JSON(strings.NewReader(abi.TokenABI))
	if err != nil {
		return nil, err
	}
	return parsed.Pack(method, params...)
}

func (e *EVM) TransferToken(from, key, to, token string, amount *big.Int, nonce uint64) (string, error) {

	privKey, err := crypto.HexToECDSA(key)
	if err != nil {
		return "", err
	}
	txOut, err := bind.NewKeyedTransactorWithChainID(privKey, big.NewInt(CHAN_ID))
	if err != nil {
		return "", err
	}
	txOut.From = common.HexToAddress(from)
	txOut.NoSend = false
	txOut.Nonce = big.NewInt(int64(nonce))
	txOut.GasPrice = big.NewInt(DefaultGasPrice)
	txOut.Value = FloatToWei(0)
	tx, err := e.tokens[0].Transfer(txOut, common.HexToAddress(to), amount)
	if err != nil {
		return "", err
	}
	return tx.Hash().String(), nil
}

func (e *EVM) TransferMeer(from, key, to string, amount *big.Int, nonce uint64) (string, error) {
	//nonce, err := e.evmClients[0].PendingNonceAt(context.Background(), common.HexToAddress(from))
	//if err != nil {
	//	return "", err
	//}
	privKey, err := crypto.HexToECDSA(key)
	if err != nil {
		return "", err
	}
	txOut, err := bind.NewKeyedTransactorWithChainID(privKey, big.NewInt(CHAN_ID))
	if err != nil {
		return "", err
	}
	txOut.From = common.HexToAddress(from)
	toAddr := common.HexToAddress(to)
	msg := ethereum.CallMsg{
		From:     txOut.From,
		GasPrice: big.NewInt(DefaultGasPrice),
		Value:    amount,
	}
	gasLimit, err := e.evmClients[0].EstimateGas(context.Background(), msg)
	if err != nil {
		return "", fmt.Errorf("failed to estimate gas needed: %v", err)
	}
	tx := types.NewTx(&types.LegacyTx{
		To:       &toAddr,
		Nonce:    nonce,
		Gas:      gasLimit,
		GasPrice: big.NewInt(DefaultGasPrice),
		Value:    amount,
	})

	signedTx, err := txOut.Signer(txOut.From, tx)
	err = e.evmClients[0].SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}
	return tx.Hash().String(), nil
}

func (e *EVM) TokenBalance(address string) (*big.Int, error) {
	return e.tokens[0].BalanceOf(nil, common.HexToAddress(address))
}

const DefaultGasPrice = 6100000000
const Wei = 1e18
const Ubi = 1e8

func FloatToWei(value float64) *big.Int {
	return big.NewInt(0).Mul(big.NewInt(int64(value*Ubi)), big.NewInt(Wei/Ubi))
}
