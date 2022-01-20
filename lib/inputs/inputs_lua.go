package inputs

import (
	lua "github.com/yuin/gopher-lua"
)

func New(L *lua.LState, in *Inputs) lua.LValue {
	var exports = map[string]lua.LGFunction{
		"len":    inputsLen(in),
		"get":    inputsGet(in),
		"set":    inputsSet(in),
		"add":    inputsAdd(in),
		"append": inputsAppend(in),
	}
	return L.SetFuncs(L.NewTable(), exports)
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

func inputsAdd(in *Inputs) lua.LGFunction {
	return func(L *lua.LState) int {
		i := L.ToInt(1)
		v := L.ToString(2)
		in.Add(i-1, v)
		return 0
	}
}

func inputsAppend(in *Inputs) lua.LGFunction {
	return func(L *lua.LState) int {
		v := L.ToString(1)
		in.Append(v)
		return 0
	}
}
