package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// tabDef holds the display data for a single tab.
type tabDef struct {
	icon  string
	title string
}

// tabs defines the ordered list of application tabs.
var tabs = []tabDef{
	{icon: "🔍", title: "Search"},
	{icon: "📦", title: "Installed"},
	{icon: "⬆", title: "Update"},
	{icon: "🧹", title: "Cleanup"},
}

// renderTabBar returns the rendered tab bar string for the given active tab index
// and total terminal width.
func renderTabBar(activeTab, width int) string {
	var renderedTabs []string

	for i, t := range tabs {
		label := fmt.Sprintf(" %s %s ", t.icon, t.title)
		if i == activeTab {
			renderedTabs = append(renderedTabs, ActiveTabStyle.Render(label))
		} else {
			renderedTabs = append(renderedTabs, InactiveTabStyle.Render(label))
		}
	}

	// Join tabs with a gap character
	row := strings.Join(renderedTabs, TabGapStyle.String())

	// Pad the bar to fill the full width
	gap := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(DarkGray).
		Width(max(width-lipgloss.Width(row), 0)).
		Render("")

	return lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)
}

// renderHeader renders the app title bar.
func renderHeader() string {
	title := TitleStyle.Render("🐧 Pacman TUI")
	subtitle := SubtitleStyle.Render(" — Arch Linux Package Manager")
	return lipgloss.JoinHorizontal(lipgloss.Center, title, subtitle)
}

// max returns the larger of two ints (for Go < 1.21 compatibility).
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
