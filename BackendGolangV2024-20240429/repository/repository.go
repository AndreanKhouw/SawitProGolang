// This file contains the repository implementation layer.
package repository

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

type Repository struct {
	Db *sql.DB
}

type NewRepositoryOptions struct {
	Dsn string
}

func NewRepository(opts NewRepositoryOptions) *Repository {
	db, err := sql.Open("postgres", opts.Dsn)
	if err != nil {
		log.Fatalf("Unable to open database: %v\n", err)
	}
	return &Repository{
		Db: db,
	}
}
