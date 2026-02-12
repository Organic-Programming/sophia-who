package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// holonTemplate generates the complete HOLON.md file.
// The YAML frontmatter is the machine-readable identity.
// The markdown body is the human-readable description.
var holonTemplate = `---
# Holon Identity v1
uuid: {{ .UUID | quote }}
given_name: {{ .GivenName | quote }}
family_name: {{ .FamilyName | quote }}
motto: {{ .Motto | quote }}
composer: {{ .Composer | quote }}
clade: {{ .Clade | quote }}
status: {{ .Status }}
born: {{ .Born | quote }}

# Lineage
parents: [{{ joinQuoted .Parents }}]
reproduction: {{ .Reproduction | quote }}

# Pinning
binary_path: {{ if .BinaryPath }}{{ .BinaryPath | quote }}{{ else }}null{{ end }}
binary_version: {{ if .BinaryVersion }}{{ .BinaryVersion | quote }}{{ else }}null{{ end }}
git_tag: {{ if .GitTag }}{{ .GitTag | quote }}{{ else }}null{{ end }}
git_commit: {{ if .GitCommit }}{{ .GitCommit | quote }}{{ else }}null{{ end }}
os: {{ if .OS }}{{ .OS | quote }}{{ else }}null{{ end }}
arch: {{ if .Arch }}{{ .Arch | quote }}{{ else }}null{{ end }}
dependencies: [{{ joinQuoted .Dependencies }}]

# Optional
aliases: [{{ joinQuoted .Aliases }}]
wrapped_license: {{ if .WrappedLicense }}{{ .WrappedLicense | quote }}{{ else }}null{{ end }}

# Metadata
generated_by: {{ .GeneratedBy | quote }}
lang: {{ .Lang | quote }}
proto_status: {{ .ProtoStatus }}
---

# {{ .GivenName }} {{ .FamilyName }}

> *"{{ .Motto }}"*

## Description

<Describe what this holon does.>

## Introspection Notes

<Any assumptions or ambiguities noted during creation.>
`

var tmplFuncs = template.FuncMap{
	"quote": func(s string) string {
		return fmt.Sprintf("%q", s)
	},
	"joinQuoted": func(ss []string) string {
		quoted := make([]string, len(ss))
		for i, s := range ss {
			quoted[i] = fmt.Sprintf("%q", s)
		}
		return strings.Join(quoted, ", ")
	},
}

// runNew interactively creates a new holon identity.
func runNew() error {
	scanner := bufio.NewScanner(os.Stdin)
	id := NewIdentity()

	fmt.Println("─── Sophia Who? — New Holon Identity ───")
	fmt.Printf("UUID: %s (generated)\n\n", id.UUID)

	id.FamilyName = ask(scanner, "Family name (the function — e.g. Transcriber, Prober)")
	id.GivenName = ask(scanner, "Given name (the character — e.g. Swift, Deep)")
	id.Composer = ask(scanner, "Composer (who is making this decision?)")
	id.Motto = ask(scanner, "Motto (the dessein in one sentence)")

	fmt.Println("\nClade (computational nature):")
	for i, c := range Clades {
		fmt.Printf("  %d. %s\n", i+1, c)
	}
	id.Clade = askChoice(scanner, "Choose clade", Clades)

	fmt.Println("\nReproduction mode:")
	for i, r := range ReproductionModes {
		fmt.Printf("  %d. %s\n", i+1, r)
	}
	id.Reproduction = askChoice(scanner, "Choose reproduction mode", ReproductionModes)

	id.Lang = askDefault(scanner, "Implementation language", "go")

	aliases := ask(scanner, "Aliases (comma-separated, or empty)")
	if aliases != "" {
		for _, a := range strings.Split(aliases, ",") {
			if trimmed := strings.TrimSpace(a); trimmed != "" {
				id.Aliases = append(id.Aliases, trimmed)
			}
		}
	}

	license := ask(scanner, "Wrapped binary license (e.g. MIT, GPL-3.0, or empty)")
	if license != "" {
		id.WrappedLicense = license
	}

	// Determine output path
	dirName := strings.ToLower(id.GivenName + "-" + strings.TrimSuffix(id.FamilyName, "?"))
	dirName = strings.ReplaceAll(dirName, " ", "-")
	outputDir := askDefault(scanner, "Output directory", filepath.Join(".holon", dirName))

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("cannot create directory %s: %w", outputDir, err)
	}

	outputPath := filepath.Join(outputDir, "HOLON.md")

	tmpl, err := template.New("holon").Funcs(tmplFuncs).Parse(holonTemplate)
	if err != nil {
		return fmt.Errorf("template error: %w", err)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("cannot create %s: %w", outputPath, err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, id); err != nil {
		return fmt.Errorf("template execution error: %w", err)
	}

	fmt.Printf("\n✓ Born: %s %s\n", id.GivenName, id.FamilyName)
	fmt.Printf("  UUID: %s\n", id.UUID)
	fmt.Printf("  File: %s\n", outputPath)

	return nil
}

// runShow reads and displays a holon's identity by UUID.
func runShow(target string) error {
	path, err := findHolonByUUID(target)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", path, err)
	}

	fmt.Println(string(data))
	return nil
}

// runList scans the current project for HOLON.md files and prints a summary.
func runList() error {
	holons, err := findAllHolons()
	if err != nil {
		return err
	}

	if len(holons) == 0 {
		fmt.Println("No holons found.")
		return nil
	}

	fmt.Printf("%-38s %-20s %-30s %s\n", "UUID", "NAME", "CLADE", "STATUS")
	fmt.Println(strings.Repeat("─", 100))

	for _, h := range holons {
		name := h.GivenName + " " + h.FamilyName
		fmt.Printf("%-38s %-20s %-30s %s\n", h.UUID, name, h.Clade, h.Status)
	}

	return nil
}

// runPin captures version, OS, and architecture information for a holon's binary.
func runPin(target string) error {
	path, err := findHolonByUUID(target)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", path, err)
	}

	// Extract YAML frontmatter
	id, body, err := parseFrontmatter(data)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("─── Pin version for %s %s ───\n\n", id.GivenName, id.FamilyName)

	id.BinaryPath = askDefault(scanner, "Binary path", id.BinaryPath)
	id.BinaryVersion = askDefault(scanner, "Binary version", id.BinaryVersion)
	id.GitTag = askDefault(scanner, "Git tag (or empty)", id.GitTag)
	id.GitCommit = askDefault(scanner, "Git commit (or empty)", id.GitCommit)
	id.OS = askDefault(scanner, "OS", id.OS)
	id.Arch = askDefault(scanner, "Arch", id.Arch)

	// Rewrite the file with updated frontmatter
	yamlData, err := yaml.Marshal(id)
	if err != nil {
		return fmt.Errorf("yaml marshal error: %w", err)
	}

	output := "---\n# Holon Identity v1\n" + string(yamlData) + "---\n" + body

	if err := os.WriteFile(path, []byte(output), 0644); err != nil {
		return fmt.Errorf("cannot write %s: %w", path, err)
	}

	fmt.Printf("\n✓ Pinned: %s %s\n", id.GivenName, id.FamilyName)
	return nil
}

// ask prompts the user and returns the answer (required, no default).
func ask(scanner *bufio.Scanner, prompt string) string {
	for {
		fmt.Printf("%s: ", prompt)
		scanner.Scan()
		answer := strings.TrimSpace(scanner.Text())
		if answer != "" {
			return answer
		}
		fmt.Println("  (required)")
	}
}

// askDefault prompts the user with a default value.
func askDefault(scanner *bufio.Scanner, prompt, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultVal)
	} else {
		fmt.Printf("%s: ", prompt)
	}
	scanner.Scan()
	answer := strings.TrimSpace(scanner.Text())
	if answer == "" {
		return defaultVal
	}
	return answer
}

// askChoice prompts the user to choose from a numbered list.
func askChoice(scanner *bufio.Scanner, prompt string, choices []string) string {
	for {
		fmt.Printf("%s (1-%d): ", prompt, len(choices))
		scanner.Scan()
		answer := strings.TrimSpace(scanner.Text())
		for i, c := range choices {
			if answer == fmt.Sprintf("%d", i+1) || answer == c {
				return c
			}
		}
		fmt.Println("  (invalid choice)")
	}
}
