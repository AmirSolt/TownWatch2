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

	base.DB = &DB{
		Conn: conn,
	}
}

func (base *Base) killDB() {
	base.Conn.Close(context.Background())
}
