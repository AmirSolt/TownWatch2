package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func pushToSql() {

	services := getServicesWithSQL()
	var allSQLs []string
	for _, service := range services {
		schemaFilePath, queryFilePath := getServicesSQLPaths(service)
		contentSchema, errSchema := os.ReadFile(schemaFilePath)
		if errSchema != nil {
			log.Fatal(errSchema)
		}
		allSQLs = append(allSQLs, string(contentSchema))

		contentQuery, errQuery := os.ReadFile(queryFilePath)
		if errQuery != nil {
			log.Fatal(errQuery)
		}
		allSQLs = append(allSQLs, string(contentQuery))
	}

	conn := loadDB()
	defer conn.Close(context.Background())

	conn.Exec(context.Background(), strings.Join(allSQLs, "\n"))
}

func resetDb() {

	resetSQL

	conn := loadDB()
	defer conn.Close(context.Background())

	conn.Exec(context.Background(), resetSQL)
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
