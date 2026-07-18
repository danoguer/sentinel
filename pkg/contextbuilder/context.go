package contextbuilder

type Context struct {
	System   SystemContext   `json:"system"`
	Logs     []LogEntry      `json:"logs"`
	Commands []CommandResult `json:"commands"`
	Files    []FileEntry     `json:"files"`
}

type SystemContext struct {
	OS          string `json:"os,omitempty"`
	LoadAverage string `json:"load_average,omitempty"`
	Memory      string `json:"memory,omitempty"`
	Uptime      string `json:"uptime,omitempty"`
	Hostname    string `json:"hostname,omitempty"`
}

type LogEntry struct {
	Source string `json:"source"`
	Line   string `json:"line"`
}

type CommandResult struct {
	Name   string `json:"name"`
	Output string `json:"output"`
}

type FileEntry struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}
