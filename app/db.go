package app

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

type DB struct {
	// Queries *models.Queries
	Conn *pgx.Conn
}

func (app *App) loadDB() {
	conn, dbErr := pgx.Connect(context.Background(), app.DATABASE_URL)
	if dbErr != nil {
		log.Fatalln("Error db:", dbErr)
	}

	// queries := models.New(conn)
	app.DB = &DB{
		Conn: conn,
		// Queries: queries,
	}
}

func (app *App) killDB() {
	app.Conn.Close(context.Background())
}
