package repository

import (
	"time"
	{{range .imports }}
	"{{.}}"
	{{- end}}
)

type (
	// Repository types will implement interface defined in domain package
	// Repository is a struct that represents a repository. use this as an example to create your own repository structs
	Repository struct {
		db *database.Connection
	}

	// row is a struct that represents a row in the database. use this as an example to create your own row structs
	row struct {
		ID        int       `db:"id"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}
)
