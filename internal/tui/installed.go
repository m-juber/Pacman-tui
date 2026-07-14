package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"pacman-tui/internal/pacman"
)

// installedModel is the Installed Packages tab model.
type installedModel struct {
	keys     KeyMap
	filter   textinput.Model
	spinner  spinner.Model
	packages []pacman.Package
	filtered []pacman.Package
	cursor   int
	loading  bool
	errMsg   string
	width    int
	height   int
}

func newInstalledModel(keys KeyMap) installedModel {
	ti := textinput.New()
	ti.Placeholder = "Filter installed packages…"
	ti.CharLimit = 100
	ti.Width = 40
	ti.PromptStyle = SearchPromptStyle
	ti.Prompt = "🔎 "

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = SpinnerStyle

	return installedModel{
		keys:    keys,
		filter:  ti,
		spinner: sp,
		loading: true,
	}
}

// init loads the installed packages list.
func (m installedModel) init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, pacman.ListInstalled())
}

func (m installedModel) update(msg tea.Msg) (installedModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
			return m, nil
		}

	case pacman.InstalledListMsg:
		m.loading = false
		if msg.Err != nil {
			m.errMsg = msg.Err.Error()
		} else {
			m.packages = msg.Packages
			m.applyFilter()
		}
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	// Update filter input
	prevVal := m.filter.Value()
	var filterCmd tea.Cmd
	m.filter, filterCmd = m.filter.Update(msg)
	cmds = append(cmds, filterCmd)

	if m.filter.Value() != prevVal {
		m.applyFilter()
		m.cursor = 0
	}

	return m, tea.Batch(cmds...)
}

// applyFilter applies the current filter text to the packages list.
func (m *installedModel) applyFilter() {
	query := strings.ToLower(m.filter.Value())
	if query == "" {
		m.filtered = make([]pacman.Package, len(m.packages))
		copy(m.filtered, m.packages)
		return
	}
	m.filtered = m.filtered[:0]
	for _, pkg := range m.packages {
		if strings.Contains(strings.ToLower(pkg.Name), query) ||
			strings.Contains(strings.ToLower(pkg.Desc), query) {
			m.filtered = append(m.filtered, pkg)
		}
	}
}

// selectedPackage returns the currently highlighted package, or nil.
func (m installedModel) selectedPackage() *pacman.Package {
	if len(m.filtered) == 0 || m.cursor >= len(m.filtered) {
		return nil
	}
	p := m.filtered[m.cursor]
	return &p
}

// refresh reloads the installed packages list.
func (m *installedModel) refresh() tea.Cmd {
	m.loading = true
	m.packages = nil
	m.filtered = nil
	m.cursor = 0
	return tea.Batch(m.spinner.Tick, pacman.ListInstalled())
}

func (m installedModel) view(width, height int) string {
	var sb strings.Builder

	// Filter input
	filterRow := SearchStyle.Width(width - 4).Render(m.filter.View())
	sb.WriteString(filterRow + "\n\n")

	availHeight := height - 6

	if m.errMsg != "" {
		sb.WriteString(ErrorStyle.Render("✗ Error: " + m.errMsg))
		return sb.String()
	}

	if m.loading {
		loadMsg := fmt.Sprintf("%s  Loading installed packages…", m.spinner.View())
		centered := lipgloss.NewStyle().
			Width(width - 4).
			Align(lipgloss.Center).
			Padding(availHeight/2-1, 0).
			Render(loadMsg)
		sb.WriteString(centered)
		return sb.String()
	}

	if len(m.filtered) == 0 {
		msg := "No installed packages found."
		if m.filter.Value() != "" {
			msg = fmt.Sprintf("No packages matching %q", m.filter.Value())
		}
		centered := lipgloss.NewStyle().
			Foreground(DarkGray).
			Width(width - 4).
			Align(lipgloss.Center).
			Padding(availHeight/2-1, 0).
			Render(msg)
		sb.WriteString(centered)
		return sb.String()
	}

	// Count info
	total := len(m.packages)
	shown := len(m.filtered)
	info := ""
	if m.filter.Value() != "" {
		info = InfoStyle.Render(fmt.Sprintf("%d / %d packages", shown, total))
	} else {
		info = InfoStyle.Render(fmt.Sprintf("%d packages installed", total))
	}
	sb.WriteString(info + "\n\n")

	// Column header
	header := lipgloss.JoinHorizontal(lipgloss.Top,
		DetailKeyStyle.Render(fmt.Sprintf("%-30s", "Name")),
		DetailKeyStyle.Render(fmt.Sprintf("%-20s", "Version")),
		DetailKeyStyle.Render("Size"),
	)
	sb.WriteString(header + "\n")
	sb.WriteString(SubtitleStyle.Render(strings.Repeat("─", width-4)) + "\n")

	// Render visible rows
	maxVisible := availHeight - 4
	start := 0
	if m.cursor >= maxVisible {
		start = m.cursor - maxVisible + 1
	}
	end := start + maxVisible
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	for i := start; i < end; i++ {
		sb.WriteString(m.renderRow(i, width))
	}

	return sb.String()
}

func (m installedModel) renderRow(i, width int) string {
	pkg := m.filtered[i]
	selected := i == m.cursor

	name := pkg.Name
	if len(name) > 28 {
		name = name[:25] + "…"
	}

	version := pkg.Version
	if len(version) > 18 {
		version = version[:15] + "…"
	}

	size := pkg.InstalledSize
	if size == "" {
		size = "—"
	}

	cursor := "  "
	var rowStyle lipgloss.Style
	if selected {
		cursor = StatusKeyStyle.Render("▶ ")
		rowStyle = lipgloss.NewStyle().Background(DarkerGray).Width(width - 4)
	} else {
		rowStyle = lipgloss.NewStyle().Width(width - 4)
	}

	namePart := PkgNameStyle.Render(fmt.Sprintf("%-28s", name))
	versionPart := PkgVersionStyle.Render(fmt.Sprintf("%-18s", version))
	sizePart := PkgDescStyle.Render(size)

	line := cursor + namePart + "  " + versionPart + "  " + sizePart
	return rowStyle.Render(line) + "\n"
}

// hints returns the contextual key hints for the installed tab.
func (m installedModel) hints() string {
	h := []string{
		"↑↓", "navigate",
		"d", "remove",
		"i", "info",
		"r", "refresh",
	}
	return renderHints(h)
}

// contextInfo returns a summary string for the status bar.
func (m installedModel) contextInfo() string {
	if m.loading {
		return "loading…"
	}
	if m.filter.Value() != "" {
		return fmt.Sprintf("%d / %d packages", len(m.filtered), len(m.packages))
	}
	return fmt.Sprintf("%d packages installed", len(m.packages))
}
