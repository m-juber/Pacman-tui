package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// ── Color Palette ───────────────────────────────────────────────────────────

var (
	// Primary – deep purple/indigo
	Purple     = lipgloss.Color("#7C3AED")
	DarkPurple = lipgloss.Color("#6D28D9")

	// Accent – cyan/teal
	Cyan      = lipgloss.Color("#06B6D4")
	LightCyan = lipgloss.Color("#22D3EE")

	// Semantic
	Orange = lipgloss.Color("#F97316")
	Green  = lipgloss.Color("#10B981")
	Red    = lipgloss.Color("#EF4444")

	// Neutrals
	White      = lipgloss.Color("#F8FAFC")
	Gray       = lipgloss.Color("#94A3B8")
	DarkGray   = lipgloss.Color("#475569")
	DarkerGray = lipgloss.Color("#1E293B")
	Darkest    = lipgloss.Color("#0F172A")
)

// ── Styles ──────────────────────────────────────────────────────────────────

var (
	// App chrome
	AppStyle      lipgloss.Style
	TitleStyle    lipgloss.Style
	SubtitleStyle lipgloss.Style

	// Tab bar
	ActiveTabStyle   lipgloss.Style
	InactiveTabStyle lipgloss.Style
	TabGapStyle      lipgloss.Style
	TabBarStyle      lipgloss.Style

	// Content area
	ContentStyle lipgloss.Style

	// Status bar
	StatusBarStyle   lipgloss.Style
	StatusTextStyle  lipgloss.Style
	StatusKeyStyle   lipgloss.Style
	StatusValueStyle lipgloss.Style

	// Package list items
	PkgNameStyle   lipgloss.Style
	PkgVersionStyle lipgloss.Style
	PkgDescStyle   lipgloss.Style
	InstalledBadge lipgloss.Style
	AURBadge       lipgloss.Style

	// Package detail overlay
	DetailStyle      lipgloss.Style
	DetailTitleStyle lipgloss.Style
	DetailKeyStyle   lipgloss.Style
	DetailValStyle   lipgloss.Style

	// Confirmation dialog
	DialogStyle       lipgloss.Style
	DialogTitleStyle  lipgloss.Style
	DialogBtnActive   lipgloss.Style
	DialogBtnInactive lipgloss.Style

	// Search
	SearchStyle       lipgloss.Style
	SearchPromptStyle lipgloss.Style

	// Messages
	SuccessStyle lipgloss.Style
	ErrorStyle   lipgloss.Style
	WarningStyle lipgloss.Style
	InfoStyle    lipgloss.Style

	// Spinner
	SpinnerStyle lipgloss.Style

	// Help
	HelpStyle lipgloss.Style
)

func init() {
	// ── App chrome ──────────────────────────────────────────────────────

	AppStyle = lipgloss.NewStyle().
		Padding(1, 2)

	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(Purple).
		PaddingLeft(1)

	SubtitleStyle = lipgloss.NewStyle().
		Foreground(Gray)

	// ── Tab bar ─────────────────────────────────────────────────────────

	ActiveTabStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(White).
		Background(Purple).
		Padding(0, 2)

	InactiveTabStyle = lipgloss.NewStyle().
		Foreground(Gray).
		Padding(0, 2)

	TabGapStyle = lipgloss.NewStyle().
		Foreground(DarkGray).
		SetString("│")

	TabBarStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(DarkGray)

	// ── Content area ────────────────────────────────────────────────────

	ContentStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(DarkGray)

	// ── Status bar ──────────────────────────────────────────────────────

	StatusBarStyle = lipgloss.NewStyle().
		Background(Darkest).
		Padding(0, 1)

	StatusTextStyle = lipgloss.NewStyle().
		Foreground(Gray)

	StatusKeyStyle = lipgloss.NewStyle().
		Foreground(Cyan).
		Bold(true)

	StatusValueStyle = lipgloss.NewStyle().
		Foreground(White)

	// ── Package list items ──────────────────────────────────────────────

	PkgNameStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(White)

	PkgVersionStyle = lipgloss.NewStyle().
		Foreground(Purple)

	PkgDescStyle = lipgloss.NewStyle().
		Foreground(Gray)

	InstalledBadge = lipgloss.NewStyle().
		Foreground(Green)

	AURBadge = lipgloss.NewStyle().
		Background(Orange).
		Foreground(Darkest).
		Padding(0, 1)

	// ── Package detail overlay ──────────────────────────────────────────

	DetailStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(Purple).
		Padding(1, 2)

	DetailTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(Purple).
		MarginBottom(1)

	DetailKeyStyle = lipgloss.NewStyle().
		Foreground(Cyan).
		Bold(true)

	DetailValStyle = lipgloss.NewStyle().
		Foreground(White)

	// ── Confirmation dialog ─────────────────────────────────────────────

	DialogStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(Orange).
		Padding(1, 3).
		Align(lipgloss.Center)

	DialogTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(White)

	DialogBtnActive = lipgloss.NewStyle().
		Background(Purple).
		Foreground(White).
		Padding(0, 2)

	DialogBtnInactive = lipgloss.NewStyle().
		Background(DarkGray).
		Foreground(Gray).
		Padding(0, 2)

	// ── Search ──────────────────────────────────────────────────────────

	SearchStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(DarkGray)

	SearchPromptStyle = lipgloss.NewStyle().
		Foreground(Cyan).
		Bold(true)

	// ── Messages ────────────────────────────────────────────────────────

	SuccessStyle = lipgloss.NewStyle().
		Foreground(Green)

	ErrorStyle = lipgloss.NewStyle().
		Foreground(Red).
		Bold(true)

	WarningStyle = lipgloss.NewStyle().
		Foreground(Orange)

	InfoStyle = lipgloss.NewStyle().
		Foreground(Cyan)

	// ── Spinner ─────────────────────────────────────────────────────────

	SpinnerStyle = lipgloss.NewStyle().
		Foreground(Purple)

	// ── Help ────────────────────────────────────────────────────────────

	HelpStyle = lipgloss.NewStyle().
		Foreground(Gray).
		Padding(1, 2)
}

// ── Helpers ─────────────────────────────────────────────────────────────────

// RepoBadge returns a styled, color-coded badge for the given pacman repository name.
func RepoBadge(repo string) string {
	var bg lipgloss.Color
	switch repo {
	case "core":
		bg = lipgloss.Color("#059669") // emerald-600
	case "extra":
		bg = lipgloss.Color("#2563EB") // blue-600
	case "community":
		bg = lipgloss.Color("#7C3AED") // violet-600
	case "multilib":
		bg = lipgloss.Color("#CA8A04") // yellow-600
	case "aur":
		bg = lipgloss.Color("#EA580C") // orange-600
	default:
		bg = lipgloss.Color("#475569") // slate-600
	}

	style := lipgloss.NewStyle().
		Background(bg).
		Foreground(lipgloss.Color("#0F172A")).
		Bold(true).
		Padding(0, 1)

	return style.Render(fmt.Sprintf(" %s ", repo))
}
