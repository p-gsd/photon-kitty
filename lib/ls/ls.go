//Package ls implements a key-value store for lua (like localStorage in js)
package ls

import (
	"log"

	"github.com/syndtr/goleveldb/leveldb"
	lua "github.com/yuin/gopher-lua"
)

type LocalStorage struct {
	path string
	db   *leveldb.DB
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

func (ls *LocalStorage) open() *leveldb.DB {
	if ls.db != nil {
		return ls.db
	}
	db, err := leveldb.OpenFile(ls.path, nil)
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
	err := db.Put([]byte(key), []byte(value), nil)
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
	v, err := db.Get([]byte(key), nil)
	if err == leveldb.ErrNotFound {
		return 0
	}
	if err != nil {
		return 0
	}
	val = lua.LString(v)
	L.Push(val)
	return 1
}

//returns the number of stored items(data)
func (ls *LocalStorage) length(L *lua.LState) int {
	db := ls.open()
	var length int
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		length++
	}
	iter.Release()
	err := iter.Error()
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
	err := db.Delete([]byte(key), nil)
	if err != nil {
		log.Fatal(err)
	}
	return 0
}

//deletes everything from the storage
func (ls *LocalStorage) clear(L *lua.LState) int {
	db := ls.open()
	var length int
	iter := db.NewIterator(nil, nil)
	batch := &leveldb.Batch{}
	for iter.Next() {
		length++
		batch.Delete(iter.Key())
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Write(batch, nil); err != nil {
		log.Fatal(err)
	}
	L.Push(lua.LNumber(length))
	return 1
}
