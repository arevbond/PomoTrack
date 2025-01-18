package main

import (
	"github.com/rivo/tview"
)

type Page struct {
	name PageName
	//item    tview.Primitive
	resize  bool
	visible bool

	render func() tview.Primitive
}

func NewPageComponent(name PageName, resize bool, visible bool,
	render func() tview.Primitive) *Page {
	return &Page{
		name: name,
		//item:    render(),
		resize:  resize,
		visible: visible,

		render: render,
	}
}

func (p *Page) WithBottomPanel() *Page {
	hotKeysPanel := constructBottomPanel()
	grid := tview.NewGrid().
		SetRows(0, 1).
		SetColumns(0, 23, 23, 0).
		SetBorders(true)
	grid.AddItem(p.render(), 0, 1, 1, 2, 0, 0, true)
	grid.AddItem(hotKeysPanel, 1, 1, 1, 2, 0, 0, false)

	p.render = func() tview.Primitive {
		return grid
	}
	return p
}
