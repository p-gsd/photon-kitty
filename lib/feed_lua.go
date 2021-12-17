package lib

import (
	"github.com/mmcdole/gofeed"
	lua "github.com/yuin/gopher-lua"
)

const (
	luaFeedTypeName   = "photon.feed"
	luaPersonTypeName = "photon.person"
)

func feedLoader(L *lua.LState) int {
	var feedMethods = map[string]lua.LGFunction{
		"title":       feedTitle,
		"description": feedDescription,
		"link":        feedLink,
		"feedLink":    feedFeedLink,
		"updated":     feedUpdated,
		"published":   feedPublished,
		"language":    feedLanguage,
		"image":       feedImage,
		"copyright":   feedCopyright,
		"generator":   feedGenerator,
		"version":     feedFeedVersion,
		"categories":  feedCategories,
		"custom":      feedCustom,
	}
	mt := L.NewTypeMetatable(luaFeedTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), feedMethods))

	var personMethods = map[string]lua.LGFunction{
		"name":  personName,
		"email": personEmail,
	}
	mt = L.NewTypeMetatable(luaPersonTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), personMethods))

	return 0
}

//Feed

func newFeed(feed *gofeed.Feed, L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = feed
	L.SetMetatable(ud, L.GetTypeMetatable(luaFeedTypeName))
	return ud
}

func checkFeed(L *lua.LState) *gofeed.Feed {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*gofeed.Feed); ok {
		return v
	}
	L.ArgError(1, luaFeedTypeName+" expected")
	return nil
}

func feedTitle(L *lua.LState) int {
	feed := checkFeed(L)
	if L.GetTop() == 2 {
		feed.Title = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(feed.Title))
	return 1
}

func feedDescription(L *lua.LState) int {
	feed := checkFeed(L)
	if L.GetTop() == 2 {
		feed.Title = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(feed.Description))
	return 1
}

func feedLink(L *lua.LState) int {
	feed := checkFeed(L)
	if L.GetTop() == 2 {
		feed.Title = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(feed.Link))
	return 1
}

func feedFeedLink(L *lua.LState) int {
	feed := checkFeed(L)
	if L.GetTop() == 2 {
		feed.Title = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(feed.FeedLink))
	return 1
}

func feedUpdated(L *lua.LState) int {
	feed := checkFeed(L)
	if L.GetTop() == 2 {
		feed.Title = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(feed.Updated))
	return 1
}

func feedPublished(L *lua.LState) int {
	feed := checkFeed(L)
	if L.GetTop() == 2 {
		feed.Title = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(feed.Published))
	return 1
}

func feedLanguage(L *lua.LState) int {
	feed := checkFeed(L)
	if L.GetTop() == 2 {
		feed.Title = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(feed.Language))
	return 1
}

func feedImage(L *lua.LState) int {
	feed := checkFeed(L)
	if L.GetTop() == 2 {
		feed.Title = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(feed.Image.URL))
	return 1
}

func feedCopyright(L *lua.LState) int {
	feed := checkFeed(L)
	if L.GetTop() == 2 {
		feed.Title = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(feed.Copyright))
	return 1
}

func feedGenerator(L *lua.LState) int {
	feed := checkFeed(L)
	if L.GetTop() == 2 {
		feed.Title = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(feed.Generator))
	return 1
}

func feedFeedVersion(L *lua.LState) int {
	feed := checkFeed(L)
	if L.GetTop() == 2 {
		feed.Title = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(feed.FeedVersion))
	return 1
}

func feedAuthor(L *lua.LState) int {
	feed := checkFeed(L)
	if L.GetTop() == 2 {
		p := checkPerson(L, 2)
		feed.Author = p
		return 0
	}
	L.Push(newPerson(feed.Author, L))
	return 1
}

func feedCategories(L *lua.LState) int {
	feed := checkFeed(L)
	if L.GetTop() == 2 {
		cs := L.CheckTable(2)
		categories := make([]string, cs.Len())
		for i := 0; i < cs.Len(); i++ {
			v := cs.RawGetInt(i + 1)
			categories[i] = lua.LVAsString(v)
		}
		feed.Categories = categories
		return 0
	}
	cs := L.NewTable()
	for i, v := range feed.Categories {
		cs.RawSetInt(i+1, lua.LString(v))
	}
	L.Push(cs)
	return 1
}

func feedCustom(L *lua.LState) int {
	feed := checkFeed(L)
	if L.GetTop() == 2 {
		cs := L.CheckTable(2)
		custom := make(map[string]string)
		cs.ForEach(func(key, value lua.LValue) {
			custom[lua.LVAsString(key)] = lua.LVAsString(value)
		})
		feed.Custom = custom
		return 0
	}
	cs := L.NewTable()
	for k, v := range feed.Custom {
		cs.RawSetString(k, lua.LString(v))
	}
	L.Push(cs)
	return 1
}

//Person
func newPerson(p *gofeed.Person, L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = p
	L.SetMetatable(ud, L.GetTypeMetatable(luaPersonTypeName))
	return ud
}

func checkPerson(L *lua.LState, index int) *gofeed.Person {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*gofeed.Person); ok {
		return v
	}
	L.ArgError(1, luaPersonTypeName+" expected")
	return nil
}

func personName(L *lua.LState) int {
	p := checkPerson(L, 1)
	if L.GetTop() == 2 {
		p.Name = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(p.Name))
	return 1
}

func personEmail(L *lua.LState) int {
	p := checkPerson(L, 1)
	if L.GetTop() == 2 {
		p.Email = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(p.Email))
	return 1
}
