package rocksdb

import (
	"fmt"
	"os"
	"time"

	"github.com/linxGnu/grocksdb"
)

var BuyOrder *grocksdb.DB
var SellOrder *grocksdb.DB

func Init() {
	cwd, _ := os.Getwd()

	bookName := ""
	if os.Getenv("ENV") == "test" {
		bookName = fmt.Sprintf("test_%d_", time.Now().UnixMilli())
	}

	buyOrderPath := fmt.Sprintf("%s/rocksdb_data/%sbuy_order", cwd, bookName)
	sellOrderPath := fmt.Sprintf("%s/rocksdb_data/%ssell_order", cwd, bookName)
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
