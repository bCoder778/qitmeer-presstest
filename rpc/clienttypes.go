package rpc

import "encoding/json"

//Params类型不确定，可以是数组，也可以是map参数
type ClientRequest struct {
	JsonRpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      interface{}   `json:"id"`
	Host    string
	User    string
	Pwd     string
}

//Result类型不确定，可以是string，也可以是int等类型
type ClientResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *Error          `json:"error"`
	ID     interface{}     `json:"id"`
}

type Error struct {
	Code    int         `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func NewReqeust(params []interface{}) *ClientRequest {
	return &ClientRequest{JsonRpc: "2.0", Id: 1, Params: params}
}

func (req *ClientRequest) SetMethod(method string) *ClientRequest {
	req.Method = method
	return req
}

type TransactionInput struct {
	Txid string `json:"txid"`
	Vout int    `json:"vout"`
}

type Amounts map[string]uint64 //{\"address\":amount,...}
