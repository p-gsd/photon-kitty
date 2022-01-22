//Package ls implements a key-value store for lua (like localStorage in js)
package ls

import (
	"log"

	"github.com/dgraph-io/badger/v3"
	lua "github.com/yuin/gopher-lua"
)

type LocalStorage struct {
	path string
	db   *badger.DB
}

func New(path string) *LocalStorage {
	return &LocalStorage{path: path}
}

// Loader is the module loader function.
func (ls *LocalStorage) Loader(L *lua.LState) int {
	t := L.NewTable()
	L.SetFuncs(t, ls.makeAPI())
	L.Push(t)
	return 1
}

func (ls *LocalStorage) makeAPI() map[string]lua.LGFunction {
	api := map[string]lua.LGFunction{
		"setItem":    ls.setItem,
		"getItem":    ls.getItem,
		"length":     ls.length,
		"removeItem": ls.removeItem,
		"clear":      ls.clear,
	}
	return api
}

func (ls *LocalStorage) open() *badger.DB {
	if ls.db != nil {
		return ls.db
	}
	db, err := badger.Open(
		badger.DefaultOptions(ls.path).WithLogger(nil),
	)
	if err != nil {
		log.Fatal(err)
	}
	ls.db = db
	return db
}

func (ls *LocalStorage) Close() error {
	if ls.db == nil {
		return nil
	}
	return ls.db.Close()
}

//stores key/value pair
func (ls *LocalStorage) setItem(L *lua.LState) int {
	key := L.CheckString(1)
	value := L.CheckString(2)
	db := ls.open()
	err := db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), []byte(value))
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
	return 0
}

//returns the value in front of key
func (ls *LocalStorage) getItem(L *lua.LState) int {
	key := L.CheckString(1)
	db := ls.open()
	var val lua.LValue = lua.LNil
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		v, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		val = lua.LString(v)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	L.Push(val)
	return 1
}

//returns the number of stored items(data)
func (ls *LocalStorage) length(L *lua.LState) int {
	db := ls.open()
	var length int
	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			length++
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	L.Push(lua.LNumber(length))
	return 1
}

//removes given key with its value
func (ls *LocalStorage) removeItem(L *lua.LState) int {
	key := L.CheckString(1)
	db := ls.open()
	err := db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(key))
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
	return 0
}

//deletes everything from the storage
func (ls *LocalStorage) clear(L *lua.LState) int {
	db := ls.open()
	var length int
	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			if err := txn.Delete(it.Item().Key()); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	L.Push(lua.LNumber(length))
	return 1
}
