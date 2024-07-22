package repository

import (
	"os"

	"github.com/linxGnu/grocksdb"
)

var RocksDB *grocksdb.DB

func InitRocksDb() {
	const path = "./data/rocksdb"
	os.MkdirAll(path, os.ModePerm)

	bbto := grocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetBlockCache(grocksdb.NewLRUCache(3 << 30))

	opts := grocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCreateIfMissing(true)

	db, err := grocksdb.OpenDb(opts, path)
	if err != nil {
		panic(err)
	}
	RocksDB = db
}
