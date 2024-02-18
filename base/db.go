package base

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

type DB struct {
	// Queries *models.Queries
	Conn *pgx.Conn
}

func (base *Base) loadDB() {
	conn, dbErr := pgx.Connect(context.Background(), base.DATABASE_URL)
	if dbErr != nil {
		log.Fatalln("Error db:", dbErr)
	}

	// queries := models.New(conn)
	base.DB = &DB{
		Conn: conn,
		// Queries: queries,
	}
}

func (base *Base) killDB() {
	base.Conn.Close(context.Background())
}
