package lib

import (
	"github.com/mmcdole/gofeed"
	lua "github.com/yuin/gopher-lua"
)

const (
	luaCardsTypeName = "photon.cards"
)

func (p *Photon) cardsLoader(L *lua.LState) int {
	p.cardLoader(L)
	var cardsMethods = map[string]lua.LGFunction{
		"len":    cardsLength,
		"get":    cardsGet,
		"set":    cardsSet,
		"del":    cardsDel,
		"add":    cardsAdd,
		"append": cardsAppend,
		"create": cardsCreate,
	}
	mt := L.NewTypeMetatable(luaCardsTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), cardsMethods))
	feedLoader(L)
	return 0
}

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

func cardsDel(L *lua.LState) int {
	cards := checkCards(L)
	index := L.ToInt(2)
	newCards := append((*cards)[:index-1], (*cards)[index:]...)
	*cards = newCards
	return 0
}

func cardsAdd(L *lua.LState) int {
	cards := checkCards(L)
	index := L.ToInt(2)
	card := checkCard(L, 3)
	(*cards) = append((*cards)[:index-1], append([]*Card{card}, (*cards)[index-1:]...)...)
	return 0
}

func cardsAppend(L *lua.LState) int {
	cards := checkCards(L)
	card := checkCard(L, 2)
	(*cards) = append((*cards), card)
	return 0
}

func cardsCreate(L *lua.LState) int {
	cardTable := L.CheckTable(1)
	card := &Card{Item: &gofeed.Item{}}
	card.Item.Link = cardTable.RawGetString("link").String()
	card.Item.Title = cardTable.RawGetString("title").String()
	card.Item.Content = cardTable.RawGetString("content").String()
	card.Item.Description = cardTable.RawGetString("description").String()
	card.Item.Published = cardTable.RawGetString("published").String()
	if img := cardTable.RawGetString("image").String(); img != "" {
		card.Item.Image = &gofeed.Image{URL: img}
	}
	feed := cardTable.RawGetString("feed").(*lua.LTable)
	if feed != nil {
		feedCreate(feed)
	}
	L.Push(newCard(card, L))
	return 1
}

func feedCreate(feedTable *lua.LTable) *gofeed.Feed {
	feed := &gofeed.Feed{}
	feed.Title = feedTable.RawGetString("title").String()
	feed.Description = feedTable.RawGetString("description").String()
	feed.Link = feedTable.RawGetString("link").String()
	feed.FeedLink = feedTable.RawGetString("feedLink").String()
	feed.Updated = feedTable.RawGetString("updated").String()
	feed.Published = feedTable.RawGetString("published").String()
	feed.Language = feedTable.RawGetString("language").String()
	feed.Copyright = feedTable.RawGetString("copyright").String()
	feed.Generator = feedTable.RawGetString("generator").String()
	feed.FeedVersion = feedTable.RawGetString("version").String()
	if cats, ok := feedTable.RawGetString("categories").(*lua.LTable); ok && cats != nil {
		cats.ForEach(func(_, v lua.LValue) {
			feed.Categories = append(feed.Categories, v.String())
		})
	}
	if img := feedTable.RawGetString("image").String(); img != "" {
		feed.Image = &gofeed.Image{URL: img}
	}
	return feed
}
