package lib

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"git.sr.ht/~ghost08/photon/lib/events"
	"git.sr.ht/~ghost08/photon/lib/inputs"
	"git.sr.ht/~ghost08/photon/lib/keybindings"
	"git.sr.ht/~ghost08/photon/lib/media"
	"git.sr.ht/~ghost08/photon/lib/states"
	lua "github.com/yuin/gopher-lua"
)

var luaState *lua.LState

func (p *Photon) loadPlugins() error {
	plugins, err := findPlugins()
	if err != nil {
		return fmt.Errorf("finding plugins: %w", err)
	}
	if len(plugins) == 0 {
		return nil
	}
	p.initLuaState()
	for _, pluginPath := range plugins {
		if err := luaState.DoFile(pluginPath); err != nil {
			return fmt.Errorf("loading plugin: %s\n%s", pluginPath, err)
		}
	}
	return nil
}

func findPlugins() ([]string, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	pluginsDir := filepath.Join(usr.HomeDir, ".config", "photon", "plugins")
	if _, err := os.Stat(pluginsDir); os.IsNotExist(err) {
		return nil, nil
	}
	des, err := os.ReadDir(pluginsDir)
	if err != nil {
		return nil, err
	}
	var plugins []string
	for _, de := range des {
		if de.IsDir() || !strings.HasSuffix(de.Name(), ".lua") {
			continue
		}
		plugins = append(plugins, filepath.Join(pluginsDir, de.Name()))
	}
	return plugins, nil
}

func (p *Photon) initLuaState() {
	luaState = lua.NewState()
	media.Loader(luaState)
	p.cardsLoader(luaState)
	luaState.PreloadModule("photon", p.photonLoader)
	luaState.PreloadModule("photon.events", events.Loader)
	luaState.PreloadModule("photon.feedInputs", inputs.Loader(&p.feedInputs))
	luaState.PreloadModule("photon.keybindings", keybindings.Loader(p.KeyBindings))
}

func (p *Photon) photonLoader(L *lua.LState) int {
	var exports = map[string]lua.LGFunction{
		"state": p.photonState,
	}
	mod := L.SetFuncs(L.NewTable(), exports)
	p.registerTypeSelectedCard(L)
	L.SetField(mod, "Normal", lua.LNumber(states.Normal))
	L.SetField(mod, "Article", lua.LNumber(states.Article))
	L.SetField(mod, "Search", lua.LNumber(states.Search))
	L.SetField(mod, "cards", newCards(&p.Cards, L))
	L.SetField(mod, "visibleCards", newCards(&p.VisibleCards, L))
	L.SetField(mod, "selectedCard", p.newSelectedCard(L))
	L.Push(mod)

	return 1
}

func (p *Photon) photonState(L *lua.LState) int {
	L.Push(lua.LNumber(p.cb.State()))
	return 1
}

const luaSelectedCardTypeName = "photon.selectedCardType"

func (p *Photon) registerTypeSelectedCard(L *lua.LState) {
	var selectedCardMethods = map[string]lua.LGFunction{
		"posX": func(L *lua.LState) int {
			if L.GetTop() == 2 {
				p.SelectedCardPos.X = L.CheckInt(2)
				return 0
			}
			L.Push(lua.LNumber(p.SelectedCardPos.X))
			return 1
		},
		"posY": func(L *lua.LState) int {
			if L.GetTop() == 2 {
				p.SelectedCardPos.Y = L.CheckInt(2)
				return 0
			}
			L.Push(lua.LNumber(p.SelectedCardPos.X))
			return 1
		},
		"card": func(L *lua.LState) int {
			L.Push(newCard(p.SelectedCard, L))
			return 1
		},
	}
	newClass := L.SetFuncs(L.NewTable(), selectedCardMethods)
	mt := L.NewTypeMetatable(luaSelectedCardTypeName)
	L.SetField(mt, "__index", newClass)
}

func (p *Photon) newSelectedCard(L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = p.SelectedCard
	L.SetMetatable(ud, L.GetTypeMetatable(luaSelectedCardTypeName))
	return ud
}
