package events

import (
	lua "github.com/yuin/gopher-lua"
)

func New(L *lua.LState) lua.LValue {
	var exports = map[string]lua.LGFunction{
		"subscribe": eventsSubscribe,
	}
	mod := L.SetFuncs(L.NewTable(), exports)
	registerEventType(L)
	L.SetField(mod, "Init", lua.LString(EventTypeInit))
	L.SetField(mod, "RunMediaStart", lua.LString(EventTypeRunMediaStart))
	L.SetField(mod, "RunMediaEnd", lua.LString(EventTypeRunMediaEnd))
	L.SetField(mod, "FeedsDownloaded", lua.LString(EventTypeFeedsDownloaded))
	return mod
}

func eventsSubscribe(L *lua.LState) int {
	t := L.ToString(1)
	f := L.ToFunction(2)
	et := EventType(t)
	fn := func(e Event) error {
		L.Push(f)
		L.Push(eventToLuaValue(L, e))
		if err := L.PCall(1, lua.MultRet, nil); err != nil {
			L.Error(lua.LString(err.Error()), 0)
			return err
		}
		return nil
	}
	Subscribe(et, fn)
	return 0
}

const luaEventTypeName = "photon.event"

func eventToLuaValue(L *lua.LState, e Event) lua.LValue {
	ud := L.NewUserData()
	ud.Value = e
	L.SetMetatable(ud, L.GetTypeMetatable(luaEventTypeName))
	return ud
}

func registerEventType(L *lua.LState) {
	var eventMethods = map[string]lua.LGFunction{
		"type": eventType,
	}
	mt := L.NewTypeMetatable(luaEventTypeName)
	L.SetGlobal(luaEventTypeName, mt)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), eventMethods))
}

func checkEvent(L *lua.LState) Event {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(Event); ok {
		return v
	}
	L.ArgError(1, "event expected")
	return nil
}

func eventType(L *lua.LState) int {
	e := checkEvent(L)
	L.Push(lua.LString(e.Type()))
	return 1
}
