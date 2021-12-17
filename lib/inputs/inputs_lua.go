package inputs

import (
	lua "github.com/yuin/gopher-lua"
)

func Loader(in *Inputs) lua.LGFunction {
	return func(L *lua.LState) int {
		var exports = map[string]lua.LGFunction{
			"len": inputsLen(in),
			"get": inputsGet(in),
			"set": inputsSet(in),
		}
		mod := L.SetFuncs(L.NewTable(), exports)
		L.Push(mod)

		return 1
	}
}

func inputsLen(in *Inputs) lua.LGFunction {
	return func(L *lua.LState) int {
		L.Push(lua.LNumber(in.Len()))
		return 1
	}
}

func inputsGet(in *Inputs) lua.LGFunction {
	return func(L *lua.LState) int {
		i := L.ToInt(1)
		L.Push(lua.LString(in.Get(i - 1)))
		return 1
	}
}

func inputsSet(in *Inputs) lua.LGFunction {
	return func(L *lua.LState) int {
		i := L.ToInt(1)
		v := L.ToString(2)
		in.Set(i-1, v)
		return 0
	}
}
