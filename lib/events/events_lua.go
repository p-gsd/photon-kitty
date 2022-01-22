package events

import (
	lua "github.com/yuin/gopher-lua"
)

func New(L *lua.LState) lua.LValue {
	var exports = map[string]lua.LGFunction{
		"subscribe": eventsSubscribe,
	}
	mod := L.SetFuncs(L.NewTable(), exports)
	registerEvents(L)
	L.SetField(mod, "Init", lua.LString(EventTypeInit))
	L.SetField(mod, "RunMediaStart", lua.LString(EventTypeRunMediaStart))
	L.SetField(mod, "RunMediaEnd", lua.LString(EventTypeRunMediaEnd))
	L.SetField(mod, "FeedsDownloaded", lua.LString(EventTypeFeedsDownloaded))
	L.SetField(mod, "ArticleOpened", lua.LString(EventTypeArticleOpened))
	L.SetField(mod, "LinkOpened", lua.LString(EventTypeLinkOpened))
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

func registerEvents(L *lua.LState) {
	//Init
	mt := L.NewTypeMetatable(string(EventTypeInit))
	L.SetGlobal(string(EventTypeInit), mt)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), nil))

	//FeedsDownloaded
	mt = L.NewTypeMetatable(string(EventTypeFeedsDownloaded))
	L.SetGlobal(string(EventTypeInit), mt)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), nil))

	//RunMediaStart
	methods := map[string]lua.LGFunction{
		"link": eventLink,
		"card": eventCard,
	}
	mt = L.NewTypeMetatable(string(EventTypeRunMediaStart))
	L.SetGlobal(string(EventTypeRunMediaStart), mt)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), methods))

	//RunMediaEnd
	mt = L.NewTypeMetatable(string(EventTypeRunMediaEnd))
	L.SetGlobal(string(EventTypeRunMediaEnd), mt)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), methods))

	//ArticleOpened
	mt = L.NewTypeMetatable(string(EventTypeArticleOpened))
	L.SetGlobal(string(EventTypeArticleOpened), mt)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), methods))

	//LinkOpened
	mt = L.NewTypeMetatable(string(EventTypeLinkOpened))
	L.SetGlobal(string(EventTypeLinkOpened), mt)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), methods))
}

func eventLink(L *lua.LState) int {
	ud := L.CheckUserData(1)
	switch e := ud.Value.(type) {
	case *Init:
		L.ArgError(1, "init event doesn't have link method")
		return 0
	case *RunMediaStart:
		L.Push(lua.LString(e.Link))
	case *RunMediaEnd:
		L.Push(lua.LString(e.Link))
	case *ArticleOpened:
		L.Push(lua.LString(e.Link))
	case *LinkOpened:
		L.Push(lua.LString(e.Link))
	case *FeedsDownloaded:
		L.ArgError(1, "feedsDownloaded event doesn't have link method")
		return 0
	default:
		L.ArgError(1, "event expected")
		return 0
	}
	return 1
}

func eventCard(L *lua.LState) int {
	ud := L.CheckUserData(1)
	switch e := ud.Value.(type) {
	case *Init:
		L.ArgError(1, "init event doesn't have link method")
		return 0
	case *RunMediaStart:
		L.Push(e.Card(L))
	case *RunMediaEnd:
		L.Push(e.Card(L))
	case *ArticleOpened:
		L.Push(e.Card(L))
	case *LinkOpened:
		L.Push(e.Card(L))
	case *FeedsDownloaded:
		L.ArgError(1, "feedsDownloaded event doesn't have link method")
		return 0
	default:
		L.ArgError(1, "event expected")
		return 0
	}
	return 1
}

func eventToLuaValue(L *lua.LState, e Event) lua.LValue {
	ud := L.NewUserData()
	ud.Value = e
	L.SetMetatable(ud, L.GetTypeMetatable(string(e.Type())))
	return ud
}
