package render

import (
	"fmt"
	"strings"
)

func HeaderBox(text string) string {
	return headerStyle.Render(text)
}

func TopBar(version, agentStatus, model, cloud string) string {
	parts := []string{
		fmt.Sprintf("%s %s", topBarLabelStyle.Render("🛡 Sentinel"), topBarValueStyle.Render(version)),
		fmt.Sprintf("%s %s", topBarLabelStyle.Render("Agent:"), topBarValueStyle.Render(agentStatus)),
		fmt.Sprintf("%s %s", topBarLabelStyle.Render("Model:"), topBarValueStyle.Render(model)),
		fmt.Sprintf("%s %s", topBarLabelStyle.Render("Cloud:"), topBarValueStyle.Render(cloud)),
	}
	return dimStyle.Render(strings.Join(parts, " │ "))
}

func Section(text string) string {
	return sectionStyle.Render(text)
}

func SubBlock(text string) string {
	return indentStyle.Render(text)
}

func Item(status Status, title, subtitle string) string {
	iconRendered := status.Style.Render(status.Icon)
	if subtitle == "" {
		return fmt.Sprintf("%s %s", iconRendered, boldStyle.Render(title))
	}
	return fmt.Sprintf("%s %s\n  %s", iconRendered, boldStyle.Render(title), dimStyle.Render(subtitle))
}

func KeyValue(key, value string) string {
	formattedKey := fmt.Sprintf("%s", key)
	return fmt.Sprintf("%s %s", keyStyle.Render(formattedKey), boldStyle.Render(value))
}

func Code(text string) string {
	return fmt.Sprintf("    %s", codeStyle.Render(text))
}

func InlineCode(text string) string {
	return codeStyle.Render(text)
}
