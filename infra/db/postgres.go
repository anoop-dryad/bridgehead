package db

// https://dev.to/jones_charles_ad50858dbc0/sqlx-your-go-to-database-toolkit-for-go-developers-53n8

import (
	"github.com/anoop-dryad/bridgehead/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewPostgresPool(cfg config.DB) *sqlx.DB {
	db, err := sqlx.Connect("postgres", cfg.DSN)
	if err != nil {
		panic("failed to connect to postgres: " + err.Error())
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	return db
}
