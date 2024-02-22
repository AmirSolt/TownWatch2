package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func pushToDB() {

	services := getServicesWithSQL()
	var allSQLs []string
	for _, service := range services {
		schemaFilePath, _ := getServicesSQLPaths(service)
		contentSchema, errSchema := os.ReadFile(schemaFilePath)
		if errSchema != nil {
			log.Fatal(errSchema)
		}
		allSQLs = append(allSQLs, string(contentSchema))

		// contentQuery, errQuery := os.ReadFile(queryFilePath)
		// if errQuery != nil {
		// 	log.Fatal(errQuery)
		// }
		// allSQLs = append(allSQLs, string(contentQuery))
	}

	conn := loadDB()
	defer conn.Close(context.Background())

	response, err := conn.Exec(context.Background(), strings.Join(allSQLs, "\n"))
	if err != nil {
		log.Fatal("Error db:", err)
	}
	fmt.Println("......")
	fmt.Println(response)
	fmt.Println("......")
}

// func resetDb() {

// 	resetSQL := "DROP DATABASE postgres; CREATE DATABASE postgres;"

// 	conn := loadDB()
// 	defer conn.Close(context.Background())

// 	response, err := conn.Exec(context.Background(), resetSQL)
// 	if err != nil {
// 		log.Fatal("Error db:", err)
// 	}
// 	fmt.Println("......")
// 	fmt.Println(response)
// 	fmt.Println("......")
// }

func resetDb() {

	ctx := context.Background()

	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error db:", err)
	}

	config, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Unable to parse configuration: %v", err)
	}

	// Open connection pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	// execResponseDrop, err := pool.Exec(ctx, "DROP DATABASE IF EXISTS postgres")
	execResponseDrop, err := pool.Exec(ctx, "DROP SCHEMA public CASCADE;")
	if err != nil {
		log.Fatalf("error dropping database: %v", err)
	}
	fmt.Println(">>>", execResponseDrop)

	// Create the database
	// execResponseCreate, err := pool.Exec(ctx, "CREATE DATABASE postgres")
	execResponseCreate, err := pool.Exec(ctx, "CREATE SCHEMA public;")
	if err != nil {
		log.Fatalf("error creating database: %v", err)
	}
	fmt.Println(">>>", execResponseCreate)
}

func loadDB() *pgx.Conn {

	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error db:", err)
	}

	conn, dbErr := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if dbErr != nil {
		log.Fatalln("Error db:", dbErr)
	}

	// defer conn.Close(context.Background())

	// conn.Exec(context.Background(), "")

	return conn
}
