package keybindings

import (
	"fmt"
	"log"
	"strings"
	"unicode"
	"unicode/utf8"

	"git.sr.ht/~ghost08/photont/lib/states"
)

type Callback func() error

type KeyEvent struct {
	Key       rune
	Modifiers Modifiers
}

func (s KeyEvent) String() string {
	k := string(s.Key)
	switch s.Key {
	case '\u0008':
		k = "<backspace>"
	case '\t':
		k = "<tab>"
	case '\u00b1':
		k = "<esc>"
	case '\n':
		k = "<enter>"
	}
	return s.Modifiers.String() + k
}

type KeyEvents []KeyEvent

func (ss KeyEvents) String() string {
	var idents string
	for _, s := range ss {
		idents += s.String()
	}
	return idents
}

func parseState(s string) (string, KeyEvent, error) {
	ns, mod := modPrefix(s)
	if len(ns) == 0 {
		return "", KeyEvent{}, fmt.Errorf("not valid keybinding string: %s", s)
	}
	r, size := utf8.DecodeRuneInString(ns)
	r = unicode.ToLower(r)
	return ns[size:], KeyEvent{Key: r, Modifiers: mod}, nil
}

func modPrefix(s string) (string, Modifiers) {
	switch {
	case strings.HasPrefix(s, "<ctrl>"):
		return s[6:], ModCtrl
	case strings.HasPrefix(s, "<command>"):
		return s[9:], ModCommand
	case strings.HasPrefix(s, "<shift>"):
		return s[7:], ModShift
	case strings.HasPrefix(s, "<alt>"):
		return s[5:], ModAlt
	case strings.HasPrefix(s, "<super>"):
		return s[7:], ModSuper
	default:
		return s, 0
	}
}

func parseStates(s string) (KeyEvents, error) {
	var ss KeyEvents
	for {
		if len(s) == 0 {
			break
		}
		var err error
		var state KeyEvent
		s, state, err = parseState(s)
		if err != nil {
			return nil, err
		}
		ss = append(ss, state)
	}
	return ss, nil
}

type Registry struct {
	currentLayout states.Func
	reg           map[states.Enum]map[string]Callback
	currentState  KeyEvents
	repeat        int
}

func New(cl states.Func) *Registry {
	return &Registry{
		reg:           make(map[states.Enum]map[string]Callback),
		currentLayout: cl,
	}
}

func (kbr *Registry) Add(layout states.Enum, keyString string, callback Callback) {
	if _, ok := kbr.reg[layout]; !ok {
		kbr.reg[layout] = make(map[string]Callback)
	}
	ss, err := parseStates(keyString)
	if err != nil {
		log.Printf("ERROR: parsing keybinding string (%s): %s", keyString, err)
		return
	}
	kbr.reg[layout][ss.String()] = callback
}

func (kbr *Registry) Run(e KeyEvent) {
	cl := kbr.currentLayout()
	reg, ok := kbr.reg[cl]
	if !ok {
		return
	}
	if e.Modifiers == 0 && unicode.IsDigit(e.Key) && len(kbr.currentState) == 0 {
		kbr.repeat = kbr.repeat*10 + (int(e.Key) - '0')
	}
	kbr.currentState = append(kbr.currentState, e)
	ident := kbr.currentState.String()
	var hasPrefix bool
	var callback Callback
	for k, c := range reg {
		if !strings.HasPrefix(k, ident) {
			continue
		}
		hasPrefix = true
		if len(ident) == len(k) {
			callback = c
			break
		}
	}
	if !hasPrefix {
		kbr.currentState = nil
		return
	}
	if callback == nil {
		return
	}
	kbr.currentState = nil
	if kbr.repeat == 0 {
		if err := callback(); err != nil {
			log.Println("ERROR:", err)
		}
		return
	}
	for i := 0; i < kbr.repeat; i++ {
		if err := callback(); err != nil {
			log.Println("ERROR:", err)
			break
		}
	}
	kbr.repeat = 0
}
