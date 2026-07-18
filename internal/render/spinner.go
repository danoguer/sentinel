package render

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

type Spinner struct {
	s *spinner.Spinner
}

func NewSpinner(message string) *Spinner {
	sp := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	sp.Suffix = " " + message + "..."
	_ = sp.Color("cyan")

	return &Spinner{s: sp}
}

func (s *Spinner) Start() { s.s.Start() }

func (s *Spinner) Stop() {
	s.s.Stop()
	fmt.Print("\r\033[K")
}
