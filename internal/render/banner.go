package render

import (
	"github.com/charmbracelet/lipgloss"
)

var bannerBaseStyle = lipgloss.NewStyle().Bold(true).Padding(0, 1).MarginTop(1).MarginBottom(1)

func Banner(status Status, text string) string {
	textColor := colorWhite
	if status.Color == colorYellow {
		textColor = colorBlack
	}

	return bannerBaseStyle.
		Foreground(textColor).
		Background(status.Color).
		Render(text)
}
