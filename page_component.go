package main

import (
	"github.com/rivo/tview"
)

type PageComponent struct {
	name    PageName
	item    tview.Primitive
	resize  bool
	visible bool
}

func NewPageComponent(name PageName, item tview.Primitive, resize bool, visible bool) *PageComponent {
	return &PageComponent{
		name:    name,
		item:    item,
		resize:  resize,
		visible: visible,
	}
}

func (page *PageComponent) WithBottomPanel() *PageComponent {
	hotKeysPanel := constructBottomPanel()
	grid := tview.NewGrid().
		SetRows(0, 1).
		SetColumns(0, 23, 23, 0).
		SetBorders(true)
	grid.AddItem(page.item, 0, 1, 1, 2, 0, 0, true)
	grid.AddItem(hotKeysPanel, 1, 1, 1, 2, 0, 0, false)

	page.item = grid
	return page
}
