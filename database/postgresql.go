package database

import (
	"database/sql"
	"fmt"
)

type PostgresDriver struct{}

// GetDatabaseConnection establishes a connection to a PostgreSQL database
// and returns the sql.DB connection object.
func (pg *PostgresDriver) GetDatabaseConnection(host string, port int, user, password, dbname string) (*sql.DB, error) {
	// Build the connection string
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Open a database connection
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	fmt.Println("Successfully connected to PostgreSQL!")
	return db, nil
}
