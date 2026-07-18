package contextbuilder

import "Sentinel/pkg/sanitize"

func (c *Context) Sanitize() {

	c.System.LoadAverage, _ = sanitize.SanitizeLine(c.System.LoadAverage)
	c.System.Memory, _ = sanitize.SanitizeLine(c.System.Memory)
	c.System.Uptime, _ = sanitize.SanitizeLine(c.System.Uptime)
	c.System.Hostname, _ = sanitize.SanitizeLine(c.System.Hostname)

	for i := range c.Commands {
		c.Commands[i].Output, _ = sanitize.SanitizeLine(c.Commands[i].Output)
	}

	for i := range c.Files {
		c.Files[i].Content, _ = sanitize.SanitizeLine(c.Files[i].Content)
	}


	var safeLogs []LogEntry
	for _, entry := range c.Logs {
		safeLine, _ := sanitize.SanitizeLine(entry.Line)
		if safeLine != "" {
			entry.Line = safeLine
			safeLogs = append(safeLogs, entry)
		}
	}
	c.Logs = safeLogs
}
