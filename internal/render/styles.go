package render

import "github.com/charmbracelet/lipgloss"

type Status struct {
	Icon  string
	Color lipgloss.Color
	Style lipgloss.Style
}

var (
	colorGreen  = lipgloss.Color("2")
	colorRed    = lipgloss.Color("9")
	colorYellow = lipgloss.Color("11")
	colorCian   = lipgloss.Color("6")
	colorGray   = lipgloss.Color("8")
	colorWhite  = lipgloss.Color("15")
	colorBlack  = lipgloss.Color("0")
	colorPurple = lipgloss.Color("5")
)

var (
	Success = Status{Icon: "✓", Color: colorGreen, Style: lipgloss.NewStyle().Foreground(colorGreen).Bold(true)}
	Warning = Status{Icon: "⚠", Color: colorYellow, Style: lipgloss.NewStyle().Foreground(colorYellow).Bold(true)}
	Error   = Status{Icon: "✗", Color: colorRed, Style: lipgloss.NewStyle().Foreground(colorRed).Bold(true)}
	Info    = Status{Icon: "•", Color: colorCian, Style: lipgloss.NewStyle().Foreground(colorCian).Bold(true)}
)

var (
	topBarLabelStyle = lipgloss.NewStyle().Foreground(colorGray)
	topBarValueStyle = lipgloss.NewStyle().Foreground(colorCian).Bold(true)

	headerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPurple).
			Width(50).
			Align(lipgloss.Center).
			Bold(true).
			Foreground(colorWhite)

	sectionStyle = lipgloss.NewStyle().
			Foreground(colorCian).
			Bold(true).
			MarginTop(1).
			MarginBottom(0)

	boldStyle = lipgloss.NewStyle().Bold(true)
	dimStyle  = lipgloss.NewStyle().Foreground(colorGray)

	indentStyle = lipgloss.NewStyle().MarginLeft(2)

	keyStyle = lipgloss.NewStyle().
			Foreground(colorGray).
			Width(18)

	codeStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			Background(lipgloss.Color("235")).
			Padding(0, 1)
)
