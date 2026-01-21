// Package scaffold provides templates for code generation.
package scaffold

import (
	"embed"
	"strings"
	"text/template"
)

//go:embed entity/*.tmpl migration/*.tmpl
var scaffoldTemplates embed.FS

// GetEntityTemplate returns the content of an entity template.
func GetEntityTemplate(name string) (string, error) {
	content, err := scaffoldTemplates.ReadFile("entity/" + name + ".tmpl")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// GetMigrationTemplate returns the content of the migration template.
func GetMigrationTemplate() (string, error) {
	content, err := scaffoldTemplates.ReadFile("migration/migration.go.tmpl")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// TemplateFuncs returns the template function map for scaffold templates.
func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"toLower":     strings.ToLower,
		"toUpper":     strings.ToUpper,
		"title":       capitalize,
		"join":        strings.Join,
		"statusList":  formatStatusList,
		"statusConst": formatStatusConst,
		"repeat":      strings.Repeat,
		"add":         func(a, b int) int { return a + b },
		"sub":         func(a, b int) int { return a - b },
	}
}

// formatStatusList formats status values for SQL CHECK constraint.
// e.g., ["draft", "active"] -> "'draft', 'active'"
func formatStatusList(statuses []string) string {
	quoted := make([]string, len(statuses))
	for i, s := range statuses {
		quoted[i] = "'" + s + "'"
	}
	return strings.Join(quoted, ", ")
}

// formatStatusConst formats a status value as a Go constant name.
// e.g., "in_progress" -> "InProgress"
func formatStatusConst(status string) string {
	words := strings.Split(status, "_")
	for i, word := range words {
		words[i] = capitalize(word)
	}
	return strings.Join(words, "")
}

// capitalize returns the string with the first letter uppercased.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
