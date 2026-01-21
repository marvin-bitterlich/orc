package scaffold

import (
	"testing"
)

func TestParseFields(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{"empty", "", 0, false},
		{"single string", "name:string", 1, false},
		{"multiple fields", "name:string,count:int,active:bool", 3, false},
		{"with nullable", "name:string,due_at:time?", 2, false},
		{"invalid type", "name:unknown", 0, true},
		{"invalid format", "namestring", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields, err := ParseFields(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(fields) != tt.want {
				t.Errorf("ParseFields() got %d fields, want %d", len(fields), tt.want)
			}
		})
	}
}

func TestParseFieldTypes(t *testing.T) {
	fields, err := ParseFields("name:string,count:int,active:bool,due:time,desc:string?")
	if err != nil {
		t.Fatalf("ParseFields() error = %v", err)
	}

	expected := []struct {
		name     string
		goType   string
		sqlType  string
		nullable bool
	}{
		{"Name", "string", "TEXT NOT NULL", false},
		{"Count", "int64", "INTEGER NOT NULL", false},
		{"Active", "bool", "INTEGER NOT NULL DEFAULT 0", false},
		{"Due", "time.Time", "DATETIME NOT NULL", false},
		{"Desc", "string", "TEXT", true},
	}

	for i, exp := range expected {
		if fields[i].Name != exp.name {
			t.Errorf("fields[%d].Name = %q, want %q", i, fields[i].Name, exp.name)
		}
		if fields[i].Type != exp.goType {
			t.Errorf("fields[%d].Type = %q, want %q", i, fields[i].Type, exp.goType)
		}
		if fields[i].Nullable != exp.nullable {
			t.Errorf("fields[%d].Nullable = %v, want %v", i, fields[i].Nullable, exp.nullable)
		}
	}
}

func TestParseStatus(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{"empty", "", 0, false},
		{"valid", "draft,active,completed", 3, false},
		{"with spaces", "draft, active, completed", 3, false},
		{"too few", "draft", 0, true},
		{"invalid chars", "Draft,Active", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statuses, err := ParseStatus(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(statuses) != tt.want {
				t.Errorf("ParseStatus() got %d statuses, want %d", len(statuses), tt.want)
			}
		})
	}
}

func TestNameTransformations(t *testing.T) {
	tests := []struct {
		input  string
		pascal string
		camel  string
		snake  string
	}{
		{"widget", "Widget", "widget", "widget"},
		{"my_widget", "MyWidget", "myWidget", "my_widget"},
		{"myWidget", "MyWidget", "myWidget", "my_widget"},
		{"MyWidget", "MyWidget", "myWidget", "my_widget"},
		{"my-widget", "MyWidget", "myWidget", "my_widget"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ToPascalCase(tt.input); got != tt.pascal {
				t.Errorf("ToPascalCase(%q) = %q, want %q", tt.input, got, tt.pascal)
			}
			if got := ToCamelCase(tt.input); got != tt.camel {
				t.Errorf("ToCamelCase(%q) = %q, want %q", tt.input, got, tt.camel)
			}
			if got := ToSnakeCase(tt.input); got != tt.snake {
				t.Errorf("ToSnakeCase(%q) = %q, want %q", tt.input, got, tt.snake)
			}
		})
	}
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"widget", "widgets"},
		{"box", "boxes"},
		{"class", "classes"},
		{"bus", "buses"},
		{"church", "churches"},
		{"dish", "dishes"},
		{"party", "parties"},
		{"day", "days"},
		{"key", "keys"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := Pluralize(tt.input); got != tt.want {
				t.Errorf("Pluralize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBuildEntitySpec(t *testing.T) {
	spec, err := BuildEntitySpec("widget", "name:string,value:int", "draft,active,done", "")
	if err != nil {
		t.Fatalf("BuildEntitySpec() error = %v", err)
	}

	if spec.Name != "Widget" {
		t.Errorf("Name = %q, want %q", spec.Name, "Widget")
	}
	if spec.NameLower != "widget" {
		t.Errorf("NameLower = %q, want %q", spec.NameLower, "widget")
	}
	if spec.NamePlural != "widgets" {
		t.Errorf("NamePlural = %q, want %q", spec.NamePlural, "widgets")
	}
	if spec.IDPrefix != "WIDGET" {
		t.Errorf("IDPrefix = %q, want %q", spec.IDPrefix, "WIDGET")
	}
	if len(spec.Fields) != 2 {
		t.Errorf("len(Fields) = %d, want 2", len(spec.Fields))
	}
	if !spec.HasStatus {
		t.Error("HasStatus = false, want true")
	}
	if len(spec.StatusValues) != 3 {
		t.Errorf("len(StatusValues) = %d, want 3", len(spec.StatusValues))
	}
}
