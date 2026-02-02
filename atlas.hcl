// Atlas project configuration
// Reference: AGENTS.md "Database Migrations (Atlas)" section

variable "home" {
  type    = string
  default = getenv("HOME")
}

env "local" {
  src = "file://internal/db/schema.sql"
  url = "sqlite:///${var.home}/.orc/orc.db"
  dev = "sqlite://dev?mode=memory"

  // CRITICAL: Exclude SQLite autoindexes to avoid migration errors
  // See AGENTS.md for explanation
  exclude = ["*.sqlite_autoindex*[type=index]"]
}

env "test" {
  src = "file://internal/db/schema.sql"
  url = "sqlite://:memory:?_fk=1"
  dev = "sqlite://dev?mode=memory"

  exclude = ["*.sqlite_autoindex*[type=index]"]
}
