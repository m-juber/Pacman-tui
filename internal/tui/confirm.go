package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// confirmAction describes what will happen if the user confirms.
type confirmAction struct {
	title   string // e.g. "Remove package?"
	message string // e.g. "vim 9.1.0"
	action  tea.Cmd
	isDestructive bool
}

// confirmModel is the confirmation dialog overlay.
type confirmModel struct {
	active  bool
	action  confirmAction
	focused bool // true = Yes is focused, false = No is focused
	width   int
	height  int
}

func newConfirmModel() confirmModel {
	return confirmModel{}
}

// show activates the dialog with the given action details.
func (c *confirmModel) show(a confirmAction) {
	c.active = true
	c.action = a
	c.focused = false // Default: No (safe default for destructive actions)
}

// hide deactivates the dialog.
func (c *confirmModel) hide() {
	c.active = false
}

// isActive returns whether the dialog is currently visible.
func (c confirmModel) isActive() bool {
	return c.active
}

// update handles keyboard input for the dialog.
// Returns the tea.Cmd to execute if confirmed, or nil if cancelled.
func (c confirmModel) update(msg tea.Msg) (confirmModel, tea.Cmd) {
	if !c.active {
		return c, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("left", "h", "shift+tab"))):
			c.focused = true // Move to Yes
		case key.Matches(msg, key.NewBinding(key.WithKeys("right", "l", "tab"))):
			c.focused = false // Move to No
		case key.Matches(msg, key.NewBinding(key.WithKeys("y"))):
			// Confirm immediately with Y
			cmd := c.action.action
			c.active = false
			return c, cmd
		case key.Matches(msg, key.NewBinding(key.WithKeys("n", "esc"))):
			// Cancel
			c.active = false
			return c, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if c.focused {
				// Yes is focused
				cmd := c.action.action
				c.active = false
				return c, cmd
			}
			// No is focused
			c.active = false
			return c, nil
		}
	}

	return c, nil
}

// view renders the confirmation dialog.
func (c confirmModel) view() string {
	if !c.active {
		return ""
	}

	var titleStyle lipgloss.Style
	if c.action.isDestructive {
		titleStyle = ErrorStyle.Copy().Bold(true)
	} else {
		titleStyle = DialogTitleStyle
	}

	title := titleStyle.Render(c.action.title)
	msg := InfoStyle.Render(c.action.message)

	hint := StatusTextStyle.Render("\n\nPress ") +
		StatusKeyStyle.Render("y") +
		StatusTextStyle.Render(" to confirm, ") +
		StatusKeyStyle.Render("n/esc") +
		StatusTextStyle.Render(" to cancel\n")

	var yesBtn, noBtn string
	if c.focused {
		yesBtn = DialogBtnActive.Render("  Yes  ")
		noBtn = DialogBtnInactive.Render("  No  ")
	} else {
		yesBtn = DialogBtnInactive.Render("  Yes  ")
		noBtn = DialogBtnActive.Render("  No  ")
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, yesBtn, "  ", noBtn)

	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		msg,
		hint,
		buttons,
	)

	dialog := DialogStyle.Render(content)

	// Center the dialog on screen
	return lipgloss.Place(c.width, c.height,
		lipgloss.Center, lipgloss.Center,
		dialog,
	)
}
