package scaffold

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// ParseFields parses the --fields DSL into a slice of Field.
// Format: "name:string,count:int,active:bool,due_at:time?"
func ParseFields(fieldsStr string) ([]Field, error) {
	if fieldsStr == "" {
		return nil, nil
	}

	var fields []Field
	parts := strings.Split(fieldsStr, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		field, err := parseField(part)
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}

	return fields, nil
}

// parseField parses a single field specification.
// Format: "name:type" or "name:type?" for nullable
func parseField(spec string) (Field, error) {
	parts := strings.SplitN(spec, ":", 2)
	if len(parts) != 2 {
		return Field{}, fmt.Errorf("invalid field spec %q: expected 'name:type'", spec)
	}

	name := strings.TrimSpace(parts[0])
	typeSpec := strings.TrimSpace(parts[1])

	if name == "" {
		return Field{}, fmt.Errorf("invalid field spec %q: empty field name", spec)
	}

	// Check for nullable marker
	nullable := strings.HasSuffix(typeSpec, "?")
	if nullable {
		typeSpec = typeSpec[:len(typeSpec)-1]
	}

	// Map DSL type to Go type and SQL type
	goType, sqlType, goNullType, err := mapFieldType(typeSpec, nullable)
	if err != nil {
		return Field{}, fmt.Errorf("invalid field spec %q: %w", spec, err)
	}

	return Field{
		Name:       ToPascalCase(name),
		NameLower:  ToCamelCase(name),
		NameSnake:  ToSnakeCase(name),
		Type:       goType,
		SQLType:    sqlType,
		Nullable:   nullable,
		GoNullType: goNullType,
	}, nil
}

// mapFieldType maps DSL type to Go type and SQL type.
func mapFieldType(dslType string, nullable bool) (goType, sqlType, goNullType string, err error) {
	switch dslType {
	case "string":
		if nullable {
			return "string", "TEXT", "sql.NullString", nil
		}
		return "string", "TEXT NOT NULL", "", nil
	case "int":
		if nullable {
			return "int64", "INTEGER", "sql.NullInt64", nil
		}
		return "int64", "INTEGER NOT NULL", "", nil
	case "bool":
		if nullable {
			return "bool", "INTEGER", "sql.NullBool", nil
		}
		return "bool", "INTEGER NOT NULL DEFAULT 0", "", nil
	case "time":
		if nullable {
			return "time.Time", "DATETIME", "sql.NullTime", nil
		}
		return "time.Time", "DATETIME NOT NULL", "", nil
	default:
		return "", "", "", fmt.Errorf("unknown type %q (valid: string, int, bool, time)", dslType)
	}
}

// ParseStatus parses the --status flag into status values.
// Format: "draft,active,completed"
func ParseStatus(statusStr string) ([]string, error) {
	if statusStr == "" {
		return nil, nil
	}

	var statuses []string
	parts := strings.Split(statusStr, ",")

	for _, part := range parts {
		status := strings.TrimSpace(part)
		if status == "" {
			continue
		}

		// Validate status value (lowercase alphanumeric + underscore)
		if !isValidIdentifier(status) {
			return nil, fmt.Errorf("invalid status value %q: must be lowercase alphanumeric with underscores", status)
		}

		statuses = append(statuses, status)
	}

	if len(statuses) < 2 {
		return nil, fmt.Errorf("at least 2 status values required, got %d", len(statuses))
	}

	return statuses, nil
}

// isValidIdentifier checks if a string is a valid lowercase identifier.
func isValidIdentifier(s string) bool {
	if s == "" {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-z][a-z0-9_]*$`, s)
	return matched
}

// BuildEntitySpec builds an EntitySpec from parsed inputs.
func BuildEntitySpec(name, fieldsStr, statusStr, idPrefix string) (*EntitySpec, error) {
	if name == "" {
		return nil, fmt.Errorf("entity name is required")
	}

	fields, err := ParseFields(fieldsStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse fields: %w", err)
	}

	statuses, err := ParseStatus(statusStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse status: %w", err)
	}

	pascalName := ToPascalCase(name)
	if idPrefix == "" {
		idPrefix = strings.ToUpper(ToSnakeCase(name))
	}

	return &EntitySpec{
		Name:         pascalName,
		NameLower:    ToCamelCase(name),
		NamePlural:   Pluralize(ToSnakeCase(name)),
		NameSnake:    ToSnakeCase(name),
		IDPrefix:     idPrefix,
		Fields:       fields,
		HasStatus:    len(statuses) > 0,
		StatusValues: statuses,
	}, nil
}

// Name transformation helpers

// ToPascalCase converts a string to PascalCase.
func ToPascalCase(s string) string {
	words := splitWords(s)
	for i, word := range words {
		words[i] = capitalize(strings.ToLower(word))
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

// ToCamelCase converts a string to camelCase.
func ToCamelCase(s string) string {
	pascal := ToPascalCase(s)
	if len(pascal) == 0 {
		return pascal
	}
	return strings.ToLower(pascal[:1]) + pascal[1:]
}

// ToSnakeCase converts a string to snake_case.
func ToSnakeCase(s string) string {
	words := splitWords(s)
	for i, word := range words {
		words[i] = strings.ToLower(word)
	}
	return strings.Join(words, "_")
}

// splitWords splits a string into words (handles camelCase, PascalCase, snake_case, kebab-case).
func splitWords(s string) []string {
	// Replace common separators with space
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")

	// Insert space before uppercase letters in camelCase/PascalCase
	var result strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			prev := rune(s[i-1])
			if !unicode.IsSpace(prev) && !unicode.IsUpper(prev) {
				result.WriteRune(' ')
			}
		}
		result.WriteRune(r)
	}

	// Split and filter empty strings
	parts := strings.Fields(result.String())
	return parts
}

// Pluralize returns a simple pluralized form of a word.
func Pluralize(s string) string {
	if s == "" {
		return s
	}

	// Simple pluralization rules
	if strings.HasSuffix(s, "s") || strings.HasSuffix(s, "x") ||
		strings.HasSuffix(s, "ch") || strings.HasSuffix(s, "sh") {
		return s + "es"
	}
	if strings.HasSuffix(s, "y") && len(s) > 1 {
		lastChar := s[len(s)-2]
		if lastChar != 'a' && lastChar != 'e' && lastChar != 'i' && lastChar != 'o' && lastChar != 'u' {
			return s[:len(s)-1] + "ies"
		}
	}
	return s + "s"
}
