# Pacman TUI

A beautiful, simple, and intuitive Terminal User Interface (TUI) wrapper for the Arch Linux `pacman` package manager, built with Go and Bubble Tea.

## Features

- **Search & Install**: Live, debounced searching of packages from official repositories and AUR.
- **Installed Packages**: Browse and filter your installed packages.
- **System Update**: Check for and apply system updates safely.
- **Orphan Cleanup**: Multi-select and remove orphaned packages easily.
- **AUR Support**: Automatically detects `yay` or `paru` if installed and uses them for AUR operations.
- **Beautiful UI**: Built with a sleek dark theme featuring purple and cyan accents, using the Bubble Tea framework.

## Installation

### Prerequisites

- Arch Linux (or an Arch-based distribution)
- `pacman`
- Go 1.20+ (if building from source)
- `yay` or `paru` (optional, for AUR support)

### Building from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/pacman-tui.git
   cd pacman-tui
   ```

2. Build the binary:
   ```bash
   go build -o pacman-tui .
   ```

3. (Optional) Install it globally:
   ```bash
   sudo cp pacman-tui /usr/local/bin/
   ```

## Usage

Simply run the binary from your terminal:

```bash
pacman-tui
```

### Key Bindings

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Switch tabs (Search, Installed, Update, Cleanup) |
| `/` | Focus search input (in Search and Installed tabs) |
| `↑` / `↓` (or `k` / `j`) | Navigate lists |
| `Enter` | Install package (Search tab) / Remove all orphans (Cleanup tab) |
| `i` | View detailed information for the selected package |
| `d` | Remove the selected package |
| `r` | Refresh the current tab |
| `Space` | Multi-select packages (in Cleanup tab) |
| `?` | Toggle help menu |
| `q` | Quit application |

## License

MIT
