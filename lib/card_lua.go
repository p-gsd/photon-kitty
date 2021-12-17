package lib

import (
	"git.sr.ht/~ghost08/photont/lib/media"
	lua "github.com/yuin/gopher-lua"
)

const (
	luaCardTypeName  = "photon.card"
	luaCardsTypeName = "photon.cards"
)

func (p *Photon) cardsLoader(L *lua.LState) int {
	var cardMethods = map[string]lua.LGFunction{
		"link":        cardItemLink,
		"image":       cardItemImage,
		"title":       cardItemTitle,
		"content":     cardItemContent,
		"description": cardItemDescription,
		"published":   cardItemPublished,
		"feed":        cardFeed,
		"getMedia":    getMedia,
		"runMedia": func(L *lua.LState) int {
			card := checkCard(L, 1)
			card.RunMedia()
			return 0
		},
		"openBrowser": func(L *lua.LState) int {
			card := checkCard(L, 1)
			card.OpenBrowser()
			return 0
		},
		"openArticle": func(L *lua.LState) int {
			card := checkCard(L, 1)
			card.OpenArticle()
			return 0
		},
	}
	mt := L.NewTypeMetatable(luaCardTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), cardMethods))

	var cardsMethods = map[string]lua.LGFunction{
		"len": cardsLength,
		"get": cardsGet,
		"set": cardsSet,
	}
	mt = L.NewTypeMetatable(luaCardsTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), cardsMethods))

	feedLoader(L)
	return 0
}

func checkCard(L *lua.LState, index int) *Card {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*Card); ok {
		return v
	}
	L.ArgError(1, luaCardTypeName+" expected")
	return nil
}

func newCard(card *Card, L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = card
	L.SetMetatable(ud, L.GetTypeMetatable(luaCardTypeName))
	return ud
}

func cardItemLink(L *lua.LState) int {
	card := checkCard(L, 1)
	if L.GetTop() == 2 {
		card.Item.Link = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(card.Item.Link))
	return 1
}

func cardItemImage(L *lua.LState) int {
	card := checkCard(L, 1)
	if L.GetTop() == 2 {
		card.Item.Image.URL = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(card.Item.Image.URL))
	return 1
}

func cardItemPublished(L *lua.LState) int {
	card := checkCard(L, 1)
	if L.GetTop() == 2 {
		card.Item.Published = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(card.Item.Published))
	return 1
}

func cardItemTitle(L *lua.LState) int {
	card := checkCard(L, 1)
	if L.GetTop() == 2 {
		card.Item.Title = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(card.Item.Title))
	return 1
}

func cardItemContent(L *lua.LState) int {
	card := checkCard(L, 1)
	if L.GetTop() == 2 {
		card.Item.Content = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(card.Item.Content))
	return 1
}

func cardItemDescription(L *lua.LState) int {
	card := checkCard(L, 1)
	if L.GetTop() == 2 {
		card.Item.Description = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(card.Item.Description))
	return 1
}

func getMedia(L *lua.LState) int {
	card := checkCard(L, 1)
	m, err := card.GetMedia()
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	ud := media.NewLuaMedia(m, L)
	L.Push(ud)
	L.Push(lua.LNil)
	return 2
}

func cardFeed(L *lua.LState) int {
	card := checkCard(L, 1)
	L.Push(newFeed(card.Feed, L))
	return 1
}

//Cards

func newCards(cards *Cards, L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = cards
	L.SetMetatable(ud, L.GetTypeMetatable(luaCardsTypeName))
	return ud
}

func checkCards(L *lua.LState) *Cards {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*Cards); ok {
		return v
	}
	L.ArgError(1, luaCardsTypeName+" expected")
	return nil
}

func cardsLength(L *lua.LState) int {
	cards := checkCards(L)
	L.Push(lua.LNumber(len(*cards)))
	return 1
}

func cardsGet(L *lua.LState) int {
	cards := checkCards(L)
	index := L.ToInt(2)
	card := (*cards)[index-1]
	L.Push(newCard(card, L))
	return 1
}

func cardsSet(L *lua.LState) int {
	cards := checkCards(L)
	index := L.ToInt(2)
	card := checkCard(L, 3)
	(*cards)[index-1] = card
	return 0
}
