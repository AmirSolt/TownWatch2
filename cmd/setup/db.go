package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func pushToSql() {

	conn := loadDB()
	defer conn.Close(context.Background())

	path := filepath.Join("path", "to", "script.sql")

	c, ioErr := ioutil.ReadFile(path)
	if ioErr != nil {
		// handle error.
	}
}

func resetDb() {

	conn := loadDB()
	defer conn.Close(context.Background())

	path := filepath.Join("path", "to", "script.sql")

	c, ioErr := os.ReadFile(path)
	if ioErr != nil {
		// handle error.
	}
}

func loadDB() *pgx.Conn {

	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error .env:", err)
	}

	conn, dbErr := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if dbErr != nil {
		log.Fatalln("Error db:", dbErr)
	}

	// defer conn.Close(context.Background())

	// conn.Exec(context.Background(), "")

	return conn
}
