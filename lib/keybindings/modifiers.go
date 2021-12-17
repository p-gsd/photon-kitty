package keybindings

import "strings"

// Modifiers
type Modifiers uint32

const (
	// ModCtrl is the ctrl modifier key.
	ModCtrl Modifiers = 1 << iota
	// ModCommand is the command modifier key
	// found on Apple keyboards.
	ModCommand
	// ModShift is the shift modifier key.
	ModShift
	// ModAlt is the alt modifier key, or the option
	// key on Apple keyboards.
	ModAlt
	// ModSuper is the "logo" modifier key, often
	// represented by a Windows logo.
	ModSuper
)

// Contain reports whether m contains all modifiers
// in m2.
func (m Modifiers) Contain(m2 Modifiers) bool {
	return m&m2 == m2
}

func (m Modifiers) String() string {
	var strs []string
	if m.Contain(ModCtrl) {
		strs = append(strs, "<ctrl>")
	}
	if m.Contain(ModCommand) {
		strs = append(strs, "<command>")
	}
	if m.Contain(ModShift) {
		strs = append(strs, "<shift>")
	}
	if m.Contain(ModAlt) {
		strs = append(strs, "<alt>")
	}
	if m.Contain(ModSuper) {
		strs = append(strs, "<super>")
	}
	return strings.Join(strs, "|")
}
