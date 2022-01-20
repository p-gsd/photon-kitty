package keybindings

import (
	"git.sr.ht/~ghost08/photon/lib/states"
	lua "github.com/yuin/gopher-lua"
)

func NewLValue(L *lua.LState, kb *Registry) lua.LValue {
	var exports = map[string]lua.LGFunction{
		"add": keybindingsAdd(kb),
	}
	return L.SetFuncs(L.NewTable(), exports)
}

func keybindingsAdd(kb *Registry) lua.LGFunction {
	return func(L *lua.LState) int {
		state := states.Enum(L.ToNumber(1))
		keyString := L.ToString(2)
		fn := L.ToFunction(3)
		kb.Add(state, keyString, func() error {
			L.Push(fn)
			return L.PCall(0, lua.MultRet, nil)
		})
		return 0
	}
}
