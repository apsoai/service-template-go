package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	if err := goose.Run(command, db, "migrations"); err != nil {
		log.Fatalf("goose %s: %v", command, err)
	}
}
