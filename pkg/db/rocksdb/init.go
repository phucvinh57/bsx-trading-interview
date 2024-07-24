package rocksdb

import (
	"os"

	"github.com/linxGnu/grocksdb"
)

var BuyOrder *grocksdb.DB
var SellOrder *grocksdb.DB

func Init() {
	const buyOrderPath = "./rocksdb_data/buy_order"
	const sellOrderPath = "./rocksdb_data/sell_order"
	os.MkdirAll(buyOrderPath, os.ModePerm)
	os.MkdirAll(sellOrderPath, os.ModePerm)

	bbto := grocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetBlockCache(grocksdb.NewLRUCache(3 << 30))

	opts := grocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCreateIfMissing(true)

	var err error
	BuyOrder, err = grocksdb.OpenDb(opts, buyOrderPath)
	if err != nil {
		panic(err)
	}
	SellOrder, err = grocksdb.OpenDb(opts, sellOrderPath)
	if err != nil {
		panic(err)
	}
}
