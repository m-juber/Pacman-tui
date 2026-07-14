package pacman

import (
	"bytes"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

// aurHelper holds the name of the detected AUR helper binary.
var aurHelper string

// DetectAURHelper checks for known AUR helpers (paru, yay) on the system PATH
// and caches the first one found. Returns the name of the detected helper,
// or an empty string if none is found.
func DetectAURHelper() string {
	helpers := []string{"paru", "yay"}
	for _, h := range helpers {
		if _, err := exec.LookPath(h); err == nil {
			aurHelper = h
			return aurHelper
		}
	}
	aurHelper = ""
	return aurHelper
}

// HasAURHelper returns true if an AUR helper has been detected.
func HasAURHelper() bool {
	return aurHelper != ""
}

// GetAURHelperName returns the name of the detected AUR helper,
// or an empty string if none was detected.
func GetAURHelperName() string {
	return aurHelper
}

// AURSearch runs the detected AUR helper with `-Ss` to search the AUR.
// Returns a SearchResultMsg with parsed results. If no AUR helper is available,
// returns an error message.
func AURSearch(query string) tea.Cmd {
	return func() tea.Msg {
		if aurHelper == "" {
			return SearchResultMsg{Err: ErrNoAURHelper}
		}

		cmd := exec.Command(aurHelper, "-Ss", query)
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				// No results found
				return SearchResultMsg{Packages: nil}
			}
			return SearchResultMsg{Err: err}
		}
		return SearchResultMsg{Packages: ParseSearchOutput(out.String())}
	}
}

// AURInstall uses tea.ExecProcess with the detected AUR helper to install packages.
// The AUR helper is invoked with `-S --noconfirm`.
func AURInstall(names ...string) tea.Cmd {
	if aurHelper == "" {
		return func() tea.Msg {
			return OperationFinishedMsg{Operation: "aur-install", Err: ErrNoAURHelper}
		}
	}

	args := append([]string{"-S", "--noconfirm"}, names...)
	c := exec.Command(aurHelper, args...)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return OperationFinishedMsg{Operation: "aur-install", Err: err}
	})
}

// AURUpdate uses tea.ExecProcess with the detected AUR helper to update AUR packages.
// The AUR helper is invoked with `-Sua --noconfirm`.
func AURUpdate() tea.Cmd {
	if aurHelper == "" {
		return func() tea.Msg {
			return OperationFinishedMsg{Operation: "aur-update", Err: ErrNoAURHelper}
		}
	}

	c := exec.Command(aurHelper, "-Sua", "--noconfirm")
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return OperationFinishedMsg{Operation: "aur-update", Err: err}
	})
}

// ErrNoAURHelper is returned when an AUR operation is attempted without a detected helper.
var ErrNoAURHelper = &noAURHelperError{}

type noAURHelperError struct{}

func (e *noAURHelperError) Error() string {
	return "no AUR helper found (install paru or yay)"
}
