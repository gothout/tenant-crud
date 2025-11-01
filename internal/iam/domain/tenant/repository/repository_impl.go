package repository

import "database/sql"

type impl struct {
	db *sql.DB
}

func New(db *sql.DB) Repository {
	return &impl{
		db: db,
	}
}
