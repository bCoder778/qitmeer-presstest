package db

import (
	"fmt"
	"git.diabin.com/BlockChain/qitmeer-presstest/config"
	"git.diabin.com/BlockChain/qitmeer-presstest/db/dbtypes"
	"git.diabin.com/BlockChain/qitmeer-presstest/db/sqldb"
	"github.com/bCoder778/qitmeer-sync/storage/types"
)

type IDB interface {
	GetSprinked(ip string, address string, coin string) (bool, error)
	GetUnSendRecord() []*dbtypes.Record
	GetUnconfirmed() []*dbtypes.Record
	Insert(ip string, address string, coin string, amount uint64)
	UpdateRecord(list []*dbtypes.Record, txid string, stat dbtypes.Stat)
	UpdateRecordByTxId(txid string, stat dbtypes.Stat)
	QueryAmountUTXO(address string, coin string, amount uint64) []types.Vout
	QueryAmountUTXOCount(address string, coin string, amount uint64) int64
	QueryAmountAllUTXO(address string, coin string, amount uint64) []types.Vout
	QueryAmountAllUTXOCount(address string, coin string, amount uint64) int64
	QuerySplitUTXO(address string, coin string) []types.Vout
	UpdateSpent(vouts []types.Vout, spent string)
	GetLastHeight() (uint64, error)
	QueryAddressUsableUTXO(address string, coinId string, lastHeight uint64) ([]*types.Vout, error)
	QueryAddressUsableBigUTXO(address string, coinId string, lastHeight uint64) ([]*types.Vout, error)
	GetUsableAmount(address string, coinId string, height uint64) (float64, error)
	GetUnconfirmedAmount(address string, coinId string) (float64, error)
	GetLockedAmount(address string, coinId string, height uint64) (float64, error)
	QueryAllUTXO(address string, coin string) []types.Vout
	Close() error
	Clear() error
}

type IList interface {
}

func ConnectDB(setting *config.Config) (IDB, error) {
	var (
		db  IDB
		err error
	)
	switch setting.DB.DBType {
	case "mysql":
		if db, err = sqldb.ConnectMysql(setting.DB); err != nil {
			return nil, fmt.Errorf("failed to connect mysql, error:%v", err)
		}
	case "sqlserver":
		if db, err = sqldb.ConnectSqlServer(setting.DB); err != nil {
			return nil, fmt.Errorf("failed to connect mysql, error:%v", err)
		}
	default:
		return nil, fmt.Errorf("unsupported database %s", setting.DB.DBType)
	}
	return db, nil
}
