package pacman

import (
	"bytes"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

// Message types returned by pacman commands.

// SearchResultMsg is returned by Search with parsed package results.
type SearchResultMsg struct {
	Packages []Package
	Err      error
}

// InstalledListMsg is returned by ListInstalled with parsed package results.
type InstalledListMsg struct {
	Packages []Package
	Err      error
}

// PackageInfoMsg is returned by GetInfo with detailed package information.
type PackageInfoMsg struct {
	Pkg *Package
	Err error
}

// OrphanListMsg is returned by ListOrphans with orphaned packages.
type OrphanListMsg struct {
	Packages []Package
	Err      error
}

// UpdatesAvailableMsg is returned by CheckUpdates with available updates.
type UpdatesAvailableMsg struct {
	Packages []Package
	Err      error
}

// OperationFinishedMsg is returned by interactive operations (install, remove, update).
type OperationFinishedMsg struct {
	Operation string
	Err       error
}

// Search runs `pacman -Ss query` and returns parsed results as a SearchResultMsg.
func Search(query string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("pacman", "-Ss", query)
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				// Exit code 1 means no results found
				return SearchResultMsg{Packages: nil}
			}
			return SearchResultMsg{Err: err}
		}
		return SearchResultMsg{Packages: ParseSearchOutput(out.String())}
	}
}

// ListInstalled runs `pacman -Q` and returns all installed packages.
func ListInstalled() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("pacman", "-Q")
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			return InstalledListMsg{Err: err}
		}
		return InstalledListMsg{Packages: ParseInstalledOutput(out.String())}
	}
}

// GetInfo retrieves detailed info for a package. Tries `pacman -Qi` (installed) first,
// then falls back to `pacman -Si` (sync database).
func GetInfo(name string) tea.Cmd {
	return func() tea.Msg {
		// Try local info first
		cmd := exec.Command("pacman", "-Qi", name)
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err == nil {
			pkg := ParseInfoOutput(out.String())
			if pkg != nil {
				pkg.IsInstalled = true
				return PackageInfoMsg{Pkg: pkg}
			}
		}

		// Fall back to sync database info
		out.Reset()
		cmd = exec.Command("pacman", "-Si", name)
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			return PackageInfoMsg{Err: err}
		}
		pkg := ParseInfoOutput(out.String())
		if pkg == nil {
			return PackageInfoMsg{Err: nil}
		}
		return PackageInfoMsg{Pkg: pkg}
	}
}

// ListOrphans runs `pacman -Qtdq` to find orphaned packages.
func ListOrphans() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("pacman", "-Qtdq")
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				// Exit code 1 means no orphans found
				return OrphanListMsg{Packages: nil}
			}
			return OrphanListMsg{Err: err}
		}
		packages := ParseInstalledOutput(out.String())
		return OrphanListMsg{Packages: packages}
	}
}

// CheckUpdates runs `checkupdates` to find available package updates.
func CheckUpdates() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("checkupdates")
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 2 {
				// Exit code 2 means no updates available
				return UpdatesAvailableMsg{Packages: nil}
			}
			return UpdatesAvailableMsg{Err: err}
		}
		return UpdatesAvailableMsg{Packages: ParseCheckUpdates(out.String())}
	}
}

// Install uses tea.ExecProcess to run `sudo pacman -S --noconfirm` for the given packages.
// This hands control to the subprocess for interactive terminal use.
func Install(names ...string) tea.Cmd {
	args := append([]string{"pacman", "-S", "--noconfirm"}, names...)
	c := exec.Command("sudo", args...)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return OperationFinishedMsg{Operation: "install", Err: err}
	})
}

// Remove uses tea.ExecProcess to run `sudo pacman -Rs --noconfirm` for the given packages.
func Remove(names ...string) tea.Cmd {
	args := append([]string{"pacman", "-Rs", "--noconfirm"}, names...)
	c := exec.Command("sudo", args...)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return OperationFinishedMsg{Operation: "remove", Err: err}
	})
}

// SystemUpdate uses tea.ExecProcess to run `sudo pacman -Syu --noconfirm`.
func SystemUpdate() tea.Cmd {
	c := exec.Command("sudo", "pacman", "-Syu", "--noconfirm")
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return OperationFinishedMsg{Operation: "sysupdate", Err: err}
	})
}

// RemoveOrphans uses tea.ExecProcess to run `sudo pacman -Rns $(pacman -Qtdq)`.
// It first collects orphan names, then removes them in one operation.
func RemoveOrphans() tea.Cmd {
	c := exec.Command("bash", "-c", "sudo pacman -Rns --noconfirm $(pacman -Qtdq)")
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return OperationFinishedMsg{Operation: "remove-orphans", Err: err}
	})
}
