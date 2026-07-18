package transport

import "Sentinel/pkg/contextbuilder"

type Context struct {
	System map[string]string `json:"system,omitempty"`

	Logs map[string][]string `json:"logs,omitempty"`

	Files map[string]string `json:"files,omitempty"`

	Commands map[string]string `json:"commands,omitempty"`
}

type ContextEnvelope struct {
	Command  string  `json:"command"`
	Target   string  `json:"target,omitempty"`
	Question string  `json:"question"`
	CWD      string  `json:"cwd"`
	Context  contextbuilder.Context `json:"context"`
}

type AgentResponse struct {
	Answer     string `json:"answer"`
	DurationMS int64  `json:"duration_ms"`
	TokensUsed int    `json:"tokens_used,omitempty"`
	Error      string `json:"error,omitempty"`
}
