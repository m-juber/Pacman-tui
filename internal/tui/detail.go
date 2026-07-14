package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"pacman-tui/internal/pacman"
)

// detailModel shows full package information in a scrollable overlay.
type detailModel struct {
	active   bool
	pkg      *pacman.Package
	viewport viewport.Model
	width    int
	height   int
	loading  bool
}

func newDetailModel() detailModel {
	vp := viewport.New(80, 20)
	vp.KeyMap.Up = key.NewBinding(key.WithKeys("up", "k"))
	vp.KeyMap.Down = key.NewBinding(key.WithKeys("down", "j"))
	return detailModel{viewport: vp}
}

// show opens the detail overlay for the given package.
func (d *detailModel) show(pkg *pacman.Package) {
	d.active = true
	d.pkg = pkg
	d.loading = pkg == nil
	if pkg != nil {
		d.viewport.SetContent(formatPackageDetail(pkg))
		d.viewport.GotoTop()
	}
}

// hide closes the detail overlay.
func (d *detailModel) hide() {
	d.active = false
	d.pkg = nil
}

// isActive returns whether the detail overlay is visible.
func (d detailModel) isActive() bool {
	return d.active
}

// setSize resizes the detail viewport.
func (d *detailModel) setSize(w, h int) {
	d.width = w
	d.height = h
	// Leave room for borders and chrome
	d.viewport.Width = w - 8
	d.viewport.Height = h - 10
}

// setPackage updates the displayed package (called when async info loads).
func (d *detailModel) setPackage(pkg *pacman.Package) {
	d.pkg = pkg
	d.loading = false
	if pkg != nil {
		d.viewport.SetContent(formatPackageDetail(pkg))
		d.viewport.GotoTop()
	}
}

// update handles keyboard input when the detail overlay is active.
func (d detailModel) update(msg tea.Msg) (detailModel, tea.Cmd) {
	if !d.active {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "q", "i"))):
			d.hide()
			return d, nil
		}
	}

	var cmd tea.Cmd
	d.viewport, cmd = d.viewport.Update(msg)
	return d, cmd
}

// view renders the detail overlay on top of the screen.
func (d detailModel) view() string {
	if !d.active {
		return ""
	}

	var content string

	if d.loading || d.pkg == nil {
		content = lipgloss.JoinVertical(lipgloss.Center,
			SpinnerStyle.Render("⣾"),
			InfoStyle.Render(" Loading package info…"),
		)
	} else {
		pkgName := DetailTitleStyle.Render(d.pkg.Name)
		version := PkgVersionStyle.Render("  " + d.pkg.Version)

		status := ""
		if d.pkg.IsInstalled {
			status = "  " + InstalledBadge.Render("✓ installed")
		}

		header := pkgName + version + status

		scrollPercent := fmt.Sprintf("%3.0f%%", d.viewport.ScrollPercent()*100)
		scrollBar := StatusTextStyle.Render(scrollPercent)

		footer := lipgloss.NewStyle().
			Foreground(DarkGray).
			Render("↑↓ scroll  esc close") + "  " + scrollBar

		content = lipgloss.JoinVertical(lipgloss.Left,
			header,
			SubtitleStyle.Render(strings.Repeat("─", d.viewport.Width)),
			d.viewport.View(),
			SubtitleStyle.Render(strings.Repeat("─", d.viewport.Width)),
			footer,
		)
	}

	panel := DetailStyle.
		Width(d.width - 6).
		Height(d.height - 4).
		Render(content)

	return lipgloss.Place(d.width, d.height,
		lipgloss.Center, lipgloss.Center,
		panel,
	)
}

// formatPackageDetail builds the richly formatted string for the viewport.
func formatPackageDetail(pkg *pacman.Package) string {
	var sb strings.Builder

	field := func(label, value string) {
		if value == "" || value == "None" {
			return
		}
		sb.WriteString(DetailKeyStyle.Render(fmt.Sprintf("%-18s", label+":")))
		sb.WriteString("  ")
		sb.WriteString(DetailValStyle.Render(value))
		sb.WriteString("\n")
	}

	listField := func(label string, values []string) {
		if len(values) == 0 {
			return
		}
		sb.WriteString(DetailKeyStyle.Render(fmt.Sprintf("%-18s", label+":")))
		sb.WriteString("  ")
		sb.WriteString(DetailValStyle.Render(strings.Join(values, ", ")))
		sb.WriteString("\n")
	}

	// Description at the top
	if pkg.Desc != "" {
		sb.WriteString(PkgDescStyle.Render(pkg.Desc))
		sb.WriteString("\n\n")
	}

	sb.WriteString(SubtitleStyle.Render("── Package Information ──────────────────────────") + "\n\n")

	field("Repository", pkg.Repository)
	field("Version", pkg.Version)
	field("Architecture", pkg.Architecture)
	field("URL", pkg.URL)
	listField("Licenses", pkg.Licenses)
	listField("Groups", pkg.Groups)

	if pkg.InstalledSize != "" {
		sb.WriteString("\n")
		sb.WriteString(SubtitleStyle.Render("── Size & Dates ─────────────────────────────────") + "\n\n")
		field("Installed Size", pkg.InstalledSize)
		field("Download Size", pkg.DownloadSize)
		field("Build Date", pkg.BuildDate)
		field("Install Date", pkg.InstallDate)
		field("Install Reason", pkg.InstallReason)
	}

	if len(pkg.DependsOn) > 0 || len(pkg.OptionalDeps) > 0 {
		sb.WriteString("\n")
		sb.WriteString(SubtitleStyle.Render("── Dependencies ─────────────────────────────────") + "\n\n")
		listField("Depends On", pkg.DependsOn)
		listField("Optional Deps", pkg.OptionalDeps)
		listField("Required By", pkg.RequiredBy)
		listField("Conflicts With", pkg.ConflictsWith)
	}

	if pkg.Packager != "" {
		sb.WriteString("\n")
		sb.WriteString(SubtitleStyle.Render("── Maintainer ───────────────────────────────────") + "\n\n")
		field("Packager", pkg.Packager)
		field("Validated By", pkg.Validated)
	}

	return sb.String()
}
