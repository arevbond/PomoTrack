package main

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"
)

func constructBottomPanel(pageName string) *tview.TextView {
	textPanel := hotKeysForPanel(pageName)
	panel := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter).
		SetText(textPanel)

	return panel
}

func hotKeysForPanel(pageName string) string {
	type keyWithPage struct {
		key  string
		page string
	}

	var keys []keyWithPage
	switch pageName {
	case detailStatsPage:
		keys = []keyWithPage{
			{"CtrlI", "Insert"},
		}
	default:
		keys = []keyWithPage{
			{"1", "Focus"},
			{"2", "Break"},
			{"3", "Detail"},
			{"4", "Summary"},
		}
	}

	strs := make([]string, 0, len(keys))
	for _, k := range keys {
		strs = append(strs, fmt.Sprintf("%s[:blue]%s[:-]", k.key, k.page))
	}
	return strings.Join(strs, " ")
}
