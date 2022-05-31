package controller

import (
	"git.diabin.com/BlockChain/qitmeer-presstest/config"
	"git.diabin.com/BlockChain/qitmeer-presstest/db"
	"git.diabin.com/BlockChain/qitmeer-presstest/evm"
	"git.diabin.com/BlockChain/qitmeer-presstest/presstest"
	"git.diabin.com/BlockChain/qitmeer-presstest/rpc"
	"github.com/Qitmeer/qitmeer/qx"
	"sync"
)

type Controller struct {
	storage   db.IDB
	rpcClient []*rpc.Client
	network   string
	tokens    map[string]bool
	press     *presstest.Engine
	sync.Mutex
	evm *evm.EVM
}

func NewController(conf *config.Config) (*Controller, error) {
	//storage, err := db.ConnectDB(conf)
	//if err != nil {
	//	return nil, err
	//}
	rpcs := []*rpc.Client{}
	for _, rpcCfg := range conf.Rpc {
		rpcs = append(rpcs, rpc.NewClient(&config.Rpc{
			Host:     rpcCfg.Host,
			Admin:    rpcCfg.Admin,
			Password: rpcCfg.Password,
		}))
	}

	// test rpc
	//tokens, err := rpcs[0].GetCoins()
	//if err != nil {
	//	return nil, err
	//}
	//for _, info := range tokens {
	//	coins.Coins[info.CoinName] = types.CoinID(info.CoinId)
	//}
	press, err := presstest.NewEngine(conf, nil)
	if err != nil {
		return nil, err
	}
	evm, err := evm.NewEVM(conf.Evm)
	if err != nil {
		return nil, err
	}
	return &Controller{storage: nil, rpcClient: rpcs, network: conf.Rpc[0].Net, press: press, evm: evm}, nil
}

func (c *Controller) Stop() {
	c.press.Stop()
}

func (c *Controller) SendPress(to, coin string, count, tps int) *presstest.Result {
	return c.press.Send(to, coin, count, tps)
}

func (c *Controller) Split(coin string, count int) *presstest.Result {
	return c.press.Split(coin, count)
}

func (c *Controller) TransferAll(from, key string, amount uint64) *presstest.Result {
	return c.press.TransferAll(from, key, amount)
}

func (c *Controller) Transfer(from, to, key string, amount uint64) *presstest.Result {
	return c.press.Transfer(from, to, key, amount)
}

func (c *Controller) MergerUTXO(from, key string) *presstest.Result {
	return c.press.MergerUTXO(from, key)
}

func (c *Controller) UnspentCount() *presstest.Result {
	return c.press.UnspentCount()
}

func (c *Controller) AddAddress(count int) *presstest.Result {
	var addresses = []string{}
	for i := 0; i < count; i++ {
		seed, err := qx.NewEntropy(32)
		if err != nil {
			return &presstest.Result{
				Success: len(addresses),
				Error:   err.Error(),
				Data:    addresses,
			}
		}
		ec, err := qx.EcNew("secp256k1", seed)
		if err != nil {
			return &presstest.Result{
				Success: len(addresses),
				Error:   err.Error(),
				Data:    addresses,
			}
		}
		pu, err := qx.EcPrivateKeyToEcPublicKey(false, ec)
		if err != nil {
			return &presstest.Result{
				Success: len(addresses),
				Error:   err.Error(),
				Data:    addresses,
			}
		}
		address, err := qx.EcPubKeyToAddress(c.network, pu)
		if err != nil {
			return &presstest.Result{
				Success: len(addresses),
				Error:   err.Error(),
				Data:    addresses,
			}
		}
		err = c.press.AddKey(address, ec)
		if err != nil {
			return &presstest.Result{
				Success: len(addresses),
				Error:   err.Error(),
				Data:    addresses,
			}
		}
		addresses = append(addresses, address)
	}

	return &presstest.Result{
		Success: len(addresses),
		Error:   "",
		Data:    addresses,
	}
}
