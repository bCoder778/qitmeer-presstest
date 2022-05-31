package sqldb

import (
	"fmt"
	"git.diabin.com/BlockChain/qitmeer-presstest/config"
	"git.diabin.com/BlockChain/qitmeer-presstest/db/dbtypes"
	"github.com/bCoder778/qitmeer-sync/storage/types"
	"github.com/bCoder778/qitmeer-sync/verify/stat"
	"time"
	"xorm.io/core"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	//"github.com/xormplus/xorm"
	"strings"

	//_ "github.com/lunny/godbc"
	_ "github.com/denisenkom/go-mssqldb"
)

type DB struct {
	engine *xorm.Engine
}

func ConnectMysql(conf *config.DB) (*DB, error) {
	path := strings.Join([]string{conf.User, ":", conf.Password, "@tcp(", conf.Address, ")/", conf.DBName}, "")
	engine, err := xorm.NewEngine("mysql", path)
	if err != nil {
		return nil, err
	}
	//engine.ShowSQL(true)

	//engine.SetMapper(core.SameMapper{})
	if err = engine.Sync2(
		dbtypes.Record{},
	); err != nil {
		return nil, err
	}

	return &DB{engine: engine}, nil
}

func ConnectSqlServer(conf *config.DB) (*DB, error) {
	path := fmt.Sprintf("server=%s;user id=%s;password=%s;database=%s", conf.Address, conf.User, conf.Password, conf.DBName)
	engine, err := xorm.NewEngine("mssql", path)

	if err != nil {
		return nil, err
	}
	engine.SetMapper(core.SameMapper{})
	engine.ShowSQL(false)
	if err = engine.Sync2(
		dbtypes.Record{},
	); err != nil {
		return nil, err
	}
	return &DB{engine}, nil
}

func (d *DB) Close() error {
	return d.engine.Close()
}

func (d *DB) Clear() error {
	return nil
}

type CustomMapper struct {
}

func (d *DB) GetSprinked(ip, address, coin string) (bool, error) {
	t := time.Now().Unix() - 60*60*24*1
	count, err := d.engine.Table(new(dbtypes.Record)).Where("address = ? and inTime > ? and coin = ? and stat != ?", address, t, coin, dbtypes.Failed).Count()
	if count > 0 {
		return true, fmt.Errorf("you can only apply for one address in one day")
	}
	count, err = d.engine.Table(new(dbtypes.Record)).Where("ip = ? and inTime > ? and coin = ? and stat != ?", ip, t, coin, dbtypes.Failed).Count()
	if count >= 10 {
		return true, fmt.Errorf("you can only apply for an IP address 10 times in one day")
	}
	fmt.Println(err)
	return false, nil
}

func (d *DB) Insert(ip string, address string, coin string, amount uint64) {
	_, err := d.engine.Table(new(dbtypes.Record)).Insert(&dbtypes.Record{
		Ip:      ip,
		Address: address,
		Coin:    coin,
		Amount:  amount,
		InTime:  time.Now().Unix(),
		TxId:    "",
		Stat:    0,
	})
	fmt.Println(err)

}
func (d *DB) GetUnSendRecord() []*dbtypes.Record {
	records := []*dbtypes.Record{}
	d.engine.Table(new(dbtypes.Record)).Where("stat = ?", dbtypes.UnSend).Find(&records)
	return records
}

func (d *DB) UpdateRecord(list []*dbtypes.Record, txid string, stat dbtypes.Stat) {
	sess := d.engine.NewSession()
	defer sess.Close()

	ids := ""
	for i, record := range list {
		if i == len(list)-1 {
			ids += fmt.Sprintf("%d", record.Id)
		} else {
			ids += fmt.Sprintf("%d,", record.Id)
		}
	}
	sess.Table(new(dbtypes.Record)).Where("find_in_set(id, ?)", ids).Update(
		map[string]interface{}{
			"txId": txid,
			"stat": stat,
		})
}

func (d *DB) GetUnconfirmed() []*dbtypes.Record {
	records := []*dbtypes.Record{}
	d.engine.Table(new(dbtypes.Record)).Where("stat = ?", dbtypes.Send).Find(&records)
	return records
}

func (d *DB) UpdateRecordByTxId(txid string, stat dbtypes.Stat) {
	d.engine.Table(new(dbtypes.Record)).Where("txId = ?", txid).Update(
		map[string]interface{}{
			"stat": stat,
		})
}

func (d *DB) QueryAmountUTXO(address string, coin string, amount uint64) []types.Vout {
	var vouts []types.Vout
	d.engine.Table(new(types.Vout)).Where("address = ? and coin_id = ? and amount = ? and spent_tx = '' and stat = 0", address, coin, amount).Find(&vouts)
	return vouts
}

func (d *DB) QueryAmountUTXOCount(address string, coin string, amount uint64) int64 {
	count, _ := d.engine.Table(new(types.Vout)).Where("address = ? and coin_id = ? and amount = ? and spent_tx = '' and stat = 0", address, coin, amount).Count()
	return count
}

func (d *DB) QueryAmountAllUTXO(address string, coin string, amount uint64) []types.Vout {
	var vouts []types.Vout
	d.engine.Table(new(types.Vout)).Where("address = ? and coin_id = ? and amount = ? and spent_tx = '' and stat < 3", address, coin, amount).Find(&vouts)
	return vouts
}

func (d *DB) QueryAmountAllUTXOCount(address string, coin string, amount uint64) int64 {
	count, _ := d.engine.Table(new(types.Vout)).Where("address = ? and coin_id = ? and amount = ? and spent_tx = '' and stat < 3", address, coin, amount).Count()
	return count
}

func (d *DB) QueryAllUTXO(address string, coin string) []types.Vout {
	var vouts []types.Vout
	d.engine.Table(new(types.Vout)).Where("address = ? and coin_id = ? and spent_tx = '' and stat = 0", address, coin).Find(&vouts)
	return vouts
}

func (d *DB) QuerySplitUTXO(address string, coin string) []types.Vout {
	var vouts []types.Vout
	d.engine.Table(new(types.Vout)).Where("address = ? and coin_id =? and amount > 10000 and spent_tx = '' and stat = 0", address, coin).Find(&vouts)
	return vouts
}

func (d *DB) UpdateSpent(vouts []types.Vout, spent string) {
	sess := d.engine.NewSession()
	defer sess.Close()
	for _, vout := range vouts {
		sess.Table(new(types.Vout)).Where("tx_id = ? and number = ?", vout.TxId, vout.Number).
			Update(map[string]interface{}{
				"spent_tx": spent,
			})

	}
	sess.Commit()
}

func (d *DB) GetLastHeight() (uint64, error) {
	var block = &types.Block{}
	_, err := d.engine.Table(new(types.Block)).Desc("height").Get(block)
	return block.Height, err
}

func (d *DB) QueryAddressUsableUTXO(address string, coinId string, lastHeight uint64) ([]*types.Vout, error) {
	vouts := []*types.Vout{}
	err := d.engine.Where("amount != 0 and vout.lock <= ? and address = ? and coin_id = ? and spent_tx = ? and  stat = ?", lastHeight, address, coinId, "", stat.TX_Confirmed).Find(&vouts)
	return vouts, err
}

func (d *DB) QueryAddressUsableBigUTXO(address string, coinId string, lastHeight uint64) ([]*types.Vout, error) {
	vouts := []*types.Vout{}
	err := d.engine.Where("amount > 10000 and vout.lock <= ? and address = ? and coin_id = ? and spent_tx = ? and  stat = ?", lastHeight, address, coinId, "", stat.TX_Confirmed).Find(&vouts)
	return vouts, err
}

func (d *DB) GetUsableAmount(address string, coinId string, height uint64) (float64, error) {
	return d.engine.Table(new(types.Vout)).Where("amount != 0 and vout.lock <= ?  and address = ? and coin_id = ? and spent_tx = ? and stat = ?",
		height, address, coinId, "", stat.TX_Confirmed).
		Sum(new(types.Vout), "amount")
}

func (d *DB) GetUnconfirmedAmount(address string, coinId string) (float64, error) {
	return d.engine.Table(new(types.Vout)).Where("amount != 0 and address = ? and coin_id = ? and spent_tx = ? and (is_coinbase = ? or (is_coinbase = ? and is_blue = ?)) and stat in (?, ?)",
		address, coinId, "", 0, 1, 1, stat.TX_Unconfirmed, stat.TX_Memry).
		Sum(new(types.Vout), "amount")
}

func (d *DB) GetLockedAmount(address string, coinId string, height uint64) (float64, error) {
	return d.engine.Table(new(types.Vout)).Where("amount != 0 and vout.lock > ? and address = ? and coin_id = ? and spent_tx = ? and stat = ?",
		height, address, coinId, "", stat.TX_Confirmed).
		Sum(new(types.Vout), "amount")
}
