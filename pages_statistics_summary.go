package main

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/rivo/tview"
)

func (m *UIManager) NewSummaryPage() *Page {
	pomodoros, err := m.pomodoroTracker.Pomodoros()
	if err != nil {
		m.logger.Error("can't get all pomodoros", slog.Any("error", err))
		return nil
	}

	render := m.renderSummaryStatsPage(
		m.pomodoroTracker.Hours(pomodoros),
		m.pomodoroTracker.CountDays(pomodoros),
		m.pomodoroTracker.HoursInWeek(pomodoros),
	)

	return NewPageComponent(summaryStatsPage, true, false, render)
}

func (m *UIManager) renderSummaryStatsPage(args ...any) func() tview.Primitive {
	totalHours, ok := args[0].(float64)
	if !ok {
		m.logger.Error("can't extract argument for rendering summary stats page",
			slog.String("func", "render summary stats page"))
		return nil
	}
	totalDays, ok := args[1].(int)
	if !ok {
		m.logger.Error("can't extract argument for rendering summary stats page",
			slog.String("func", "render summary stats page"))
		return nil
	}
	weekdayHours, ok := args[2].([7]int)
	if !ok {
		m.logger.Error("can't extract argument for rendering summary stats page",
			slog.String("func", "render summary stats page"))
		return nil
	}

	return func() tview.Primitive {
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

		return grid
	}
}
