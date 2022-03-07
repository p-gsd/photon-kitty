package main

import (
	"image"

	"git.sr.ht/~ghost08/photon/lib"
	"git.sr.ht/~ghost08/photon/lib/states"
)

type Callbacks struct {
	grid *Grid
}

func (cb Callbacks) Redraw() {
	redraw(false)
}

func (cb Callbacks) SelectedCard() *lib.Card {
	return SelectedCard
}

func (cb Callbacks) SelectedCardPos() image.Point {
	return SelectedCardPos
}

func (cb Callbacks) State() states.Enum {
	switch {
	case openedArticle != nil:
		return states.Article
	case command != "" && commandFocus:
		return states.Search
	default:
		return states.Normal
	}
}

func (cb Callbacks) ArticleChanged(article *lib.Article) {
	openedArticle = &Article{Article: article}
}

func (cb Callbacks) Move() lib.Move {
	return cb
}

func (cb Callbacks) Left() {
	cb.grid.SelectedChildMoveLeft()
	redraw(false)
}

func (cb Callbacks) Right() {
	cb.grid.SelectedChildMoveRight()
	redraw(false)
}

func (cb Callbacks) Up() {
	cb.grid.SelectedChildMoveUp()
	redraw(false)
}

func (cb Callbacks) Down() {
	cb.grid.SelectedChildMoveDown()
	redraw(false)
}
