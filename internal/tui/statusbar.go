package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// statusBarModel holds the current state of the status bar.
type statusBarModel struct {
	message   string
	msgStyle  lipgloss.Style
	contextInfo string
	width     int
}

func newStatusBar() statusBarModel {
	return statusBarModel{
		msgStyle: InfoStyle,
	}
}

// setInfo sets an informational message in the status bar.
func (s *statusBarModel) setInfo(msg string) {
	s.message = msg
	s.msgStyle = InfoStyle
}

// setSuccess sets a success message in the status bar.
func (s *statusBarModel) setSuccess(msg string) {
	s.message = msg
	s.msgStyle = SuccessStyle
}

// setError sets an error message in the status bar.
func (s *statusBarModel) setError(msg string) {
	s.message = msg
	s.msgStyle = ErrorStyle
}

// setWarning sets a warning message in the status bar.
func (s *statusBarModel) setWarning(msg string) {
	s.message = msg
	s.msgStyle = WarningStyle
}

// clear removes any message from the status bar.
func (s *statusBarModel) clear() {
	s.message = ""
}

// render produces the status bar string at the given width.
func (s statusBarModel) render(width int) string {
	// Left side: context info + message
	left := ""
	if s.contextInfo != "" {
		left = StatusTextStyle.Render(s.contextInfo)
	}
	if s.message != "" {
		if left != "" {
			left += "  "
		}
		left += s.msgStyle.Render(s.message)
	}

	// Right side: key hints
	right := StatusKeyStyle.Render("tab") + StatusTextStyle.Render(" next tab  ") +
		StatusKeyStyle.Render("?") + StatusTextStyle.Render(" help  ") +
		StatusKeyStyle.Render("q") + StatusTextStyle.Render(" quit")

	// Calculate padding between left and right
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	padding := width - leftWidth - rightWidth - 2 // 2 for side padding
	if padding < 1 {
		padding = 1
	}

	bar := StatusBarStyle.Width(width).Render(
		left + strings.Repeat(" ", padding) + right,
	)

	return bar
}

// renderHints produces a compact one-line key hints bar.
func renderHints(hints []string) string {
	parts := make([]string, 0, len(hints))
	for i := 0; i+1 < len(hints); i += 2 {
		key := StatusKeyStyle.Render(hints[i])
		desc := StatusTextStyle.Render(" " + hints[i+1])
		parts = append(parts, fmt.Sprintf("%s%s", key, desc))
	}
	return strings.Join(parts, StatusTextStyle.Render("  "))
}
