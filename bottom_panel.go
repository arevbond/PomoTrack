package main

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"
)

func constructBottomPanel(curPage PageName) *tview.TextView {
	textPanel := hotKeysForPanel(curPage)
	panel := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter).
		SetText(textPanel)

	return panel
}

func hotKeysForPanel(pageName PageName) string {
	type keyWithPage struct {
		key            string
		prettyPageName string

		insidePages []PageName
	}

	keys := []keyWithPage{
		{"1", "Focus", []PageName{pauseFocusPage, activeFocusPage}},
		{"2", "Break", []PageName{pauseBreakPage, activeBreakPage}},
		{"3", "Tasks", []PageName{allTasksPage, addNewTaskPage, deleteTaskPage}},
		{"4", "Summary", []PageName{summaryStatsPage}},
		{"5", "Detail", []PageName{detailStatsPage, insertStatsPage}},
	}

	strs := make([]string, 0, len(keys))
	for _, k := range keys {
		str := fmt.Sprintf("%s[:gray]%s[:-]", k.key, k.prettyPageName)
		if containPage(pageName, k.insidePages) {
			str = fmt.Sprintf("%s[:brown]%s[:-]", k.key, k.prettyPageName)
		}
		strs = append(strs, str)
	}
	return strings.Join(strs, " ")
}

func containPage(target PageName, pages []PageName) bool {
	for _, page := range pages {
		if target == page {
			return true
		}
	}
	return false
}
