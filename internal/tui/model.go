package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"pacman-tui/internal/pacman"
)

// tabIndex names the tabs by their numeric index.
const (
	tabSearch    = 0
	tabInstalled = 1
	tabUpdate    = 2
	tabCleanup   = 3
)

// Model is the root Bubble Tea model for the Pacman TUI.
type Model struct {
	keys      KeyMap
	help      help.Model
	showHelp  bool
	statusBar statusBarModel
	confirm   confirmModel
	detail    detailModel

	// Tab sub-models
	activeTab int
	search    searchModel
	installed installedModel
	update    updateModel
	cleanup   cleanupModel

	// Window dimensions
	width  int
	height int
}

// NewModel creates a fresh root model with default settings.
func NewModel() Model {
	keys := DefaultKeyMap()
	aurHelper := pacman.DetectAURHelper()

	return Model{
		keys:      keys,
		help:      help.New(),
		statusBar: newStatusBar(),
		confirm:   newConfirmModel(),
		detail:    newDetailModel(),
		activeTab: tabSearch,
		search:    newSearchModel(keys),
		installed: newInstalledModel(keys),
		update:    newUpdateModel(keys, aurHelper),
		cleanup:   newCleanupModel(keys),
	}
}

// Init implements tea.Model. Starts initial data loads.
func (m Model) Init() tea.Cmd {
	m.search.resetFocus()
	return m.installed.init()
}

// Update implements tea.Model — the central event dispatcher.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// ── Window resize ────────────────────────────────────────────────────
	if wsMsg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = wsMsg.Width
		m.height = wsMsg.Height
		m.help.Width = wsMsg.Width
		m.detail.setSize(wsMsg.Width, wsMsg.Height)
		m.confirm.width = wsMsg.Width
		m.confirm.height = wsMsg.Height
		return m, nil
	}

	// ── Confirmation dialog has priority ─────────────────────────────────
	if m.confirm.isActive() {
		var cmd tea.Cmd
		m.confirm, cmd = m.confirm.update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		// After confirm, refresh installed list if needed
		return m, tea.Batch(cmds...)
	}

	// ── Detail overlay has priority ──────────────────────────────────────
	if m.detail.isActive() {
		if infoMsg, ok := msg.(pacman.PackageInfoMsg); ok {
			if infoMsg.Err == nil && infoMsg.Pkg != nil {
				m.detail.setPackage(infoMsg.Pkg)
			} else if infoMsg.Err != nil {
				m.statusBar.setError(infoMsg.Err.Error())
				m.detail.hide()
			}
			return m, nil
		}
		var cmd tea.Cmd
		m.detail, cmd = m.detail.update(msg)
		return m, cmd
	}

	// ── Global key handling ──────────────────────────────────────────────
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(keyMsg, m.keys.Help):
			m.showHelp = !m.showHelp
			return m, nil

		case key.Matches(keyMsg, m.keys.NextTab):
			m.activeTab = (m.activeTab + 1) % len(tabs)
			m.statusBar.clear()
			m.onTabSwitch()
			return m, m.tabInitCmd()

		case key.Matches(keyMsg, m.keys.PrevTab):
			m.activeTab = (m.activeTab + len(tabs) - 1) % len(tabs)
			m.statusBar.clear()
			m.onTabSwitch()
			return m, m.tabInitCmd()

		// Info overlay — open for the selected package in any tab
		case key.Matches(keyMsg, m.keys.Info):
			pkg := m.activeSelectedPackage()
			if pkg != nil {
				m.detail.show(nil) // show loading spinner
				return m, pacman.GetInfo(pkg.Name)
			}

		// Install — from search tab; Enter on cleanup = remove all
		case key.Matches(keyMsg, m.keys.Install):
			if m.activeTab == tabSearch {
				pkg := m.search.selectedPackage()
				if pkg != nil && !pkg.IsInstalled {
					m.confirm.width = m.width
					m.confirm.height = m.height
					m.confirm.show(confirmAction{
						title:         "Install package?",
						message:       pkg.Name + "  " + pkg.Version,
						action:        pacman.Install(pkg.Name),
						isDestructive: false,
					})
					return m, nil
				}
			} else if m.activeTab == tabCleanup {
				if m.cleanup.state == cleanupStateReady && len(m.cleanup.orphans) > 0 {
					m.confirm.width = m.width
					m.confirm.height = m.height
					m.confirm.show(confirmAction{
						title:         "Remove all orphans?",
						message:       "This will remove all orphaned packages and their dependencies.",
						action:        m.cleanup.removeAllOrphans(),
						isDestructive: true,
					})
					return m, nil
				}
			}

		// Remove — from installed, search, or cleanup tab
		case key.Matches(keyMsg, m.keys.Remove):
			if m.activeTab == tabCleanup && len(m.cleanup.selected) > 0 {
				m.confirm.width = m.width
				m.confirm.height = m.height
				m.confirm.show(confirmAction{
					title:         fmt.Sprintf("Remove %d selected package(s)?", len(m.cleanup.selected)),
					message:       "This will remove the selected orphaned packages.",
					action:        m.cleanup.removeSelected(),
					isDestructive: true,
				})
				return m, nil
			}
			pkg := m.activeSelectedPackage()
			if pkg != nil && pkg.IsInstalled {
				m.confirm.width = m.width
				m.confirm.height = m.height
				m.confirm.show(confirmAction{
					title:         "Remove package?",
					message:       pkg.Name + "  " + pkg.Version,
					action:        pacman.Remove(pkg.Name),
					isDestructive: true,
				})
				return m, nil
			}

		// Refresh current tab
		case key.Matches(keyMsg, m.keys.Refresh):
			return m, m.refreshActiveTab()
		}
	}

	// ── OperationFinishedMsg — global result handler ──────────────────────
	if opMsg, ok := msg.(pacman.OperationFinishedMsg); ok {
		if opMsg.Err != nil {
			m.statusBar.setError("Error: " + opMsg.Err.Error())
		} else {
			switch opMsg.Operation {
			case "install":
				m.statusBar.setSuccess("Package installed successfully!")
				cmds = append(cmds, m.installed.refresh())
			case "remove":
				m.statusBar.setSuccess("Package removed successfully!")
				cmds = append(cmds, m.installed.refresh())
			case "remove-orphans":
				m.statusBar.setSuccess("Orphans removed!")
				cmds = append(cmds, m.cleanup.init())
			case "sysupdate":
				m.statusBar.setSuccess("System updated successfully!")
			case "aur-install":
				m.statusBar.setSuccess("AUR package installed!")
				cmds = append(cmds, m.installed.refresh())
			case "aur-update":
				m.statusBar.setSuccess("AUR packages updated!")
			}
		}
		// Route to active tab too
		var cmd tea.Cmd
		m, cmd = m.routeToActiveTab(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	// ── PackageInfoMsg (when detail not open) ─────────────────────────────
	if infoMsg, ok := msg.(pacman.PackageInfoMsg); ok {
		if infoMsg.Err != nil {
			m.statusBar.setError("Could not load package info: " + infoMsg.Err.Error())
		} else if infoMsg.Pkg != nil {
			m.detail.show(infoMsg.Pkg)
		}
		return m, nil
	}

	// ── Route messages to active tab ─────────────────────────────────────
	var cmd tea.Cmd
	m, cmd = m.routeToActiveTab(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// routeToActiveTab forwards the message to the currently active tab.
func (m Model) routeToActiveTab(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.activeTab {
	case tabSearch:
		m.search, cmd = m.search.update(msg)
	case tabInstalled:
		m.installed, cmd = m.installed.update(msg)
	case tabUpdate:
		m.update, cmd = m.update.update(msg)
	case tabCleanup:
		m.cleanup, cmd = m.cleanup.update(msg)
	}
	return m, cmd
}

// View implements tea.Model — assembles the full screen.
func (m Model) View() string {
	if m.width == 0 {
		return "Initialising…"
	}

	// Overlays take the entire screen
	if m.confirm.isActive() {
		return m.confirm.view()
	}
	if m.detail.isActive() {
		return m.detail.view()
	}

	// ── Header & tab bar ─────────────────────────────────────────────────
	header := renderHeader()
	tabBar := renderTabBar(m.activeTab, m.width)

	// ── Help bar at bottom ───────────────────────────────────────────────
	m.help.ShowAll = m.showHelp
	helpView := HelpStyle.Render(m.help.View(m.keys))

	// ── Status bar ───────────────────────────────────────────────────────
	m.statusBar.contextInfo = m.activeContextInfo()
	statusBar := m.statusBar.render(m.width)

	// ── Content area ─────────────────────────────────────────────────────
	headerH := lipgloss.Height(header)
	tabBarH := lipgloss.Height(tabBar)
	statusH := lipgloss.Height(statusBar)
	helpH := lipgloss.Height(helpView)
	contentHeight := m.height - headerH - tabBarH - statusH - helpH - 2 // 2 for panel borders
	if contentHeight < 1 {
		contentHeight = 1
	}

	var contentStr string
	switch m.activeTab {
	case tabSearch:
		contentStr = m.search.view(m.width, contentHeight)
	case tabInstalled:
		contentStr = m.installed.view(m.width, contentHeight)
	case tabUpdate:
		contentStr = m.update.view(m.width, contentHeight)
	case tabCleanup:
		contentStr = m.cleanup.view(m.width, contentHeight)
	}

	// Ensure content doesn't overflow
	contentLines := strings.Split(contentStr, "\n")
	if len(contentLines) > contentHeight {
		contentLines = contentLines[:contentHeight]
		contentStr = strings.Join(contentLines, "\n")
	}

	panel := ContentStyle.
		Width(m.width - 4).
		Height(contentHeight).
		Render(contentStr)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		tabBar,
		panel,
		statusBar,
		helpView,
	)
}

// ── Helpers ──────────────────────────────────────────────────────────────────

// activeSelectedPackage returns the selected package from the current tab.
func (m Model) activeSelectedPackage() *pacman.Package {
	switch m.activeTab {
	case tabSearch:
		return m.search.selectedPackage()
	case tabInstalled:
		return m.installed.selectedPackage()
	case tabCleanup:
		return m.cleanup.selectedPackage()
	}
	return nil
}

// activeContextInfo returns the status bar context text for the active tab.
func (m Model) activeContextInfo() string {
	switch m.activeTab {
	case tabSearch:
		return m.search.contextInfo()
	case tabInstalled:
		return m.installed.contextInfo()
	case tabUpdate:
		return m.update.contextInfo()
	case tabCleanup:
		return m.cleanup.contextInfo()
	}
	return ""
}

// onTabSwitch handles state setup when switching tabs.
func (m *Model) onTabSwitch() {
	if m.activeTab == tabSearch {
		m.search.resetFocus()
	}
}

// tabInitCmd returns the init command for a freshly-switched-to tab.
func (m *Model) tabInitCmd() tea.Cmd {
	switch m.activeTab {
	case tabUpdate:
		if m.update.state == updateStateIdle {
			return m.update.init()
		}
	case tabCleanup:
		if m.cleanup.state == cleanupStateIdle {
			return m.cleanup.init()
		}
	}
	return nil
}

// refreshActiveTab triggers a data refresh on the current tab.
func (m *Model) refreshActiveTab() tea.Cmd {
	switch m.activeTab {
	case tabInstalled:
		return m.installed.refresh()
	case tabUpdate:
		return m.update.init()
	case tabCleanup:
		return m.cleanup.init()
	}
	return nil
}
