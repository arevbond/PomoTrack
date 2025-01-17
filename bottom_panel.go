package main

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"
)

func constructBottomPanel() *tview.TextView {
	textPanel := hotKeysForPanel()
	panel := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter).
		SetText(textPanel)

	return panel
}

func hotKeysForPanel() string {
	type keyWithPage struct {
		key  string
		page string
	}

	keys := []keyWithPage{
		{"1", "Focus"},
		{"2", "Break"},
		{"3", "Detail"},
		{"4", "Summary"},
		{"5", "Tasks"},
	}

	strs := make([]string, 0, len(keys))
	for _, k := range keys {
		strs = append(strs, fmt.Sprintf("%s[:gray]%s[:-]", k.key, k.page))
	}
	return strings.Join(strs, " ")
}
