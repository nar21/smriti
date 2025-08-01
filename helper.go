package main

import (
    "crypto/rand"
    "fmt"
    "time"
    "log"
	"db-archive/extract"
	"db-archive/compress"
	"db-archive/objectStorage"
    "db-archive/parser"
	"db-archive/database"
	"os"
	"strconv"
)

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

// Function that will carry out the archival process. To be used in a Go-routine.
//func launchArchivalWorker(query string, db *sql.DB, workerID string, executionID string) {
func launchArchivalWorker(workerID int, ap *parser.ArchivalPlan) {
    threadIndex := workerID
    dryRun := ap.RuntimeParameters.DryRun

    fmt.Println("WorkerID: ", workerID)

    query := ap.RuntimeParameters.Queries[threadIndex]

    threadWorkingDir := fmt.Sprintf("%s/%s/%s", workDir, ap.RuntimeParameters.ExecutionID, strconv.Itoa(workerID))
    fmt.Println("Execution Directory: ", threadWorkingDir)
    CreateDirIfNotExist(threadWorkingDir)

    fmt.Println("Executing query: ", query)
    uncompressedFilepath := fmt.Sprintf("%s/%s", threadWorkingDir, "output.txt")
    compressedFilePath := uncompressedFilepath + ".gz"
    fmt.Println("Plaintext filepath: ", uncompressedFilepath)
    fmt.Println("Compressed filepath: ", compressedFilePath)

    if !dryRun {
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

        // Run the query on the DB connection
        data, columns, err := extract.QueryDynamic(db, query)
        if err != nil {
            log.Fatal("Query failed:", err)
        }

        // Export to CSV
        err = extract.WriteCSV(uncompressedFilepath, data, columns)
        if err != nil {
            log.Fatal("Failed to write CSV:", err)
        }
        fmt.Println("CSV Exported Successfully!")

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