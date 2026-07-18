package render

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func NewTable(headers []string, rows [][]string) *table.Table {
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(colorGray)).
		Headers(headers...).
		Rows(rows...)

	t.StyleFunc(func(row, col int) lipgloss.Style {
		if row == 0 {
			return lipgloss.NewStyle().Foreground(colorCian).Bold(true)
		}
		return lipgloss.NewStyle().Foreground(colorWhite)
	})

	return t
}
