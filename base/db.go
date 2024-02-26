package base

// Database connection to postgres

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	// Queries *models.Queries
	Pool *pgxpool.Pool
}

func (base *Base) loadDB() {
	pool, dbErr := pgxpool.New(context.Background(), base.DATABASE_URL)
	if dbErr != nil {
		log.Fatalln("Error db:", dbErr)
	}

	base.DB = &DB{
		Pool: pool,
	}
}

func (base *Base) killDB() {
	base.Pool.Close()
}
