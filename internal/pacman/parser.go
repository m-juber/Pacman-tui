package pacman

import (
	"strings"
)

// ParseSearchOutput parses the output of `pacman -Ss`.
// The format alternates between header and description lines:
//
//	repo/name version (group1 group2)
//	    Description text here
func ParseSearchOutput(output string) []Package {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil
	}

	lines := strings.Split(output, "\n")
	var packages []Package

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Header lines start without whitespace: repo/name version [installed] (groups)
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			// This is a description line without a preceding header; skip it
			continue
		}

		pkg := Package{}

		// Parse header: repo/name version [installed] (group1 group2)
		headerLine := strings.TrimSpace(line)

		// Extract groups if present (in parentheses at the end)
		if idx := strings.LastIndex(headerLine, "("); idx != -1 {
			if endIdx := strings.LastIndex(headerLine, ")"); endIdx > idx {
				groupStr := headerLine[idx+1 : endIdx]
				groups := strings.Fields(groupStr)
				if len(groups) > 0 {
					pkg.Groups = groups
				}
				headerLine = strings.TrimSpace(headerLine[:idx])
			}
		}

		// Check for [installed] marker
		if strings.Contains(headerLine, "[installed]") {
			pkg.IsInstalled = true
			headerLine = strings.Replace(headerLine, "[installed]", "", 1)
			headerLine = strings.TrimSpace(headerLine)
		}

		// Split into repo/name and version
		parts := strings.Fields(headerLine)
		if len(parts) < 2 {
			continue
		}

		repoName := parts[0]
		pkg.Version = parts[1]

		// Split repo/name
		slashIdx := strings.Index(repoName, "/")
		if slashIdx != -1 {
			pkg.Repository = repoName[:slashIdx]
			pkg.Name = repoName[slashIdx+1:]
		} else {
			pkg.Name = repoName
		}

		// Next line(s) should be the description (indented with spaces)
		var descParts []string
		for i+1 < len(lines) {
			nextLine := lines[i+1]
			if strings.HasPrefix(nextLine, " ") || strings.HasPrefix(nextLine, "\t") {
				descParts = append(descParts, strings.TrimSpace(nextLine))
				i++
			} else {
				break
			}
		}
		pkg.Desc = strings.Join(descParts, " ")

		packages = append(packages, pkg)
	}

	return packages
}

// ParseInstalledOutput parses the output of `pacman -Q` or `pacman -Qtdq`.
// For `pacman -Q`, each line is: name version
// For `pacman -Qtdq`, each line is just: name
func ParseInstalledOutput(output string) []Package {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil
	}

	lines := strings.Split(output, "\n")
	var packages []Package

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		pkg := Package{IsInstalled: true}
		parts := strings.Fields(line)

		pkg.Name = parts[0]
		if len(parts) >= 2 {
			pkg.Version = parts[1]
		}

		packages = append(packages, pkg)
	}

	return packages
}

// ParseInfoOutput parses the output of `pacman -Qi` or `pacman -Si`.
// The format is key-value pairs separated by " : " with potential multi-line values.
//
//	Name            : package-name
//	Version         : 1.2.3-1
//	Description     : Some description
func ParseInfoOutput(output string) *Package {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil
	}

	pkg := &Package{}
	lines := strings.Split(output, "\n")

	var currentKey string
	var currentValue string

	flushField := func() {
		if currentKey == "" {
			return
		}
		value := strings.TrimSpace(currentValue)
		applyInfoField(pkg, currentKey, value)
	}

	for _, line := range lines {
		// Check if this line has a key-value separator
		sepIdx := strings.Index(line, " : ")
		if sepIdx != -1 {
			// Flush previous field
			flushField()

			currentKey = strings.TrimSpace(line[:sepIdx])
			currentValue = strings.TrimSpace(line[sepIdx+3:])
		} else {
			// Continuation line for multi-line values
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && currentKey != "" {
				currentValue += " " + trimmed
			}
		}
	}

	// Flush the last field
	flushField()

	if pkg.Name == "" {
		return nil
	}

	return pkg
}

// applyInfoField maps a parsed key-value pair to the appropriate Package field.
func applyInfoField(pkg *Package, key, value string) {
	if value == "None" {
		value = ""
	}

	switch key {
	case "Name":
		pkg.Name = value
	case "Version":
		pkg.Version = value
	case "Description":
		pkg.Desc = value
	case "Architecture":
		pkg.Architecture = value
	case "URL":
		pkg.URL = value
	case "Repository":
		pkg.Repository = value
	case "Licenses":
		if value != "" {
			pkg.Licenses = splitTrimmed(value, "  ")
		}
	case "Groups":
		if value != "" {
			pkg.Groups = splitTrimmed(value, "  ")
		}
	case "Depends On":
		if value != "" {
			pkg.DependsOn = splitTrimmed(value, "  ")
		}
	case "Optional Deps":
		if value != "" {
			pkg.OptionalDeps = splitTrimmed(value, "  ")
		}
	case "Required By":
		if value != "" {
			pkg.RequiredBy = splitTrimmed(value, "  ")
		}
	case "Conflicts With":
		if value != "" {
			pkg.ConflictsWith = splitTrimmed(value, "  ")
		}
	case "Installed Size":
		pkg.InstalledSize = value
	case "Download Size":
		pkg.DownloadSize = value
	case "Packager":
		pkg.Packager = value
	case "Build Date":
		pkg.BuildDate = value
	case "Install Date":
		pkg.InstallDate = value
		if value != "" {
			pkg.IsInstalled = true
		}
	case "Install Reason":
		pkg.InstallReason = value
		if strings.Contains(value, "Explicitly") {
			pkg.IsExplicit = true
		}
	case "Validated By":
		pkg.Validated = value
	}
}

// splitTrimmed splits a string by a separator and trims whitespace from each element.
// It handles both single-space and double-space separated lists that pacman uses.
func splitTrimmed(s, sep string) []string {
	parts := strings.Split(s, sep)
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// ParseCheckUpdates parses the output of the `checkupdates` command.
// Each line has the format: name old_version -> new_version
func ParseCheckUpdates(output string) []Package {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil
	}

	lines := strings.Split(output, "\n")
	var packages []Package

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 4 {
			// Minimum: name old_version -> new_version
			continue
		}

		pkg := Package{
			Name:        parts[0],
			Version:     parts[len(parts)-1], // new version is the last field
			IsInstalled: true,
		}

		packages = append(packages, pkg)
	}

	return packages
}
