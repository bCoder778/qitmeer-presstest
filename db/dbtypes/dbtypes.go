package dbtypes

type Stat int8

const (
	UnSend    = 0
	Send      = 1
	Confirmed = 2
	Failed    = 3
)

type Record struct {
	Id      int64  `xorm:"bigint autoincr pk" `
	Ip      string `xorm:"varchar(20) ip" `
	Address string `xorm:"varchar(40) address"`
	Coin    string `xorm:"varchar(20) coin"`
	Amount  uint64 `xorm:"bigint amount"`
	InTime  int64  `xorm:"bigint inTime"`
	TxId    string `xorm:"varchar(70) txId"`
	Stat    int8   `xorm:"int stat"`
}
