package database

import (
	"database/sql"
)

type DatabaseDriver interface {
    GetDatabaseConnection(host string, port int, user, password, dbname string) (*sql.DB, error)
}