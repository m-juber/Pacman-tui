package pacman

import "fmt"

// Package represents a pacman package with all metadata fields.
type Package struct {
	Name          string
	Version       string
	Desc          string // package description (named Desc to avoid clash with Description() method)
	Repository    string // core, extra, community, multilib, aur
	Architecture  string
	URL           string
	Licenses      []string
	Groups        []string
	DependsOn     []string
	OptionalDeps  []string
	RequiredBy    []string
	ConflictsWith []string
	InstalledSize string
	DownloadSize  string
	Packager      string
	BuildDate     string
	InstallDate   string
	InstallReason string
	Validated     string
	IsInstalled   bool
	IsExplicit    bool
}

// FilterValue implements list.Item for filtering in the bubbles list component.
func (p Package) FilterValue() string {
	return p.Name
}

// Title implements list.Item for display in the bubbles list component.
func (p Package) Title() string {
	return p.Name
}

// Description implements list.Item for display in the bubbles list component.
// Returns a formatted string with version, repository, and description.
func (p Package) Description() string {
	return fmt.Sprintf("%s  %s  %s", p.Version, p.Repository, p.Desc)
}
