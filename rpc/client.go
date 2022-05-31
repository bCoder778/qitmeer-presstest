package rpc

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"git.diabin.com/BlockChain/qitmeer-presstest/config"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	rpcAuth *config.Rpc
}

func NewClient(auth *config.Rpc) *Client {

	return &Client{rpcAuth: auth}
}

func (c *Client) GetBlock(h uint64) (*Block, error) {
	params := []interface{}{h, true}
	resp := NewReqeust(params).SetMethod("getBlockByOrder").call(c.rpcAuth)
	blk := new(Block)
	if resp.Error != nil {
		return blk, errors.New(resp.Error.Message)
	}
	if err := json.Unmarshal(resp.Result, blk); err != nil {
		return blk, err
	}
	return blk, nil
}

func (c *Client) GetBlockByHash(hash string) (*Block, error) {
	params := []interface{}{hash, true}
	resp := NewReqeust(params).SetMethod("getBlock").call(c.rpcAuth)
	blk := new(Block)
	if resp.Error != nil {
		return blk, errors.New(resp.Error.Message)
	}
	if err := json.Unmarshal(resp.Result, blk); err != nil {
		return blk, err
	}
	return blk, nil
}

func (c *Client) GetBlockCount() (uint64, error) {
	var params []interface{}
	resp := NewReqeust(params).SetMethod("getBlockCount").call(c.rpcAuth)
	if resp.Error != nil {
		return 0, errors.New(resp.Error.Message)
	}
	count, err := strconv.ParseUint(string(resp.Result), 10, 64)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (c *Client) GetMainChainHeight() (uint64, error) {
	var params []interface{}
	resp := NewReqeust(params).SetMethod("getMainChainHeight").call(c.rpcAuth)
	if resp.Error != nil {
		return 0, errors.New(resp.Error.Message)
	}
	height, err := strconv.ParseUint(string(resp.Result), 10, 64)
	if err != nil {
		return 0, err
	}
	return height, nil
}

func (c *Client) SendTransaction(tx string) (string, error) {
	params := []interface{}{strings.Trim(tx, "\n"), true}
	resp := NewReqeust(params).SetMethod("sendRawTransaction").call(c.rpcAuth)
	if resp.Error != nil {
		return resp.Error.Message, errors.New(resp.Error.Message)
	}
	txid := ""
	json.Unmarshal(resp.Result, &txid)
	return txid, nil
}

func (c *Client) GetTransaction(txId string) (*Transaction, error) {
	params := []interface{}{txId, true}
	resp := NewReqeust(params).SetMethod("getRawTransaction").call(c.rpcAuth)
	if resp.Error != nil {
		return nil, errors.New(resp.Error.Message)
	}
	var rs *Transaction
	if err := json.Unmarshal(resp.Result, &rs); err != nil {
		return nil, err
	}
	return rs, nil
}

func (c *Client) CreateTransaction(inputs []TransactionInput, amounts Amounts) (string, error) {
	jsonInput, err := json.Marshal(inputs)
	if err != nil {
		return "", err
	}
	jsonAmount, err := json.Marshal(amounts)
	if err != nil {
		return "", err
	}
	params := []interface{}{json.RawMessage(jsonInput), json.RawMessage(jsonAmount)}
	resp := NewReqeust(params).SetMethod("createRawTransaction").call(c.rpcAuth)
	if resp.Error != nil {
		return "", errors.New(resp.Error.Message)
	}
	encode := string(resp.Result)
	return encode, nil
}

func (c *Client) GetMemoryPool() ([]string, error) {
	params := []interface{}{"", false}
	resp := NewReqeust(params).SetMethod("getMempool").call(c.rpcAuth)
	if resp.Error != nil {
		return nil, errors.New(resp.Error.Message)
	}
	var rs []string
	if err := json.Unmarshal(resp.Result, &rs); err != nil {
		return nil, err
	}
	return rs, nil
}

func (c *Client) GetBlockById(id uint64) (*Block, error) {
	params := []interface{}{id, true}
	resp := NewReqeust(params).SetMethod("getBlockByID").call(c.rpcAuth)
	blk := new(Block)
	if resp.Error != nil {
		return blk, errors.New(resp.Error.Message)
	}
	if err := json.Unmarshal(resp.Result, blk); err != nil {
		return blk, err
	}
	return blk, nil
}

func (c *Client) IsBlue(hash string) (int, error) {
	params := []interface{}{hash}
	resp := NewReqeust(params).SetMethod("isBlue").call(c.rpcAuth)
	if resp.Error != nil {
		return 0, errors.New(resp.Error.Message)
	}
	state, err := strconv.Atoi(string(resp.Result))
	if err != nil {
		return 0, err
	}
	return state, nil
}

func (c *Client) GetFees(hash string) (uint64, error) {
	params := []interface{}{hash}
	resp := NewReqeust(params).SetMethod("getFees").call(c.rpcAuth)
	if resp.Error != nil {
		return 0, errors.New(resp.Error.Message)
	}

	return strconv.ParseUint(string(resp.Result), 10, 64)
}

func (c *Client) GetPeerInfo() ([]PeerInfo, error) {
	var params []interface{}
	resp := NewReqeust(params).SetMethod("getPeerInfo").call(c.rpcAuth)
	if resp.Error != nil {
		return nil, errors.New(resp.Error.Message)
	}
	var rs []PeerInfo
	if err := json.Unmarshal(resp.Result, &rs); err != nil {
		return nil, err
	}
	return rs, nil
}

func (c *Client) GetCoins() ([]Coin, error) {
	resp := NewReqeust(nil).SetMethod("getTokenInfo").call(c.rpcAuth)
	coins := []Coin{}
	if resp.Error != nil {
		return coins, errors.New(resp.Error.Message)
	}
	if err := json.Unmarshal(resp.Result, &coins); err != nil {
		return coins, err
	}
	return coins, nil
}

func (req *ClientRequest) call(auth *config.Rpc) *ClientResponse {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr, Timeout: time.Second * 5}
	defer client.CloseIdleConnections()
	//convert struct to []byte
	marshaledData, err := json.Marshal(req)
	if err != nil {
		return &ClientResponse{Error: &Error{Message: err.Error()}}
	}

	httpRequest, err :=
		http.NewRequest(http.MethodPost, auth.Host, bytes.NewReader(marshaledData))
	if err != nil {
		return &ClientResponse{Error: &Error{Message: err.Error()}}
	}
	if httpRequest == nil {
		return &ClientResponse{Error: &Error{Message: "the httpRequest is nil"}}
	}
	httpRequest.Close = true
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.SetBasicAuth(auth.Admin, auth.Password)
	//log.Debugf("u:%s;p:%s", cfg.User, cfg.Pwd)

	response, err := client.Do(httpRequest)
	if err != nil {
		return &ClientResponse{Error: &Error{Message: err.Error()}}
	}

	body := response.Body

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return &ClientResponse{Error: &Error{Message: err.Error()}}
	}

	resp := &ClientResponse{}
	if err := json.Unmarshal(bodyBytes, resp); err != nil {
		return &ClientResponse{Error: &Error{Message: err.Error()}}
	}

	err = response.Body.Close()
	if err != nil {
		return &ClientResponse{Error: &Error{Message: err.Error()}}
	}
	return resp
}
