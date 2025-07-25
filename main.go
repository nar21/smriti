package main

import (
	"database/sql"
	"db-archive/database"
	"db-archive/extract"
	"db-archive/parser"
	"fmt"
	"log"
)

type Aircraft struct {
	Code  string
	Model string
	Range int
}

func get_aircraft(db *sql.DB) {
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

func main() {
	extract.HelloTest()

	var ap, err2 = parser.LoadArchivalPlan("archival-plan/aircraft.yaml")
	if err2 != nil {
		log.Fatal("Could not load archival plan")
	}
	fmt.Println(ap.DatabaseCredential)

	// Database connection parameters
	host := ap.DatabaseCredential.Host
	port := ap.DatabaseCredential.Port
	user := ap.DatabaseCredential.User
	password := ap.DatabaseCredential.Password
	dbname := ap.DatabaseCredential.DBName

	// Get a DB connection object from database package
	db, err := database.ConnectPostgres(host, port, user, password, dbname)
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}
	defer db.Close() // Ensure DB connection is closed when done

	var query = "SELECT * FROM flights;"

	data, columns, err := extract.QueryDynamic(db, query)
	if err != nil {
		log.Fatal("Query failed:", err)
	}

	// Export to CSV
	err = extract.WriteCSV("output.csv", data, columns)
	if err != nil {
		log.Fatal("Failed to write CSV:", err)
	}

	fmt.Println("CSV Exported Successfully!")

}
