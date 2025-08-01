package main

import (
	"db-archive/database"
	"db-archive/extract"
	"db-archive/parser"
	"db-archive/sqlgen"
	"db-archive/compress"
	"db-archive/objectStorage"
	"database/sql"
	"fmt"
	"log"
	"crypto/rand"
	"os"
	"time"
	"strconv"
)

// Declare global variables
var workDir string

// Function that will carry out the archival process. To be used in a Go-routine.
func launchArchivalWorker(query string, db *sql.DB, workerID string, executionID string) {
    fmt.Println("WorkerID: ", workerID)

    threadWorkingDir := fmt.Sprintf("%s/%s/%s", workDir, executionID, workerID)
    fmt.Println("Execution Directory: ", threadWorkingDir)
    CreateDirIfNotExist(threadWorkingDir)

    fmt.Println("Executing query: ", query)

    data, columns, err := extract.QueryDynamic(db, query)
	if err != nil {
		log.Fatal("Query failed:", err)
	}

	// Export to CSV
	uncompressedFilepath := fmt.Sprintf("%s/%s", threadWorkingDir, "output.txt")
	err = extract.WriteCSV(uncompressedFilepath, data, columns)
	if err != nil {
		log.Fatal("Failed to write CSV:", err)
	}

	fmt.Println("CSV Exported Successfully!")

    compressedFilePath := uncompressedFilepath + ".gz"
	err = compress.CompressFile(uncompressedFilepath, compressedFilePath)
	if err != nil {
        fmt.Println("Compression failed:", err)
    } else {
        fmt.Println("File compressed successfully!")
    }

    var driver objectStorage.ObjectStorageDriver
    storageDriver := "azureblob"

    switch storageDriver {
        case "s3":
            driver = &objectStorage.S3Driver{}
        case "azureblob":
            driver = &objectStorage.AzureBlobDriver{}
    }

    err = driver.Upload(compressedFilePath)
    if err != nil {
        log.Fatal("Error uploading file: %v\n", err)

    }
}

func CreateDirIfNotExist(dir string) error {
    // Check if directory exists
    if _, err := os.Stat(dir); os.IsNotExist(err) {
        // Directory does not exist, create it (including parents)
        err := os.MkdirAll(dir, 0755) // 0755 is typical permission
        if err != nil {
            return err
        }
        fmt.Println("Directory created:", dir)
    } else {
        fmt.Println("Directory already exists:", dir)
    }
    return nil
}

// GenerateDateHexString returns the current date in YYYYMMDD format appended with a random 6-hex-character string.
func GenerateExecutionID() (string, error) {
    // Get current date in YYYYMMDD format
    dateStr := time.Now().Format("2006-01-02")

    // Generate 3 random bytes (6 hex characters)
    randomBytes := make([]byte, 3)
    _, err := rand.Read(randomBytes)
    if err != nil {
        return "", err
    }

    // Format bytes as hex string
    randomHex := fmt.Sprintf("%06x", randomBytes)

    return fmt.Sprintf("%s-%s", dateStr, randomHex), nil
}

func main() {
    workDir = "./workDir"
    // Create working directory if it does not exist
    CreateDirIfNotExist(workDir)

    archivalPlanName := os.Getenv("ARCHIVAL_PLAN")
    if archivalPlanName == "" {
        log.Fatal("Could not find ARCHIVAL_PLAN env variable")
    }
    archivalPlanFilepath := fmt.Sprintf("archival-plan/%s.yaml", archivalPlanName)
    fmt.Println("Loading archival plan at ", archivalPlanFilepath)
    var ap, err2 = parser.LoadArchivalPlan(archivalPlanFilepath)

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

    // Call the appropriate DB plugin to generate SQL queries
    var sqlGenerator sqlgen.SQLGeneratorDriver
    sqlGenName := "postgresql"

    switch sqlGenName {
        case "postgresql":
            sqlGenerator = &sqlgen.PostgresqlSQLGenerator{}
    }

    // Generic function call to the interface
    queryList := sqlGenerator.GenerateSQL(ap)
    executionID, err := GenerateExecutionID()
    if err != nil {
		log.Fatal("Could not generate Execution ID", err)
	}
    fmt.Println("ExecutionID: ", executionID)

    // Execute all the queries through workers
    for i := 0; i < len(queryList); i ++ {
        query := queryList[i]
        launchArchivalWorker(query, db, strconv.Itoa(i), executionID)
    }
}
