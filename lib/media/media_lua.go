package media

import lua "github.com/yuin/gopher-lua"

const luaMediaTypeName = "photon.media"

func Loader(L *lua.LState) {
	mt := L.NewTypeMetatable(luaMediaTypeName)
	L.SetField(mt, "__index", L.NewFunction(mediaIndex))
}

func checkMedia(L *lua.LState) *Media {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*Media); ok {
		return v
	}
	L.ArgError(1, luaMediaTypeName+" expected")
	return nil
}

func NewLuaMedia(media *Media, L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = media
	L.SetMetatable(ud, L.GetTypeMetatable(luaMediaTypeName))
	return ud
}

func mediaIndex(L *lua.LState) int {
	media := checkMedia(L)

	switch L.CheckString(2) {
	case "originalLink":
		return mediaOriginalLink(media, L)
	case "links":
		return mediaLinks(media, L)
	case "contentType":
		return mediaContentType(media, L)
	case "run":
		media.Run()
	}

	return 0
}

func mediaOriginalLink(media *Media, L *lua.LState) int {
	L.Push(lua.LString(media.OriginalLink))
	return 1
}

func mediaLinks(media *Media, L *lua.LState) int {
	links := L.NewTable()
	for i, link := range media.Links {
		links.RawSetInt(i+1, lua.LString(link))
	}
	L.Push(links)
	return 1
}

func mediaContentType(media *Media, L *lua.LState) int {
	L.Push(lua.LString(media.ContentType))
	return 1
}
