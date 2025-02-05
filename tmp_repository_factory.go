package repository

import (
{{range.imports}}
"{{.}}"
{{- end}}
)

// NewRepository returns an instance of Repository.
// NewRepository is an example of a factory function.
// Use this as an example to create your own repository factory functions
func NewRepository(connection *database.Connection) *Repository {
	return &Repository{
		db: connection,
	}
}
