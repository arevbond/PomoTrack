package main

import (
	"fmt"
	"strings"
	"time"
)

func CreateBarGraph(data [7]int) string {
	var graph strings.Builder
	maxValue := 6
	for _, value := range data {
		maxValue = max(maxValue, value)
	}

	for i := maxValue; i > 0; i-- {
		graph.WriteString(fmt.Sprintf("%02d| ", i))
		for _, v := range data {
			if v >= i {
				graph.WriteString("##   ")
			} else {
				graph.WriteString("     ")
			}
		}
		graph.WriteString("\n")
	}

	graph.WriteString("    ")
	for i, day := range []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"} {
		if int((time.Now().Weekday()+6)%7) == i {
			day = fmt.Sprintf("[green]%s[-]", day)
		}
		graph.WriteString(day + "  ")
	}
	graph.WriteString("\n    ---------------------------------")
	graph.WriteString("\n")

	return graph.String()
}
