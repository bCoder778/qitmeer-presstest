package external

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	getIpInfo = "http://ip-api.com/json/"
	//getPrices = "https://api.hotbit.io/api/v1/market.status?market=PMEER/USDT&period=1800"
	getPrices  = "https://apiv4.upex.io/exchange-open-api/open/api/get_ticker?symbol=pmeerusdt"
	getBalance = "https://testnet.meerscan.io/kahf/api/v1/status/address"
	getUTXO    = "https://testnet.meerscan.io/kahf/api/v1/utxo"
)

type Balance struct {
	Code   int     `json:"code"`
	Msg    string  `json:"msg"`
	Result *Result `json:"rs"`
}

type Result struct {
	Address    string  `json:"address"`
	Balance    float64 `json:"balance"`
	Usable     float64 `json:"usable"`
	Locaked    float64 `json:"locaked"`
	Uncofirmed float64 `json:"uncofirmed"`
}

type UXTOS struct {
	Code   int         `json:"code"`
	Msg    string      `json:"msg"`
	Result *UTXOResult `json:"rs"`
}

type UTXOResult struct {
	Outs   []Out  `json:"outs"`
	Tomany bool   `json:"tomany"`
	Enough bool   `json:"enough"`
	Height uint64 `json:"height"`
}

type Out struct {
	Txid       string  `json:"txid"`
	Number     int     `json:"number"`
	Address    string  `json:"address"`
	Amount     float64 `json:"amount"`
	LockHeight uint64  `json:"lockheight"`
	PkHex      string  `json:"pkhex"`
}

/*
func GetBalance(address string, coin string) uint64 {
	param := map[string]interface{}{
		"address": address,
		"coin":    coin,
	}
	url := fmt.Sprintf("%s?address=%s&coin=%s", getBalance, address, coin)
	bytes, err := getBody("GET", url, param)
	if err != nil {
		return 0
	}
	balance := &Balance{Result: &Result{}}
	err = json.Unmarshal(bytes, balance)
	if err != nil {
		return 0
	}
	return uint64(balance.Result.Usable * 1e8)
}

func GetUTXO(address string, coin string, amount, fees float64) ([]Out, uint64) {
	param := map[string]interface{}{
		"address": address,
		"coin":    coin,
		"amount":  amount,
		"fees":    fees,
	}
	url := fmt.Sprintf("%s?address=%s&coin=%s&amount=%.8f&fees=%.8f", getUTXO, address, coin, amount, fees)
	bytes, err := getBody("GET", url, param)
	if err != nil {
		return []Out{}, 0
	}
	outs := &UXTOS{}
	err = json.Unmarshal(bytes, outs)
	if err != nil {
		return []Out{}, 0
	}
	if outs.Result.Enough {
		return outs.Result.Outs, outs.Result.Height
	}
	return []Out{}, 0
}*/

func getBody(method string, url string, param map[string]interface{}) ([]byte, error) {
	paramBytes, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}
	buffer := bytes.NewBuffer(paramBytes)
	req, err := http.NewRequest(method, url, buffer)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Second * 20}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}
