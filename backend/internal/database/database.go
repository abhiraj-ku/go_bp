package database

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type Database struct {
	pool *pgxpool.Pool
	log  *zerolog.Logger
}
