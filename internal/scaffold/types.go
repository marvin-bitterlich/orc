// Package scaffold provides code generation for ORC entities.
package scaffold

// EntitySpec contains all information needed to generate an entity.
type EntitySpec struct {
	Name         string   // PascalCase: "Widget"
	NameLower    string   // camelCase: "widget"
	NamePlural   string   // lowercase plural: "widgets"
	NameSnake    string   // snake_case: "widget"
	IDPrefix     string   // ID prefix: "WIDGET"
	Fields       []Field  // Custom fields
	HasStatus    bool     // Whether entity has FSM status
	StatusValues []string // e.g., ["draft", "active", "completed"]

	// Parent relationship (optional)
	HasParent         bool   // Whether entity has a parent FK
	ParentEntity      string // PascalCase: "Shipment"
	ParentEntityLower string // camelCase: "shipment"
	ParentFK          string // snake_case: "shipment_id"
	ParentTable       string // Plural snake: "shipments"
	ParentCardinality string // "1:1" or "n:1"
}

// Field represents a field in an entity.
type Field struct {
	Name       string // PascalCase: "Title"
	NameLower  string // camelCase: "title"
	NameSnake  string // snake_case: "title"
	Type       string // Go type: string, int, bool, time.Time
	SQLType    string // SQLite: TEXT, INTEGER, DATETIME
	Nullable   bool   // Whether the field is optional
	GoNullType string // sql.NullString, sql.NullInt64, etc.
}

// GeneratedFile represents a file to be created or modified.
type GeneratedFile struct {
	Path      string // File path relative to project root
	Content   string // File content
	IsNew     bool   // True if creating new file, false if appending
	Snippet   string // For append operations, the snippet to add
	InsertAt  string // Marker string for insertion point
	Operation string // "create", "append", "insert_after"
}

// GeneratorResult contains the result of a scaffold operation.
type GeneratorResult struct {
	Files     []GeneratedFile
	NextSteps []string
}
