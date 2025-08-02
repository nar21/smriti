package main

import (
	"crypto/rand"
	"db-archive/compress"
	"db-archive/database"
	"db-archive/extract"
	"db-archive/objectStorage"
	"db-archive/parser"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"time"
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
func launchArchivalWorker(workerID int, ap *parser.ArchivalPlan) {
	threadIndex := workerID
	dryRun := ap.RuntimeParameters.DryRun
	fmt.Println("WorkerID: ", workerID)
	query := ap.RuntimeParameters.Queries[threadIndex]

	threadWorkingDir := fmt.Sprintf(
		"%s/%s",
		path.Join(workDir),
		ap.RuntimeParameters.ExecutionID,
	)
	fmt.Println("Execution Directory: ", threadWorkingDir)
	CreateDirIfNotExist(threadWorkingDir)

	fmt.Println("Executing query: ", query)
	uncompressedFilepath := fmt.Sprintf(
		"%s/output-%s.txt",
		threadWorkingDir,
		strconv.Itoa(workerID),
	)

	compressedFilePath := uncompressedFilepath + ".gz"
	fmt.Println("Plaintext filepath: ", uncompressedFilepath)
	fmt.Println("Compressed filepath: ", compressedFilePath)

	archiveFilePath := fmt.Sprintf(
		"uploads/%s/%s/%s/%s/%s",
		ap.DatabaseID,
		ap.DatabaseCredential.DBName,
		ap.Query.Table,
		ap.RuntimeParameters.ExecutionID,
		path.Base(compressedFilePath),
	)
	fmt.Println("Archive File Path: ", archiveFilePath)

	if !dryRun {
		// Database connection parameters
		host := ap.DatabaseCredential.Host
		port := ap.DatabaseCredential.Port
		user := ap.DatabaseCredential.User
		password := ap.DatabaseCredential.Password
		dbname := ap.DatabaseCredential.DBName
		dbEngine := ap.DatabaseCredential.Engine

		fmt.Printf("Connecting to %s database \"%s\" at %s:%d\n", dbEngine, dbname, host, port)
		var dbDriver database.DatabaseDriver

		switch dbEngine {
		case "postgres":
			dbDriver = &database.PostgresDriver{}
		default:
			log.Fatal("Unsupported database engine:", dbEngine)
		}

		// Establish a connection to the database
		db, err := dbDriver.GetDatabaseConnection(host, port, user, password, dbname)
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

		// Upload compressed file to object storage
		fmt.Println("Uploading file to object storage...")

		// Initialize the object storage driver based on the archival plan
		var driver objectStorage.ObjectStorageDriver
		storageDriver := "s3"

		switch storageDriver {
		case "s3":
			s3Driver, err := objectStorage.NewS3Driver(
				ap.ArchiveStorage.Bucket.BucketName,
			)
			if err != nil {
				fmt.Println("Could not create S3 driver:", err)
			}
			driver = s3Driver
		case "azureblob":
			driver = &objectStorage.AzureBlobDriver{}
		}

		err = driver.Upload(compressedFilePath, archiveFilePath)
		if err != nil {
			log.Fatal("Error uploading file", err)
		}

		if ap.Cleanup.Enabled {
			fmt.Printf("Deleting files: %s and %s\n", uncompressedFilepath, compressedFilePath)
			os.Remove(uncompressedFilepath)
			os.Remove(compressedFilePath)
		} else {
			fmt.Println("Cleanup not enabled, files retained.")
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
