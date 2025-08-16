package core

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func StartDB() {
	log.Println("attempting to connect to Postgres")
	conn, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		panic("failed to connect to db")
	}
	DB = conn
	log.Printf("connected to Postgres")
}
