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
	return Move{grid: cb.grid}
}

type Move struct {
	grid *Grid
}

func (m Move) Left() {
	m.grid.SelectedChildMoveLeft()
	redraw(false)
}

func (m Move) Right() {
	m.grid.SelectedChildMoveRight()
	redraw(false)
}

func (m Move) Up() {
	m.grid.SelectedChildMoveUp()
	redraw(false)
}

func (m Move) Down() {
	m.grid.SelectedChildMoveDown()
	redraw(false)
}
