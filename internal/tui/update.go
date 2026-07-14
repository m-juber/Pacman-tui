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

// updateTabState is the current state of the update tab.
type updateTabState int

const (
	updateStateIdle updateTabState = iota
	updateStateChecking
	updateStateReady
	updateStateUpdating
	updateStateDone
	updateStateError
)

// updateModel is the System Update tab model.
type updateModel struct {
	keys     KeyMap
	spinner  spinner.Model
	state    updateTabState
	updates  []pacman.Package
	errMsg   string
	doneMsg  string
	aurHelper string
}

func newUpdateModel(keys KeyMap, aurHelper string) updateModel {
	sp := spinner.New()
	sp.Spinner = spinner.Points
	sp.Style = SpinnerStyle

	return updateModel{
		keys:      keys,
		spinner:   sp,
		state:     updateStateIdle,
		aurHelper: aurHelper,
	}
}

// init checks for available updates when the tab becomes active.
func (m updateModel) init() tea.Cmd {
	m.state = updateStateChecking
	return tea.Batch(m.spinner.Tick, pacman.CheckUpdates())
}

func (m updateModel) update(msg tea.Msg) (updateModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Update):
			if m.state == updateStateReady {
				m.state = updateStateUpdating
				return m, tea.Batch(m.spinner.Tick, pacman.SystemUpdate())
			}
		case key.Matches(msg, m.keys.Refresh):
			m.state = updateStateChecking
			m.updates = nil
			m.errMsg = ""
			return m, tea.Batch(m.spinner.Tick, pacman.CheckUpdates())
		}

	case pacman.UpdatesAvailableMsg:
		if msg.Err != nil {
			m.state = updateStateError
			m.errMsg = msg.Err.Error()
		} else {
			m.state = updateStateReady
			m.updates = msg.Packages
		}
		return m, nil

	case pacman.OperationFinishedMsg:
		if msg.Operation == "sysupdate" || msg.Operation == "aur-update" {
			if msg.Err != nil {
				m.state = updateStateError
				m.errMsg = msg.Err.Error()
			} else {
				m.state = updateStateDone
				m.doneMsg = "System updated successfully!"
			}
		}
		return m, nil

	case spinner.TickMsg:
		if m.state == updateStateChecking || m.state == updateStateUpdating {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m updateModel) view(width, height int) string {
	var sb strings.Builder

	// Title
	title := TitleStyle.Render("System Update")
	sb.WriteString(title + "\n")
	sb.WriteString(SubtitleStyle.Render(strings.Repeat("─", width-4)) + "\n\n")

	center := func(s string) string {
		return lipgloss.NewStyle().Width(width - 4).Align(lipgloss.Center).Render(s)
	}

	switch m.state {
	case updateStateIdle:
		sb.WriteString(center(InfoStyle.Render("Press r to check for updates")))

	case updateStateChecking:
		sb.WriteString(center(fmt.Sprintf("%s  Checking for updates…", m.spinner.View())))

	case updateStateError:
		sb.WriteString(ErrorStyle.Render("✗ Error: " + m.errMsg) + "\n")
		sb.WriteString(center(SubtitleStyle.Render("Press r to try again")))

	case updateStateDone:
		sb.WriteString(center(SuccessStyle.Render("✓ " + m.doneMsg)) + "\n\n")
		sb.WriteString(center(SubtitleStyle.Render("Press r to check for updates again")))

	case updateStateUpdating:
		sb.WriteString(center(fmt.Sprintf("%s  Updating system…\n\n", m.spinner.View())))
		sb.WriteString(center(WarningStyle.Render("The terminal will be taken over by pacman.\nDo not close the window.")))

	case updateStateReady:
		if len(m.updates) == 0 {
			sb.WriteString(center(SuccessStyle.Render("✓ Your system is up to date!")) + "\n\n")
			sb.WriteString(center(SubtitleStyle.Render("Press r to check again")))
		} else {
			count := InfoStyle.Render(fmt.Sprintf("%d update(s) available", len(m.updates)))
			sb.WriteString(count + "\n\n")

			// Table header
			header := lipgloss.JoinHorizontal(lipgloss.Top,
				DetailKeyStyle.Render(fmt.Sprintf("%-30s", "Package")),
				DetailKeyStyle.Render("New Version"),
			)
			sb.WriteString(header + "\n")
			sb.WriteString(SubtitleStyle.Render(strings.Repeat("─", width-4)) + "\n")

			// Show packages (up to available height)
			maxRows := height - 16
			if maxRows < 1 {
				maxRows = 1
			}
			for i, pkg := range m.updates {
				if i >= maxRows {
					more := SubtitleStyle.Render(fmt.Sprintf("  … and %d more", len(m.updates)-maxRows))
					sb.WriteString(more + "\n")
					break
				}
				name := pkg.Name
				if len(name) > 28 {
					name = name[:25] + "…"
				}
				sb.WriteString(
					PkgNameStyle.Render(fmt.Sprintf("  %-28s", name)) + "  " +
						SuccessStyle.Render(pkg.Version) + "\n",
				)
			}

			sb.WriteString("\n")
			// Update button
			btn := DialogBtnActive.
				Width(24).
				Align(lipgloss.Center).
				Render("⬆  Update System  [u]")
			sb.WriteString(center(btn))

			if m.aurHelper != "" {
				sb.WriteString("\n\n")
				aurBtn := DialogBtnInactive.
					Width(24).
					Align(lipgloss.Center).
					Render(fmt.Sprintf("  Update AUR (%s)", m.aurHelper))
				sb.WriteString(center(aurBtn))
			}
		}
	}

	return sb.String()
}

// hints returns key hints for the update tab.
func (m updateModel) hints() string {
	h := []string{"u", "update system", "r", "refresh"}
	if m.aurHelper != "" {
		h = append(h, "a", "update AUR")
	}
	return renderHints(h)
}

// contextInfo returns a summary for the status bar.
func (m updateModel) contextInfo() string {
	switch m.state {
	case updateStateChecking:
		return "checking for updates…"
	case updateStateReady:
		if len(m.updates) == 0 {
			return "system up to date"
		}
		return fmt.Sprintf("%d update(s) available", len(m.updates))
	case updateStateUpdating:
		return "updating…"
	case updateStateDone:
		return "update complete"
	default:
		return ""
	}
}
