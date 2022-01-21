package main

import (
	"git.sr.ht/~ghost08/photon/lib"
	"git.sr.ht/~ghost08/photon/lib/states"
	"github.com/gdamore/tcell/v2"
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

func (cb Callbacks) Status(text string) {
	if text == "" {
		statusBarText = nil
	} else {
		statusBarText = richtext{{Text: text, Style: tcell.StyleDefault}}
	}
	redraw(false)
}
