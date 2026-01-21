package scaffold

import (
	"bytes"
	"fmt"
	"text/template"

	scaffoldtmpl "github.com/example/orc/internal/templates/scaffold"
)

// Generator generates code from templates.
type Generator struct {
	funcs template.FuncMap
}

// NewGenerator creates a new Generator.
func NewGenerator() *Generator {
	return &Generator{
		funcs: scaffoldtmpl.TemplateFuncs(),
	}
}

// GenerateEntity generates all files for an entity.
func (g *Generator) GenerateEntity(spec *EntitySpec) (*GeneratorResult, error) {
	result := &GeneratorResult{}

	// Generate new files
	newFiles := []struct {
		template string
		path     string
	}{
		{"model.go", fmt.Sprintf("internal/models/%s.go", spec.NameSnake)},
		{"primary_port.go", fmt.Sprintf("internal/ports/primary/%s.go", spec.NameSnake)},
		{"service.go", fmt.Sprintf("internal/app/%s_service.go", spec.NameSnake)},
		{"sqlite_repo.go", fmt.Sprintf("internal/adapters/sqlite/%s_repo.go", spec.NameSnake)},
		{"cli.go", fmt.Sprintf("internal/cli/%s.go", spec.NameSnake)},
	}

	for _, f := range newFiles {
		content, err := g.renderTemplate(f.template, spec)
		if err != nil {
			return nil, fmt.Errorf("failed to render %s: %w", f.template, err)
		}
		result.Files = append(result.Files, GeneratedFile{
			Path:      f.path,
			Content:   content,
			IsNew:     true,
			Operation: "create",
		})
	}

	// Generate snippets for existing files
	snippets := []struct {
		template  string
		path      string
		insertAt  string
		operation string
	}{
		{"secondary_port_snippet.go", "internal/ports/secondary/persistence.go", "// END SCAFFOLD MARKER", "insert_before"},
		{"wire_snippet.go", "internal/wire/wire.go", "// END SCAFFOLD MARKER", "insert_before"},
		{"main_snippet.go", "cmd/orc/main.go", "// END SCAFFOLD MARKER", "insert_before"},
	}

	for _, s := range snippets {
		content, err := g.renderTemplate(s.template, spec)
		if err != nil {
			return nil, fmt.Errorf("failed to render %s: %w", s.template, err)
		}
		result.Files = append(result.Files, GeneratedFile{
			Path:      s.path,
			Snippet:   content,
			InsertAt:  s.insertAt,
			Operation: s.operation,
		})
	}

	// Add next steps
	result.NextSteps = []string{
		fmt.Sprintf("Run 'orc scaffold migration create_%s_table' to create DB migration", spec.NamePlural),
		"Run 'make dev && ./orc " + spec.NameSnake + " --help' to test",
	}

	return result, nil
}

// GenerateMigration generates a migration file.
func (g *Generator) GenerateMigration(spec *MigrationSpec, entitySpec *EntitySpec) (*GeneratorResult, error) {
	result := &GeneratorResult{}

	// Combine specs for template
	data := struct {
		*MigrationSpec
		Entity *EntitySpec
	}{
		MigrationSpec: spec,
		Entity:        entitySpec,
	}

	content, err := g.renderMigrationTemplate(data)
	if err != nil {
		return nil, fmt.Errorf("failed to render migration: %w", err)
	}

	result.Files = append(result.Files, GeneratedFile{
		Path:      "internal/db/migrations.go",
		Snippet:   content,
		InsertAt:  "// END MIGRATION MARKER",
		Operation: "insert_before",
	})

	result.NextSteps = []string{
		"Edit migrationV" + fmt.Sprint(spec.Version) + "() in internal/db/migrations.go to add your schema changes",
		"Run './orc init' to apply migration (on dev DB)",
	}

	return result, nil
}

// renderTemplate renders an entity template.
func (g *Generator) renderTemplate(name string, spec *EntitySpec) (string, error) {
	tmplContent, err := scaffoldtmpl.GetEntityTemplate(name)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New(name).Funcs(g.funcs).Parse(tmplContent)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, spec); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// renderMigrationTemplate renders the migration template.
func (g *Generator) renderMigrationTemplate(data any) (string, error) {
	tmplContent, err := scaffoldtmpl.GetMigrationTemplate()
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("migration").Funcs(g.funcs).Parse(tmplContent)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
