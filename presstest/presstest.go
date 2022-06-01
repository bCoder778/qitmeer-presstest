package presstest

import (
	"context"
	"encoding/hex"
	"fmt"
	"git.diabin.com/BlockChain/qitmeer-presstest/coins"
	"git.diabin.com/BlockChain/qitmeer-presstest/config"
	"git.diabin.com/BlockChain/qitmeer-presstest/db"
	"git.diabin.com/BlockChain/qitmeer-presstest/keystore"
	"git.diabin.com/BlockChain/qitmeer-presstest/rpc"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/address"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/qx"
	"github.com/bCoder778/log"
	stypes "github.com/bCoder778/qitmeer-sync/storage/types"
	"sync"
	"time"
)

type Engine struct {
	ctx    context.Context
	cancel context.CancelFunc
	db     db.IDB
	key    *keystore.Keystore
	cfg    *config.Config
	rpcs   []*rpc.Client
}

func NewEngine(cfg *config.Config, idb db.IDB) (*Engine, error) {
	//key, err := keystore.LoadKey(cfg.Key.Dir)
	//if err != nil {
	//	return nil, err
	//}
	rpcs := []*rpc.Client{}
	for _, rpcCfg := range cfg.Rpc {
		rpcs = append(rpcs, rpc.NewClient(&config.Rpc{
			Host:     rpcCfg.Host,
			Admin:    rpcCfg.Admin,
			Password: rpcCfg.Password,
		}))
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Engine{key: nil, db: idb, cfg: cfg, ctx: ctx, cancel: cancel, rpcs: rpcs}, nil
}

func (e *Engine) Stop() {
	e.cancel()
}

type Result struct {
	Success int
	Error   string
	Data    interface{}
}

func (e *Engine) AddKey(address, key string) error {
	return e.key.AddKey(address, key)
}

func (e *Engine) Send(to, coin string, count int, tps int) *Result {
	if coin == "" {
		coin = "MEER"
	}
	var amount uint64 = 10000
	coinId, exist := coins.Coins[coin]
	if !exist {
		return &Result{
			Success: 0,
			Error:   fmt.Errorf("no coin %s", coin).Error(),
		}
	}
	utxos := []stypes.Vout{}
	for _, address := range e.key.AddressList() {
		utxoList := e.db.QueryAmountUTXO(address, coin, amount)
		utxos = append(utxos, utxoList...)
		if len(utxos) >= count {
			utxos = utxos[0:count]
			break
		}
	}
	log.Debugf("utxo length %d", len(utxos))
	log.Debugf("utxo length %d", len(utxos))
	log.Debugf("utxo length %d", len(utxos))
	length := len(utxos)

	clientCount := len(e.rpcs)
	oneCount := length / clientCount
	var oneTps int
	if tps == 0 {
		oneTps = 0
	} else {
		oneTps = tps / clientCount
		if oneTps < 1 {
			oneTps = 1
		}
	}

	start := 0
	wg := &sync.WaitGroup{}
	var allSuccess uint64 = 0
	successList := []*uint64{}
	startTime := time.Now().Unix()
	for times := 0; times < clientCount; times++ {
		var oneSuccess uint64
		oneUtxos := []stypes.Vout{}
		if times == clientCount-1 {
			oneUtxos = utxos[start:]
		} else {
			oneUtxos = utxos[start : start+oneCount]
			start += oneCount
		}
		wg.Add(1)
		go e.send(oneUtxos, e.rpcs[times], &oneSuccess, coinId, oneTps, wg)
		successList = append(successList, &oneSuccess)
	}
	wg.Wait()
	endTime := time.Now().Unix()
	times := float64(endTime) - float64(startTime)
	if times == 0 {
		times = 1
	}
	for i, successCount := range successList {
		log.Infof("%d success %d, tps = %.2f, times = %f", i, *successCount, float64(*successCount)/times, times)
		allSuccess += *successCount
	}
	log.Infof("allSuccess %d, tps = %.2f, times = %f", allSuccess, float64(allSuccess)/times, times)

	return &Result{
		Success: int(allSuccess),
		Error:   "",
		Data:    float64(allSuccess) / times,
	}
}

func (e *Engine) send(utxos []stypes.Vout, client *rpc.Client, success *uint64, coinId types.CoinID, tps int, group *sync.WaitGroup) {
	defer group.Done()

	var amount uint64 = 10000
	start := time.Now().Unix()
	for i := 0; i < len(utxos)-1; i += 2 {
		select {
		case <-e.ctx.Done():
			stop := time.Now().Unix()
			log.Infof("stop send press %.2f", float64(*success)/float64(stop-start))
			return
		default:
			outputs := []Amount{}
			inputs := []qx.Input{}
			pks := []string{}
			key := []string{}
			total := utxos[i].Amount
			total += utxos[i+1].Amount
			inputs = append(inputs, qx.Input{
				TxID:     utxos[i].TxId,
				OutIndex: uint32(utxos[i].Number),
			})
			pks = append(pks, utxos[i].ScriptPubKey.Hex)
			key = append(key, e.key.Key(utxos[i].Address))

			inputs = append(inputs, qx.Input{
				TxID:     utxos[i+1].TxId,
				OutIndex: uint32(utxos[i+1].Number),
			})
			pks = append(pks, utxos[i+1].ScriptPubKey.Hex)
			key = append(key, e.key.Key(utxos[i+1].Address))

			outputs = append(outputs, Amount{
				TargetLockTime: 0,
				Value:          int64(amount),
				Id:             coinId,
				Address:        utxos[i+1].Address,
			})

			now := time.Now()
			tx, err := TxEncode(1, 0, &now, inputs, outputs)
			if err != nil {
				continue
			}
			signed, err := qx.TxSign(key, tx, e.cfg.Rpc[0].Net, pks)
			if err != nil {
				continue
			}
			go func() {
				txSigned := signed
				txid, err := client.SendTransaction(txSigned)
				if err != nil {
					log.Warnf("Send transaction %s", err.Error())
					return
				}
				log.Debugf("Send transaction %s", txid)
				*success++
			}()

			if tps != 0 {
				interval := 1000 / tps
				time.Sleep(time.Duration(interval) * time.Millisecond)
			}
		}

	}
}

func (e *Engine) Split(coin string, count int) *Result {
	if coin == "" {
		coin = "MEER"
	}
	oneCount := 500
	success := 0
	times := count / oneCount
	left := count % oneCount
	if left != 0 {
		times += 1
	}
	for h := 0; h < times; h++ {
		for _, address := range e.key.AddressList() {

			timeCount := oneCount
			if timeCount == times-1 {
				timeCount = left
			}
			var preFees uint64 = 100000000
			var oneAmount uint64 = 10000
			var amount uint64 = oneAmount * uint64(timeCount)
			coinId, exist := coins.Coins[coin]
			if !exist {
				return &Result{
					Success: success,
					Error:   fmt.Errorf("no coin %s", coin).Error(),
				}
			}
			var total uint64
			utxos := e.db.QuerySplitUTXO(address, coin)
			outputs := []Amount{}
			inputs := []qx.Input{}
			pks := []string{}
			vins := []stypes.Vout{}
			for _, uxto := range utxos {
				total += uint64(uxto.Amount)
				inputs = append(inputs, qx.Input{
					TxID:     uxto.TxId,
					OutIndex: uint32(uxto.Number),
				})
				vins = append(vins, uxto)
				pks = append(pks, uxto.ScriptPubKey.Hex)
				if total >= amount+preFees {
					break
				}
			}
			for i := 0; i < timeCount; i++ {
				outputs = append(outputs, Amount{
					TargetLockTime: 0,
					Value:          int64(oneAmount),
					Id:             coinId,
					Address:        address,
				})
			}
			fess := uint64(len(inputs)*1500+(len(outputs)+1)*450+2000) * 2
			if total < amount+fess {
				log.Warnf("%s total %d no enough, need %d", address, total, amount+fess)
				continue
			}
			change := total - (amount + fess)
			outputs = append(outputs, Amount{
				TargetLockTime: 0,
				Value:          int64(change),
				Id:             coinId,
				Address:        address,
			})

			now := time.Now()
			tx, err := TxEncode(1, uint32(988), &now, inputs, outputs)
			if err != nil {
				log.Warnf("TxEncode %s", err.Error())
				continue
			}
			signed, err := qx.TxSign([]string{e.key.Key(address)}, tx, e.cfg.Rpc[0].Net, pks)
			if err != nil {
				log.Warnf("TxSign %s", err.Error())
				continue
			}
			txid, err := e.rpcs[0].SendTransaction(signed)
			if err != nil {
				log.Warnf("send %s", err.Error())
				continue
			}

			log.Debugf("Send transaction %s", txid)
			success += timeCount
			e.db.UpdateSpent(vins, txid)

		}

	}

	return &Result{
		Success: success,
		Error:   "",
	}
}

type Amount struct {
	TargetLockTime int64
	Value          int64
	Id             types.CoinID
	Address        string
}

func TxEncode(version uint32, lockTime uint32, timestamp *time.Time, inputs []qx.Input, outputs []Amount) (string, error) {
	mtx := types.NewTransaction()
	mtx.Version = uint32(version)
	if lockTime != 0 {
		mtx.LockTime = uint32(lockTime)
	}
	if timestamp != nil {
		mtx.Timestamp = *timestamp
	}

	for _, vout := range inputs {
		txHash, err := hash.NewHashFromStr(vout.TxID)
		if err != nil {
			return "", err
		}
		prevOut := types.NewOutPoint(txHash, vout.OutIndex)
		txIn := types.NewTxInput(prevOut, []byte{})
		if lockTime != 0 {
			txIn.Sequence = types.MaxTxInSequenceNum - 1
		}
		mtx.AddTxIn(txIn)
	}

	/*outputsSlice := []string{}
	for _, k := range outputs {
		outputsSlice = append(outputsSlice, k.Address)
	}
	sort.Strings(outputsSlice)*/

	for _, out := range outputs {
		amount := out
		if amount.Value <= 0 || amount.Value > types.MaxAmount {
			return "", fmt.Errorf("invalid amount: 0 >= %v "+
				"> %v", amount, types.MaxAmount)
		}

		addr, err := address.DecodeAddress(out.Address)
		if err != nil {
			return "", fmt.Errorf("could not decode "+
				"address: %v", err)
		}

		switch addr.(type) {
		case *address.PubKeyHashAddress:
		case *address.ScriptHashAddress:
		default:
			return "", fmt.Errorf("invalid type: %T", addr)
		}

		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return "", err
		}
		if amount.TargetLockTime > 0 {
			pkScript, err = txscript.PayToCLTVPubKeyHashScript(addr.Script(), amount.TargetLockTime)
			if err != nil {
				return "", err
			}
		}
		txOut := types.NewTxOutput(types.Amount{Value: amount.Value, Id: amount.Id}, pkScript)
		mtx.AddTxOut(txOut)
	}
	mtxHex, err := mtx.Serialize()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(mtxHex), nil
}

func (e *Engine) UnspentCount() *Result {
	var usableCount int64 = 0
	var allCount int64 = 0
	addressList := e.key.AddressList()
	var amount uint64 = 10000
	for _, address := range addressList {
		UTXO := e.db.QueryAmountUTXOCount(address, "MEER", amount)
		usableCount += UTXO
		allUTXTO := e.db.QueryAmountAllUTXOCount(address, "MEER", amount)
		allCount += allUTXTO
	}
	return &Result{
		Success: 0,
		Error:   "",
		Data: map[string]int64{
			"usableCount": usableCount,
			"allCount":    allCount,
		},
	}
}

func (e *Engine) TransferAll(from, key string, amount uint64) *Result {
	addressList := e.key.AddressList()
	addrCount := len(addressList)

	oneCount := 500
	balance := e.AddrBalance(from, "MEER")

	coinId, exist := coins.Coins["MEER"]
	if !exist {
		return &Result{
			Success: 0,
			Error:   fmt.Errorf("no coin %s", "MEER").Error(),
		}
	}
	successIndex := 0
	times := addrCount / oneCount
	left := addrCount % oneCount
	if left != 0 {
		times += 1
	}
	addressIndex := 0
	for t := 0; t < times; t++ {
		if t == times-1 && left != 0 {
			oneCount = left
		}
		sum := amount * uint64(oneCount)
		var fees uint64 = 100000000
		if balance > sum+fees {
			outs, _ := e.GetAddressBigUTXO(from, "MEER", sum, fees)

			outputs := []Amount{}
			inputs := []qx.Input{}
			pks := []string{}
			vins := []stypes.Vout{}
			var total uint64
			for _, uxto := range outs {
				total += uint64(uxto.Amount * 1e8)
				inputs = append(inputs, qx.Input{
					TxID:     uxto.Txid,
					OutIndex: uint32(uxto.Number),
				})
				vins = append(vins, stypes.Vout{
					TxId:   uxto.Txid,
					Number: uxto.Number,
				})
				pks = append(pks, uxto.PkHex)
			}

			for i := 0; i < oneCount; i++ {
				outputs = append(outputs, Amount{
					TargetLockTime: 0,
					Value:          int64(amount),
					Id:             coinId,
					Address:        addressList[addressIndex],
				})
				addressIndex++
			}

			fess := uint64(len(inputs)*1500+(len(outputs)+1)*450+2000) * 2
			if total < amount+fess {
				log.Warnf("%s total %d no enough, need %d", from, total, amount+fess)
				continue
			}
			change := total - (sum + fess)
			outputs = append(outputs, Amount{
				TargetLockTime: 0,
				Value:          int64(change),
				Id:             coinId,
				Address:        from,
			})

			now := time.Now()
			tx, err := TxEncode(1, uint32(988), &now, inputs, outputs)
			if err != nil {
				log.Warnf("TxEncode %s", err.Error())
				continue
			}
			signed, err := qx.TxSign([]string{key}, tx, e.cfg.Rpc[0].Net, pks)
			if err != nil {
				log.Warnf("TxSign %s", err.Error())
				continue
			}
			txid, err := e.rpcs[0].SendTransaction(signed)
			if err != nil {
				log.Warnf("send %s", err.Error())
				break
			}
			successIndex = addressIndex
			log.Debugf("Send transaction %s", txid)
			e.db.UpdateSpent(vins, txid)

		} else {
			return &Result{
				Success: 0,
				Error:   "balance no enough",
				Data:    nil,
			}
		}
	}
	return &Result{
		Success: successIndex,
		Error:   "",
		Data:    addressList[:successIndex],
	}
}

func (e *Engine) MergerUTXO(from, key string) *Result {

	coinId, exist := coins.Coins["MEER"]
	if !exist {
		return &Result{
			Success: 0,
			Error:   fmt.Errorf("no coin %s", "MEER").Error(),
		}
	}
	allOuts := e.db.QueryAllUTXO(from, "MEER")
	var all uint64
	var allFees uint64
	for i := 0; i < len(allOuts); i += 500 {
		count := 500
		if i > len(allOuts)-1 {
			count = len(allOuts)
		}
		outs := allOuts[i : i+count]
		outputs := []Amount{}
		inputs := []qx.Input{}
		pks := []string{}
		vins := []stypes.Vout{}
		var total uint64
		for _, uxto := range outs {
			if uxto.Amount == 0 {
				continue
			}
			total += uint64(uxto.Amount)
			inputs = append(inputs, qx.Input{
				TxID:     uxto.TxId,
				OutIndex: uint32(uxto.Number),
			})
			vins = append(vins, stypes.Vout{
				TxId:   uxto.TxId,
				Number: uxto.Number,
			})
			pks = append(pks, uxto.ScriptPubKey.Hex)
		}

		fess := uint64(len(inputs)*1500+(len(outputs)+1)*450+2000) * 2
		if total < fess {
			err := fmt.Errorf("%s total %d no enough, need %d", from, total, fess)
			return &Result{
				Success: 0,
				Error:   err.Error(),
				Data:    nil,
			}
		}
		change := total - fess
		outputs = append(outputs, Amount{
			TargetLockTime: 0,
			Value:          int64(change),
			Id:             coinId,
			Address:        from,
		})

		now := time.Now()
		tx, err := TxEncode(1, uint32(90818), &now, inputs, outputs)
		if err != nil {
			log.Warnf("TxEncode %s", err.Error())

		}
		signed, err := qx.TxSign([]string{key}, tx, e.cfg.Rpc[0].Net, pks)
		if err != nil {
			return &Result{
				Success: 0,
				Error:   err.Error(),
				Data:    nil,
			}
		}
		txid, err := e.rpcs[0].SendTransaction(signed)
		if err != nil {
			log.Warnf("send %s", err.Error())
			return &Result{
				Success: 0,
				Error:   err.Error(),
				Data: map[string]uint64{
					"all":     all,
					"allFees": allFees,
				},
			}
		}
		log.Infof("%s", txid)
		allFees += fess
		all += change
	}
	return &Result{
		Success: 0,
		Error:   "no enough",
		Data: map[string]uint64{
			"all":     all,
			"allFees": allFees,
		},
	}
}

func (e *Engine) Transfer(from, to, key string, amount uint64) *Result {

	balance := e.AddrBalance(from, "MEER")

	coinId, exist := coins.Coins["MEER"]
	if !exist {
		return &Result{
			Success: 0,
			Error:   fmt.Errorf("no coin %s", "MEER").Error(),
		}
	}
	sum := amount
	var fees uint64 = 100000000
	if balance > sum+fees {
		outs, _ := e.GetAddressUTXO(from, "MEER", sum, fees)

		outputs := []Amount{}
		inputs := []qx.Input{}
		pks := []string{}
		vins := []stypes.Vout{}
		var total uint64
		for _, uxto := range outs {
			total += uint64(uxto.Amount * 1e8)
			inputs = append(inputs, qx.Input{
				TxID:     uxto.Txid,
				OutIndex: uint32(uxto.Number),
			})
			vins = append(vins, stypes.Vout{
				TxId:   uxto.Txid,
				Number: uxto.Number,
			})
			pks = append(pks, uxto.PkHex)
		}

		outputs = append(outputs, Amount{
			TargetLockTime: 0,
			Value:          int64(amount),
			Id:             coinId,
			Address:        to,
		})

		fess := uint64(len(inputs)*1500+(len(outputs)+1)*450+2000) * 2
		if total < amount+fess {
			err := fmt.Errorf("%s total %d no enough, need %d", from, total, amount+fess)
			return &Result{
				Success: 0,
				Error:   err.Error(),
				Data:    nil,
			}
		}
		change := total - (sum + fess)
		outputs = append(outputs, Amount{
			TargetLockTime: 0,
			Value:          int64(change),
			Id:             coinId,
			Address:        from,
		})

		now := time.Now()
		tx, err := TxEncode(1, uint32(90818), &now, inputs, outputs)
		if err != nil {
			log.Warnf("TxEncode %s", err.Error())

		}
		signed, err := qx.TxSign([]string{key}, tx, e.cfg.Rpc[0].Net, pks)
		if err != nil {
			return &Result{
				Success: 0,
				Error:   err.Error(),
				Data:    nil,
			}
		}
		txid, err := e.rpcs[0].SendTransaction(signed)
		if err != nil {
			log.Warnf("send %s", err.Error())
			return &Result{
				Success: 0,
				Error:   err.Error(),
				Data:    nil,
			}
		}
		return &Result{
			Success: 0,
			Error:   "",
			Data:    txid,
		}
	}
	return &Result{
		Success: 0,
		Error:   "no enough",
		Data:    nil,
	}
}

type Out struct {
	Txid       string  `json:"txid"`
	Number     int     `json:"number"`
	Address    string  `json:"address"`
	Amount     float64 `json:"amount"`
	LockHeight uint64  `json:"lockheight"`
	PkHex      string  `json:"pkhex"`
}

func (e *Engine) GetAddressUTXO(address string, coinId string, amount uint64, fees uint64) ([]Out, error) {
	var sum uint64
	lastHeight, err := e.db.GetLastHeight()
	if err != nil {
		return nil, err
	}
	outs, err := e.db.QueryAddressUsableUTXO(address, coinId, lastHeight)
	if err != nil {
		return nil, err
	}
	maxAmount := amount + fees
	for i, out := range outs {
		sum += out.Amount
		if sum >= maxAmount {
			outs = outs[0 : i+1]
			break
		}
	}
	outputs := []Out{}
	for _, out := range outs {
		outputs = append(outputs, Out{
			Txid:       out.TxId,
			Number:     out.Number,
			Address:    out.Address,
			Amount:     float64(out.Amount) / 1e8,
			LockHeight: out.Lock,
			PkHex:      out.ScriptPubKey.Hex,
		})
	}

	return outputs, nil
}

func (e *Engine) GetAddressBigUTXO(address string, coinId string, amount uint64, fees uint64) ([]Out, error) {
	var sum uint64
	lastHeight, err := e.db.GetLastHeight()
	if err != nil {
		return nil, err
	}
	outs, err := e.db.QueryAddressUsableBigUTXO(address, coinId, lastHeight)
	if err != nil {
		return nil, err
	}
	maxAmount := amount + fees
	for i, out := range outs {
		sum += out.Amount
		if sum >= maxAmount {
			outs = outs[0 : i+1]
			break
		}
	}
	outputs := []Out{}
	for _, out := range outs {
		outputs = append(outputs, Out{
			Txid:       out.TxId,
			Number:     out.Number,
			Address:    out.Address,
			Amount:     float64(out.Amount) / 1e8,
			LockHeight: out.Lock,
			PkHex:      out.ScriptPubKey.Hex,
		})
	}

	return outputs, nil
}

type Balance struct {
	Address     string  `json:"address"`
	Balance     float64 `json:"balance"`
	Usable      float64 `json:"usable"`
	Locked      float64 `json:"locaked"`
	Unconfirmed float64 `json:"uncofirmed"`
}

func (e *Engine) AddrBalance(address string, coinId string) uint64 {
	height, err := e.db.GetLastHeight()
	if err != nil {
		return 0
	}
	usable, err := e.db.GetUsableAmount(address, coinId, height)
	if err != nil {
		return 0
	}
	return uint64(usable)
}
