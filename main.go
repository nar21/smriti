package main

import (
//	"database/sql"
	"fmt"
	"log"
	"db-archive/database"
	"db-archive/extract"
)

type Aircraft struct {
    Code string
    Model string
    Range int
}

func main() {
    extract.HelloTest()

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

    rows, err := db.Query("SELECT aircraft_code, model, range FROM aircrafts_data;")
    if err != nil {
        log.Fatal("Failed to execute query:", err)
    }
    defer rows.Close()

    var aircrafts []Aircraft

    // Loop through each row
    for rows.Next() {
        var aircraft Aircraft
        err := rows.Scan(&aircraft.Code, &aircraft.Model, &aircraft.Range)
        if err != nil {
            log.Fatal("Failed to scan row:", err)
        }
        aircrafts = append(aircrafts, aircraft)
    }

    // Check for errors after loop ends
    if err = rows.Err(); err != nil {
        log.Fatal("Rows iteration error:", err)
    }

    // Now 'aircrafts' slice has all the rows
    for _, a := range aircrafts {
        fmt.Printf("Aircraft: Code=%s, Model=%s, Range=%d\n", a.Code, a.Model, a.Range)
    }

}

