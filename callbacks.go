package main

import (
	"git.sr.ht/~ghost08/photont/lib"
	"git.sr.ht/~ghost08/photont/lib/states"
)

type Callbacks struct {
	grid *Grid
}

func (cb Callbacks) Redraw() {
	redraw(false)
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

func (cb Callbacks) SelectedCardMoveLeft() {
	cb.grid.SelectedChildMoveLeft()
}

func (cb Callbacks) SelectedCardMoveRight() {
	cb.grid.SelectedChildMoveRight()
}

func (cb Callbacks) SelectedCardMoveDown() {
	cb.grid.SelectedChildMoveDown()
}

func (cb Callbacks) SelectedCardMoveUp() {
	cb.grid.SelectedChildMoveUp()
}
