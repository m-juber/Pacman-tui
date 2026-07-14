package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"pacman-tui/internal/pacman"
)

// cleanupTabState is the current state of the cleanup tab.
type cleanupTabState int

const (
	cleanupStateIdle cleanupTabState = iota
	cleanupStateLoading
	cleanupStateReady
	cleanupStateCleaning
	cleanupStateDone
	cleanupStateError
)

// cleanupModel is the Orphan Cleanup tab model.
type cleanupModel struct {
	keys    KeyMap
	spinner spinner.Model
	orphans []pacman.Package
	cursor  int
	selected map[int]bool
	state   cleanupTabState
	errMsg  string
	doneMsg string
}

func newCleanupModel(keys KeyMap) cleanupModel {
	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = SpinnerStyle

	return cleanupModel{
		keys:     keys,
		spinner:  sp,
		state:    cleanupStateIdle,
		selected: make(map[int]bool),
	}
}

// init loads the orphan list when the tab becomes active.
func (m cleanupModel) init() tea.Cmd {
	m.state = cleanupStateLoading
	return tea.Batch(m.spinner.Tick, pacman.ListOrphans())
}

func (m cleanupModel) update(msg tea.Msg) (cleanupModel, tea.Cmd) {
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
			if m.cursor < len(m.orphans)-1 {
				m.cursor++
			}
			return m, nil
		case key.Matches(msg, m.keys.Select):
			if len(m.orphans) > 0 {
				m.selected[m.cursor] = !m.selected[m.cursor]
			}
			return m, nil
		case key.Matches(msg, m.keys.Refresh):
			m.state = cleanupStateLoading
			m.orphans = nil
			m.selected = make(map[int]bool)
			m.cursor = 0
			return m, tea.Batch(m.spinner.Tick, pacman.ListOrphans())
		}

	case pacman.OrphanListMsg:
		if msg.Err != nil {
			m.state = cleanupStateError
			m.errMsg = msg.Err.Error()
		} else {
			m.state = cleanupStateReady
			m.orphans = msg.Packages
		}
		return m, nil

	case pacman.OperationFinishedMsg:
		if msg.Operation == "remove-orphans" || msg.Operation == "remove" {
			if msg.Err != nil {
				m.state = cleanupStateError
				m.errMsg = msg.Err.Error()
			} else {
				m.state = cleanupStateDone
				m.doneMsg = "Orphans removed successfully!"
				m.orphans = nil
				m.selected = make(map[int]bool)
			}
		}
		return m, nil

	case spinner.TickMsg:
		if m.state == cleanupStateLoading || m.state == cleanupStateCleaning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// removeAllOrphans starts removing all orphan packages.
func (m *cleanupModel) removeAllOrphans() tea.Cmd {
	m.state = cleanupStateCleaning
	return tea.Batch(m.spinner.Tick, pacman.RemoveOrphans())
}

// removeSelected removes only the selected orphans.
func (m *cleanupModel) removeSelected() tea.Cmd {
	if len(m.selected) == 0 {
		return nil
	}
	var names []string
	for idx := range m.selected {
		if idx < len(m.orphans) {
			names = append(names, m.orphans[idx].Name)
		}
	}
	if len(names) == 0 {
		return nil
	}
	m.state = cleanupStateCleaning
	return tea.Batch(m.spinner.Tick, pacman.Remove(names...))
}

// selectedPackage returns the currently highlighted orphan, or nil.
func (m cleanupModel) selectedPackage() *pacman.Package {
	if len(m.orphans) == 0 || m.cursor >= len(m.orphans) {
		return nil
	}
	p := m.orphans[m.cursor]
	return &p
}

func (m cleanupModel) view(width, height int) string {
	var sb strings.Builder

	title := TitleStyle.Render("Orphan Cleanup")
	sb.WriteString(title + "\n")
	sb.WriteString(SubtitleStyle.Render(strings.Repeat("─", width-4)) + "\n\n")

	center := func(s string) string {
		return lipgloss.NewStyle().Width(width - 4).Align(lipgloss.Center).Render(s)
	}

	switch m.state {
	case cleanupStateIdle:
		sb.WriteString(center(InfoStyle.Render("Press r to scan for orphan packages")))

	case cleanupStateLoading:
		sb.WriteString(center(fmt.Sprintf("%s  Scanning for orphan packages…", m.spinner.View())))

	case cleanupStateError:
		sb.WriteString(ErrorStyle.Render("✗ Error: " + m.errMsg) + "\n")
		sb.WriteString(center(SubtitleStyle.Render("Press r to try again")))

	case cleanupStateDone:
		sb.WriteString(center(SuccessStyle.Render("✓ " + m.doneMsg)) + "\n\n")
		sb.WriteString(center(SubtitleStyle.Render("Press r to scan again")))

	case cleanupStateCleaning:
		sb.WriteString(center(fmt.Sprintf("%s  Removing packages…", m.spinner.View())))

	case cleanupStateReady:
		if len(m.orphans) == 0 {
			sb.WriteString(center(SuccessStyle.Render("✓ No orphan packages found!")) + "\n\n")
			sb.WriteString(center(SubtitleStyle.Render("Your system is clean.")))
		} else {
			count := WarningStyle.Render(fmt.Sprintf("%d orphan package(s) found", len(m.orphans)))
			sb.WriteString(count + "\n")
			hint := SubtitleStyle.Render("Use space to select, d to remove selected, or use the buttons below")
			sb.WriteString(hint + "\n\n")

			// Table header
			header := lipgloss.JoinHorizontal(lipgloss.Top,
				DetailKeyStyle.Render(fmt.Sprintf("  %-3s", "")),
				DetailKeyStyle.Render(fmt.Sprintf("%-30s", "Package")),
				DetailKeyStyle.Render("Version"),
			)
			sb.WriteString(header + "\n")
			sb.WriteString(SubtitleStyle.Render(strings.Repeat("─", width-4)) + "\n")

			// Show packages
			maxRows := height - 18
			if maxRows < 1 {
				maxRows = 1
			}
			for i, pkg := range m.orphans {
				if i >= maxRows {
					more := SubtitleStyle.Render(fmt.Sprintf("  … and %d more", len(m.orphans)-maxRows))
					sb.WriteString(more + "\n")
					break
				}
				sb.WriteString(m.renderOrphanRow(i, width, pkg))
			}

			sb.WriteString("\n")
			// Action buttons
			removeAllBtn := DialogBtnActive.
				Width(26).
				Align(lipgloss.Center).
				Render("🗑  Remove All  [enter]")

			removeSelBtn := ""
			if len(m.selected) > 0 {
				removeSelBtn = "  " + DialogBtnInactive.
					Width(26).
					Align(lipgloss.Center).
					Render(fmt.Sprintf("  Remove Selected (%d)  [d]", len(m.selected)))
			}

			sb.WriteString(center(removeAllBtn + removeSelBtn))
		}
	}

	return sb.String()
}

func (m cleanupModel) renderOrphanRow(i, width int, pkg pacman.Package) string {
	selected := i == m.cursor
	checked := m.selected[i]

	checkmark := "[ ]"
	if checked {
		checkmark = SuccessStyle.Render("[✓]")
	}

	cursor := "  "
	var rowStyle lipgloss.Style
	if selected {
		cursor = StatusKeyStyle.Render("▶ ")
		rowStyle = lipgloss.NewStyle().Background(DarkerGray).Width(width - 4)
	} else {
		rowStyle = lipgloss.NewStyle().Width(width - 4)
	}

	name := pkg.Name
	if len(name) > 28 {
		name = name[:25] + "…"
	}

	line := cursor + checkmark + "  " +
		PkgNameStyle.Render(fmt.Sprintf("%-28s", name)) + "  " +
		PkgVersionStyle.Render(pkg.Version)

	return rowStyle.Render(line) + "\n"
}

// hints returns key hints for the cleanup tab.
func (m cleanupModel) hints() string {
	h := []string{
		"↑↓", "navigate",
		"space", "select",
		"d", "remove selected",
		"enter", "remove all",
		"r", "refresh",
	}
	return renderHints(h)
}

// contextInfo returns a summary for the status bar.
func (m cleanupModel) contextInfo() string {
	switch m.state {
	case cleanupStateLoading:
		return "scanning…"
	case cleanupStateReady:
		if len(m.orphans) == 0 {
			return "clean — no orphans"
		}
		if len(m.selected) > 0 {
			return fmt.Sprintf("%d orphans  (%d selected)", len(m.orphans), len(m.selected))
		}
		return fmt.Sprintf("%d orphan(s) found", len(m.orphans))
	case cleanupStateCleaning:
		return "removing…"
	case cleanupStateDone:
		return "cleanup complete"
	default:
		return ""
	}
}
