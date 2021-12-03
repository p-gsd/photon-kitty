package main

import (
	"git.sr.ht/~ghost08/libphoton"
	"git.sr.ht/~ghost08/libphoton/states"
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
		/*
			case searchEditor != nil && searchEditorFocus:
				return states.Search
		*/
	default:
		return states.Normal
	}
}

func (cb Callbacks) ArticleChanged(article *libphoton.Article) {
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
