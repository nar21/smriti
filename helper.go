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
	"sync"
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

func createJobExecutionState(ap *parser.ArchivalPlan, JobStateMutexLock *sync.Mutex) {
	// Initialize the execution state for this worker
	// Append a new job execution state after acquiring mutex lock
	JobStateMutexLock.Lock()
	defer JobStateMutexLock.Unlock()
	ap.RuntimeParameters.JobExecutionStates = append(
		ap.RuntimeParameters.JobExecutionStates,
		parser.ExecutionState{},
	)
}

func updateJobExecutionState(ap *parser.ArchivalPlan, workerID int, stateKey string, stateValue bool, JobStateMutexLock *sync.Mutex) error {
	// Acquire mutex lock
	JobStateMutexLock.Lock()
	defer JobStateMutexLock.Unlock()

	// Update the specific state field based on the stateKey
	switch stateKey {
	case "Initialized":
		ap.RuntimeParameters.JobExecutionStates[workerID].Initialized = stateValue
	case "DatabaseConnected":
		ap.RuntimeParameters.JobExecutionStates[workerID].DatabaseConnected = stateValue
	case "QueryExecuted":
		ap.RuntimeParameters.JobExecutionStates[workerID].QueryExecuted = stateValue
	case "FileExported":
		ap.RuntimeParameters.JobExecutionStates[workerID].FileExported = stateValue
	case "FileCompressed":
		ap.RuntimeParameters.JobExecutionStates[workerID].FileCompressed = stateValue
	case "FileUploaded":
		ap.RuntimeParameters.JobExecutionStates[workerID].FileUploaded = stateValue
	case "CleanupDone":
		ap.RuntimeParameters.JobExecutionStates[workerID].CleanupDone = stateValue
	default:
		return fmt.Errorf("invalid state key: %s", stateKey)
	}
	return nil
}

// Function that will carry out the archival process. To be used in a Go-routine.
func launchArchivalWorker(workerID int, ap *parser.ArchivalPlan, JobStateMutexLock *sync.Mutex) {
	// Initialize worker-specific variables
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

	// Initialize the execution state for this worker
	ap.RuntimeParameters.JobExecutionStates = append(
		ap.RuntimeParameters.JobExecutionStates,
		parser.ExecutionState{},
	)

	// Set worker initialized state value
	updateJobExecutionState(ap, workerID, "Initialized", true, JobStateMutexLock)
	// COMPLETED: initialization of worker-specific variables

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
		} else {
			updateJobExecutionState(ap, workerID, "DatabaseConnected", true, JobStateMutexLock)
		}
		defer db.Close() // Ensure DB connection is closed when done

		// Run the query on the DB connection
		data, columns, err := extract.QueryDynamic(db, query)
		if err != nil {
			log.Fatal("Query failed:", err)
		} else {
			updateJobExecutionState(ap, workerID, "QueryExecuted", true, JobStateMutexLock)
		}

		// Export to CSV
		err = extract.WriteCSV(uncompressedFilepath, data, columns)
		if err != nil {
			log.Fatal("Failed to write CSV:", err)
		} else {
			updateJobExecutionState(ap, workerID, "FileExported", true, JobStateMutexLock)
		}
		fmt.Println("CSV Exported Successfully!")

		err = compress.CompressFile(uncompressedFilepath, compressedFilePath)
		if err != nil {
			fmt.Println("Compression failed:", err)
		} else {
			fmt.Println("File compressed successfully!")
			updateJobExecutionState(ap, workerID, "FileCompressed", true, JobStateMutexLock)
		}

		// If archival storage is enabled, upload the compressed file
		if ap.ArchiveStorage.Enabled {
			// Upload compressed file to object storage
			fmt.Println("Uploading file to object storage...")

			// Initialize the object storage driver based on the archival plan
			var driver objectStorage.ObjectStorageDriver
			storageDriver := ap.ArchiveStorage.Type
			if storageDriver == "" {
				log.Fatal("No storage type specified in archival plan")
			}

			// Create the appropriate storage driver based on the type specified
			switch storageDriver {
			case "s3":
				s3Driver, err := objectStorage.NewS3Driver(
					ap.ArchiveStorage.S3.Bucket,
				)
				if err != nil {
					fmt.Println("Could not create S3 driver:", err)
				}
				driver = s3Driver
			case "local":
				localFSDriver, err := objectStorage.NewLocalFSDriver(
					ap.ArchiveStorage.Local.Path,
				)
				if err != nil {
					fmt.Println("Could not create LocalFS driver:", err)
				}
				driver = localFSDriver
			default:
				log.Fatal("Unsupported storage type:", storageDriver)
			}

			// Upload the compressed file to the configured storage
			err = driver.Upload(compressedFilePath, archiveFilePath)
			if err != nil {
				log.Fatal("Error uploading file", err)
			} else {
				updateJobExecutionState(ap, workerID, "FileUploaded", true, JobStateMutexLock)
			}

			// Cleanup: delete the uncompressed and compressed files if cleanup is enabled
			// Cleanup is inside this block because we only want to delete if upload was enabled (?)
			if ap.Cleanup.Enabled {
				fmt.Printf("Deleting file: %s \n", uncompressedFilepath)
				err = os.Remove(uncompressedFilepath)
				if err != nil {
					fmt.Println("Error deleting uncompressed file:", err)
				}

				fmt.Printf("Deleting file: %s\n", compressedFilePath)
				err = os.Remove(compressedFilePath)
				if err != nil {
					fmt.Println("Error deleting compressed file:", err)
				}
				updateJobExecutionState(ap, workerID, "CleanupDone", true, JobStateMutexLock)

			} else {
				fmt.Println("Cleanup not enabled, data files retained")
			}

		} else {
			fmt.Println("Archival storage not enabled, skipping upload. Retaining data files.")
		}

		if err = ap.SaveExecutionState(); err != nil {
			log.Fatal("Failed to save execution state:", err)
		}
	}

}

// CreateDirIfNotExist checks if a directory exists, and creates it if it does not.
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
