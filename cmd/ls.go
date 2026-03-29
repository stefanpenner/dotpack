package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/stefanpenner/devlayer/internal/config"
	"github.com/stefanpenner/devlayer/internal/nvimplugins"
)

var (
	heading = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")). // bright blue
		MarginBottom(1)

	subtext = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")) // dim gray

	checkMark = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")). // green
			SetString("✓")

	crossMark = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")). // red
			SetString("✗")

	runtimeBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")). // cyan
			Bold(true)

	dimText = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))
)

// Ls lists installed devlayer tools, dotfiles, and nvim plugins.
func Ls() error {
	prefix := defaultPrefix()

	if err := lsTools(prefix); err != nil {
		return err
	}
	fmt.Println()
	lsDotfiles()
	fmt.Println()
	lsNvimPlugins()

	return nil
}

// lsTools lists binaries and tool directories under the prefix.
func lsTools(prefix string) error {
	fmt.Println(heading.Render("Tools") + " " + subtext.Render(filepath.Join(prefix, "bin")))

	binDir := filepath.Join(prefix, "bin")
	entries, err := os.ReadDir(binDir)
	if os.IsNotExist(err) {
		fmt.Println(dimText.Render("  (no tools installed)"))
		return nil
	}
	if err != nil {
		return err
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		// Skip macOS resource fork files
		if strings.HasPrefix(name, "._") {
			continue
		}
		if runtime.GOOS == "windows" {
			name = strings.TrimSuffix(name, ".cmd")
			name = strings.TrimSuffix(name, ".exe")
		}
		names = append(names, name)
	}
	sort.Strings(names)

	if len(names) == 0 {
		fmt.Println(dimText.Render("  (no tools installed)"))
		return nil
	}

	// Render tools as a compact multi-column table
	cols := 4
	rows := (len(names) + cols - 1) / cols

	t := table.New().
		Border(lipgloss.HiddenBorder()).
		StyleFunc(func(row, col int) lipgloss.Style {
			return lipgloss.NewStyle().PaddingRight(2)
		})

	for r := range rows {
		row := make([]string, cols)
		for c := range cols {
			idx := c*rows + r
			if idx < len(names) {
				row[c] = names[idx]
			}
		}
		t.Row(row...)
	}
	fmt.Println(t)

	// Show bundled subdirectories
	subdirs := []string{"go", "git", "zsh", "nvim", "zig", "share"}
	var present []string
	for _, d := range subdirs {
		if info, err := os.Stat(filepath.Join(prefix, d)); err == nil && info.IsDir() {
			present = append(present, runtimeBadge.Render(d))
		}
	}
	if len(present) > 0 {
		fmt.Println(subtext.Render("  Bundled runtimes: ") + strings.Join(present, subtext.Render(", ")))
	}

	return nil
}

// lsDotfiles lists synced dotfiles from the config.
func lsDotfiles() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println(heading.Render("Dotfiles"))
		fmt.Printf("  %s\n", dimText.Render("error loading config: "+err.Error()))
		return
	}

	fmt.Println(heading.Render("Dotfiles") + " " + subtext.Render(config.Path()))

	if len(cfg.Dotfiles.Sync) == 0 {
		fmt.Println(dimText.Render("  (none configured)"))
		return
	}

	home, _ := os.UserHomeDir()
	items := make([]any, 0, len(cfg.Dotfiles.Sync))
	for _, rel := range cfg.Dotfiles.Sync {
		full := filepath.Join(home, rel)
		marker := checkMark.String()
		if _, err := os.Stat(full); os.IsNotExist(err) {
			marker = crossMark.String() + " " + dimText.Render("missing")
		}
		items = append(items, rel+" "+marker)
	}

	l := list.New(items...).Enumerator(list.Dash)
	fmt.Println(l)
}

// lsNvimPlugins lists installed nvim plugins from the vim.pack directory.
func lsNvimPlugins() {
	pluginDir := nvimplugins.LocalPluginDir()
	fmt.Println(heading.Render("Nvim Plugins") + " " + subtext.Render(pluginDir))

	entries, err := os.ReadDir(pluginDir)
	if os.IsNotExist(err) {
		fmt.Println(dimText.Render("  (no plugins installed)"))
		return
	}
	if err != nil {
		fmt.Printf("  %s\n", dimText.Render("error: "+err.Error()))
		return
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	if len(names) == 0 {
		fmt.Println(dimText.Render("  (no plugins installed)"))
		return
	}

	// Multi-column table for plugins
	cols := 3
	rows := (len(names) + cols - 1) / cols

	t := table.New().
		Border(lipgloss.HiddenBorder()).
		StyleFunc(func(row, col int) lipgloss.Style {
			return lipgloss.NewStyle().PaddingRight(2)
		})

	for r := range rows {
		row := make([]string, cols)
		for c := range cols {
			idx := c*rows + r
			if idx < len(names) {
				row[c] = names[idx]
			}
		}
		t.Row(row...)
	}
	fmt.Println(t)

	fmt.Println(subtext.Render(fmt.Sprintf("  %d plugins", len(names))))
}
