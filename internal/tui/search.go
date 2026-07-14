package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"pacman-tui/internal/pacman"
)

// searchTabState is the current state of the search tab.
type searchTabState int

const (
	searchStateIdle searchTabState = iota
	searchStateSearching
	searchStateResults
	searchStateError
)

// searchModel is the Search & Install tab model.
type searchModel struct {
	keys      KeyMap
	input     textinput.Model
	spinner   spinner.Model
	results   []pacman.Package
	cursor    int
	state     searchTabState
	errMsg    string
	lastQuery string
	width     int
	height    int
	debounce  *time.Timer
}

// debounceMsg is sent after a debounce delay.
type debounceMsg struct{ query string }

func newSearchModel(keys KeyMap) searchModel {
	ti := textinput.New()
	ti.Placeholder = "Search packages… (try: vim, firefox, python)"
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 40
	ti.PromptStyle = SearchPromptStyle
	ti.Prompt = "🔍 "

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = SpinnerStyle

	return searchModel{
		keys:    keys,
		input:   ti,
		spinner: sp,
		state:   searchStateIdle,
	}
}

// resetFocus restores input focus when the tab becomes active.
func (s *searchModel) resetFocus() {
	s.input.Focus()
}

func (s searchModel) update(msg tea.Msg) (searchModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, s.keys.Up):
			if len(s.results) > 0 && s.cursor > 0 {
				s.cursor--
			}
			return s, nil
		case key.Matches(msg, s.keys.Down):
			if len(s.results) > 0 && s.cursor < len(s.results)-1 {
				s.cursor++
			}
			return s, nil
		}

	case pacman.SearchResultMsg:
		s.state = searchStateResults
		if msg.Err != nil {
			s.state = searchStateError
			s.errMsg = msg.Err.Error()
		} else {
			s.results = msg.Packages
			s.cursor = 0
		}
		return s, nil

	case debounceMsg:
		if msg.query == s.input.Value() && msg.query != "" {
			s.state = searchStateSearching
			cmds = append(cmds, s.spinner.Tick, pacman.Search(msg.query))
		}
		return s, tea.Batch(cmds...)

	case spinner.TickMsg:
		if s.state == searchStateSearching {
			var cmd tea.Cmd
			s.spinner, cmd = s.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	// Update text input
	prevVal := s.input.Value()
	var inputCmd tea.Cmd
	s.input, inputCmd = s.input.Update(msg)
	cmds = append(cmds, inputCmd)

	// Debounce search when query changes
	if s.input.Value() != prevVal {
		newQuery := s.input.Value()
		if newQuery == "" {
			s.state = searchStateIdle
			s.results = nil
			s.cursor = 0
		} else {
			// Schedule debounced search
			cmds = append(cmds, func() tea.Msg {
				time.Sleep(300 * time.Millisecond)
				return debounceMsg{query: newQuery}
			})
		}
	}

	return s, tea.Batch(cmds...)
}

// selectedPackage returns the currently highlighted package, or nil.
func (s searchModel) selectedPackage() *pacman.Package {
	if len(s.results) == 0 || s.cursor >= len(s.results) {
		return nil
	}
	p := s.results[s.cursor]
	return &p
}

func (s searchModel) view(width, height int) string {
	var sb strings.Builder

	// Search input
	searchRow := SearchStyle.Width(width - 4).Render(s.input.View())
	sb.WriteString(searchRow + "\n\n")

	availHeight := height - 6 // reserve rows for input, hints, borders

	switch s.state {
	case searchStateIdle:
		placeholder := lipgloss.NewStyle().
			Foreground(DarkGray).
			Width(width - 4).
			Align(lipgloss.Center).
			Padding(availHeight/2-1, 0).
			Render("Type to search packages…\n\nUse / to focus the search bar")
		sb.WriteString(placeholder)

	case searchStateSearching:
		loadMsg := fmt.Sprintf("%s  Searching for %q…", s.spinner.View(), s.input.Value())
		centered := lipgloss.NewStyle().
			Width(width - 4).
			Align(lipgloss.Center).
			Padding(availHeight/2-1, 0).
			Render(loadMsg)
		sb.WriteString(centered)

	case searchStateError:
		errView := ErrorStyle.Render("✗ " + s.errMsg)
		centered := lipgloss.NewStyle().
			Width(width - 4).
			Align(lipgloss.Center).
			Padding(availHeight/2-1, 0).
			Render(errView)
		sb.WriteString(centered)

	case searchStateResults:
		if len(s.results) == 0 {
			noResults := lipgloss.NewStyle().
				Foreground(DarkGray).
				Width(width - 4).
				Align(lipgloss.Center).
				Padding(availHeight/2-1, 0).
				Render(fmt.Sprintf("No packages found for %q", s.input.Value()))
			sb.WriteString(noResults)
		} else {
			// Count line
			count := InfoStyle.Render(fmt.Sprintf("%d results", len(s.results)))
			sb.WriteString(count + "\n\n")

			// Render visible items
			maxVisible := availHeight - 2
			start := 0
			if s.cursor >= maxVisible {
				start = s.cursor - maxVisible + 1
			}
			end := start + maxVisible
			if end > len(s.results) {
				end = len(s.results)
			}

			for i := start; i < end; i++ {
				sb.WriteString(s.renderItem(i, width))
			}
		}
	}

	return sb.String()
}

func (s searchModel) renderItem(i, width int) string {
	pkg := s.results[i]
	selected := i == s.cursor

	// Build repo badge
	repo := pkg.Repository
	if repo == "" {
		repo = "unknown"
	}
	badge := RepoBadge(repo)

	// Installed marker
	installedStr := ""
	if pkg.IsInstalled {
		installedStr = "  " + InstalledBadge.Render("✓")
	}

	// Truncate description
	desc := pkg.Desc
	maxDescLen := width - 40
	if maxDescLen < 20 {
		maxDescLen = 20
	}
	if len(desc) > maxDescLen {
		desc = desc[:maxDescLen-1] + "…"
	}

	namePart := PkgNameStyle.Render(pkg.Name)
	versionPart := PkgVersionStyle.Render("  " + pkg.Version)
	descPart := "  " + PkgDescStyle.Render(desc)
	right := badge + installedStr

	// Title line
	titleLine := lipgloss.JoinHorizontal(lipgloss.Center,
		namePart, versionPart, descPart,
	)

	cursor := "  "
	var style lipgloss.Style
	if selected {
		cursor = StatusKeyStyle.Render("▶ ")
		style = lipgloss.NewStyle().
			Background(DarkerGray).
			Width(width - 4)
	} else {
		style = lipgloss.NewStyle().Width(width - 4)
	}

	// Pad title to make room for right badge
	titleWidth := lipgloss.Width(titleLine)
	rightWidth := lipgloss.Width(right)
	padWidth := width - 4 - len(cursor) - titleWidth - rightWidth
	if padWidth < 0 {
		padWidth = 0
	}
	pad := strings.Repeat(" ", padWidth)

	line := cursor + titleLine + pad + right
	return style.Render(line) + "\n"
}

// hints returns the contextual key hints for the search tab.
func (s searchModel) hints() string {
	h := []string{
		"↑↓", "navigate",
		"enter", "install",
		"i", "info",
		"d", "remove",
	}
	return renderHints(h)
}

// contextInfo returns a summary string for the status bar.
func (s searchModel) contextInfo() string {
	switch s.state {
	case searchStateResults:
		if len(s.results) > 0 {
			return fmt.Sprintf("%d results", len(s.results))
		}
		return "no results"
	case searchStateSearching:
		return fmt.Sprintf("searching %q…", s.lastQuery)
	default:
		return ""
	}
}
