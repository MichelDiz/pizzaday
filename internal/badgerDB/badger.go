package badgerDB

import (
	"log"
	"sync"

	badger "github.com/dgraph-io/badger/v4"
)

var (
	instance *badger.DB
	once     sync.Once
)

func InitDB(path string) *badger.DB {
	once.Do(func() {
		opts := badger.DefaultOptions(path)
		opts.Logger = nil
		var err error
		instance, err = badger.Open(opts)
		if err != nil {
			log.Fatal(err)
		}
	})
	return instance
}

func GetDB() *badger.DB {
	return instance
}
