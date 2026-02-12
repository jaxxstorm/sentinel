package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jaxxstorm/sentinel/internal/event"
)

type Renderer struct {
	noColor bool
	hdr     lipgloss.Style
	ok      lipgloss.Style
	warn    lipgloss.Style
}

func NewRenderer(noColor bool) *Renderer {
	if os.Getenv("NO_COLOR") != "" {
		noColor = true
	}
	hdr := lipgloss.NewStyle().Bold(true)
	ok := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	warn := lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	if noColor {
		hdr = lipgloss.NewStyle().Bold(true)
		ok = lipgloss.NewStyle()
		warn = lipgloss.NewStyle()
	}
	return &Renderer{noColor: noColor, hdr: hdr, ok: ok, warn: warn}
}

func (r *Renderer) FormatDiff(events []event.Event) string {
	var b strings.Builder
	b.WriteString(r.hdr.Render("Sentinel Diff"))
	b.WriteString("\n")
	if len(events) == 0 {
		b.WriteString(r.warn.Render("No changes detected"))
		return b.String()
	}
	for _, evt := range events {
		line := fmt.Sprintf("- %s %s (%s)", evt.EventType, evt.SubjectID, evt.Timestamp.Format("2006-01-02 15:04:05"))
		b.WriteString(r.ok.Render(line))
		b.WriteString("\n")
	}
	return strings.TrimSuffix(b.String(), "\n")
}
