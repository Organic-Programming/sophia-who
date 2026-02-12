package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// findAllHolons scans the current directory tree for HOLON.md files
// and returns the parsed identities.
func findAllHolons() ([]Identity, error) {
	var holons []Identity

	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable directories
		}
		if d.IsDir() {
			// Skip hidden directories (except .holon/ and the root ".")
			name := d.Name()
			if name != "." && name != ".holon" && strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() != "HOLON.md" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil // skip unreadable files
		}

		id, _, err := parseFrontmatter(data)
		if err != nil {
			return nil // skip unparsable files
		}

		holons = append(holons, id)
		return nil
	})

	return holons, err
}

// findHolonByUUID locates a HOLON.md file by UUID.
// Searches the current directory tree.
func findHolonByUUID(target string) (string, error) {
	var found string

	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || d.Name() != "HOLON.md" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		id, _, err := parseFrontmatter(data)
		if err != nil {
			return nil
		}

		// Match by full UUID or prefix
		if id.UUID == target || strings.HasPrefix(id.UUID, target) {
			found = path
			return filepath.SkipAll
		}

		return nil
	})

	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("holon not found: %s", target)
	}
	return found, nil
}

// parseFrontmatter extracts the YAML frontmatter and the remaining
// markdown body from a HOLON.md file.
func parseFrontmatter(data []byte) (Identity, string, error) {
	content := string(data)

	if !strings.HasPrefix(content, "---") {
		return Identity{}, "", fmt.Errorf("no YAML frontmatter found")
	}

	// Skip opening "---\n"
	rest := content[3:]
	if len(rest) > 0 && rest[0] == '\n' {
		rest = rest[1:]
	}

	// Find the closing "---"
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return Identity{}, "", fmt.Errorf("unclosed YAML frontmatter")
	}

	yamlBlock := rest[:end]
	body := rest[end+4:] // skip \n---

	var id Identity
	if err := yaml.Unmarshal([]byte(yamlBlock), &id); err != nil {
		return Identity{}, "", fmt.Errorf("YAML parse error: %w", err)
	}

	return id, body, nil
}
