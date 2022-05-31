package rpc

import (
	"time"
)

type Block struct {
	Id            uint64        `json:"-"`
	Hash          string        `json:"hash"`
	Weight        uint64        `json:"weight"`
	Height        uint64        `json:"height"`
	Txsvalid      bool          `json:"txsvalid"`
	Confirmations uint64        `json:"confirmations"`
	Version       uint32        `json:"version"`
	Order         uint64        `json:"order"`
	TxRoot        string        `json:"txRoot"`
	Transactions  []Transaction `json:"transactions"`
	StateRoot     string        `json:"stateroot"`
	Bits          string        `json:"bits"`
	Difficulty    uint64        `json:"difficulty"`
	Timestamp     time.Time     `json:"timestamp"`
	ParentRoot    string        `json:"parentroot"`
	Parents       []string      `json:"parents"`
	Children      []string      `json:"children"`
	Pow           *Pow          `json:"pow"`
	IsBlue        int           `json:"isblue"`
}

type Pow struct {
	PowName   string     `json:"pow_name"`
	PowType   int        `json:"pow_type"`
	Nonce     uint64     `json:"nonce"`
	ProofData *ProofData `json:"proof_data"`
}

type ProofData struct {
	EdgeBits     int    `json:"edge_bits"`
	CircleNonces string `json:"circle_nonces"`
}

//区块中对应的每一个transaction
type Transaction struct {
	Hex           string    `json:"hex"`
	Hexwit        string    `json:"hexwit"`
	Hexnowit      string    `json:"hexnowit"`
	Txid          string    `json:"txid"`
	Txhash        string    `json:"txhash"`
	Version       uint32    `json:"version"`
	Locktime      uint64    `json:"locktime"`
	Timestamp     time.Time `json:"timestamp"`
	Expire        uint64    `json:"expire"`
	Vin           []Vin     `json:"vin"`
	Vout          []Vout    `json:"vout"`
	Confirmations uint64    `json:"confirmations"`
	Txsvalid      bool      `json:"txsvalid"`
	Duplicate     bool      `json:"duplicate"`
	BlockHash     string    `json:"blockhash"`
	BlockOrder    uint64    `json:"blockorder"`
	Size          int       `json:"size"`
}

type Vin struct {
	//在有交易打包后才有此字段，也就是说接收过他人转账，并且有可用的utxo都块才会在vin中包含这个字段
	Txid string `json:"txid"`
	Type string `json:"type"`
	//在有交易打包后才有此字段
	Vout        int    `json:"vout"`
	Amountin    uint64 `json:"amountin"`
	Blockheight uint64 `json:"blockheight"`
	Blockindex  uint64 `json:"blockindex"`
	Coinbase    string `json:"coinbase"`
	Txindex     uint64 `json:"txindex"`
	//在无交易打包时才有此字段
	Sequence uint64 `json:"sequence"`
	//在有交易打包后才有此字段(代表私钥加签)
	ScriptSig ScriptSig `json:"scriptSig"`
}

type ScriptSig struct {
	Asm string `json:"asm"`
	Hex string `json:"hex"`
}

type Vout struct {
	CoinID       string       `json:"coin"`
	Amount       uint64       `json:"amount"`
	ScriptPubKey ScriptPubKey `json:"scriptpubkey"`
}

type ScriptPubKey struct {
	Asm       string   `json:"asm"`
	Hex       string   `json:"hex"`
	ReqSigs   int      `json:"reqSigs"`
	Type      string   `json:"type"`
	Addresses []string `json:"addresses"`
}

type GraphState struct {
	Tips       []string `json:"tips"`
	MainOrder  uint64   `json:"mainorder"`
	MainHeight uint64   `json:"mainheight"`
	Layer      uint64   `json:"layer"`
}

type PeerInfo struct {
	Id         uint64     `json:"id"`
	Addr       string     `json:"addr"`
	AddrLocal  string     `json:"addrlocal"`
	Services   string     `json:"services"`
	Relaytxes  bool       `json:"relaytxes"`
	LastSend   uint64     `json:"lastsend"`
	LastRecv   uint64     `json:"lastrecv"`
	BytesSent  uint64     `json:"bytessent"`
	BytesRecv  uint64     `json:"bytesrecv"`
	Conntime   uint64     `json:"conntime"`
	TimeOffSet int64      `json:"timeoffset"`
	PingTime   uint64     `json:"pingtime"`
	Version    uint64     `json:"version"`
	Subver     string     `json:"subver"`
	Inbound    bool       `json:"inbound"`
	BansCore   uint64     `json:"banscore"`
	SyncNode   bool       `json:"syncnode"`
	GraphState GraphState `json:"graphstate"`
}

type Coin struct {
	CoinId   uint16 `xorm:"bigint pk id" json:"coinId"`
	CoinName string `xorm:"varchar(35)" json:"coinName"`
}
