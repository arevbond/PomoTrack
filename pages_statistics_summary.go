package main

import (
	"fmt"
	"strconv"

	"github.com/rivo/tview"
)

func (m *UIManager) renderSummaryStatsPage(totalHours float64, totalDays int, weekdayHours [7]int) *PageComponent {
	table := tview.NewTable().
		SetBorders(true)

	table.SetCell(0, 0, tview.NewTableCell("Hours focused"))
	table.SetCell(1, 0, tview.NewTableCell(fmt.Sprintf("%.2f", totalHours)).SetAlign(tview.AlignCenter))

	table.SetCell(0, 1, tview.NewTableCell("Days accessed"))
	table.SetCell(1, 1, tview.NewTableCell(strconv.Itoa(totalDays)).SetAlign(tview.AlignCenter))

	bar := tview.NewTextView().
		SetDynamicColors(true).
		SetText("\n\n\n" + CreateBarGraph(weekdayHours))

	grid := tview.NewGrid().
		SetRows(5, 0).
		SetColumns(0, 40, 0).
		SetBorders(true)

	grid.AddItem(table, 0, 1, 1, 1, 0, 0, false)
	grid.AddItem(bar, 1, 1, 1, 1, 0, 0, false)

	return NewPageComponent(summaryStatsPage, grid, true, false)
}
