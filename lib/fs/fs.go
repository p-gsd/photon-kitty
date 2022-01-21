//Package fs implements various fs functions for lua
package fs

import (
	"os"

	lua "github.com/yuin/gopher-lua"
)

func Preload(L *lua.LState) {
	L.PreloadModule("fs", Loader)
}

// Loader is the module loader function.
func Loader(L *lua.LState) int {
	t := L.NewTable()
	L.SetFuncs(t, api)
	L.Push(t)
	return 1
}

var api = map[string]lua.LGFunction{
	"stat":     Stat,
	"mkdirAll": MkdirAll,
	"home":     Home,
}

// Stat lua goos.stat(filename) returns (table, err)
func Stat(L *lua.LState) int {
	filename := L.CheckString(1)
	stat, err := os.Stat(filename)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	result := L.NewTable()
	result.RawSetString(`is_dir`, lua.LBool(stat.IsDir()))
	result.RawSetString(`size`, lua.LNumber(stat.Size()))
	result.RawSetString(`mod_time`, lua.LNumber(stat.ModTime().Unix()))
	result.RawSetString(`mode`, lua.LString(stat.Mode().String()))
	L.Push(result)
	return 1
}

func MkdirAll(L *lua.LState) int {
	path := L.CheckString(1)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	return 0
}

func Home(L *lua.LState) int {
	path, err := os.UserHomeDir()
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	L.Push(lua.LString(path))
	return 1
}
