package main

import (
//	"database/sql"
	"fmt"
	"log"

	"db-archive/database" // Replace with your actual module name
)

type Aircraft struct {

}

func main() {
	// Database connection parameters
	host := "localhost"
	port := 54321
	user := "airline"
	password := "airline"
	dbname := "airline"

	// Get a DB connection object from database package
	db, err := database.ConnectPostgres(host, port, user, password, dbname)
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}
	defer db.Close() // Ensure DB connection is closed when done

	// Query current time from DB
	var currentTime string
	err = db.QueryRow("SELECT NOW()").Scan(&currentTime)
	if err != nil {
		log.Fatal("Failed to query time:", err)
	}

	fmt.Println("Current Time from DB:", currentTime)
}

