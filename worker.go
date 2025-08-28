package main

import (
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
)

func getOrCreateJobExecutionState(ap *parser.ArchivalPlan) {
	// Initialize the execution state for the job by either creating a new one (new execution)
	// or retrieving the existing one (resumed execution)
	// Append a new job execution state after acquiring mutex lock

	switch ap.RuntimeParameters.ExecutionMode {
	case EXECUTION_MODE_NEW:
		ap.RuntimeParameters.MutexLocks["JobStateMutexLock"].Lock()
		defer ap.RuntimeParameters.MutexLocks["JobStateMutexLock"].Unlock()
		ap.RuntimeParameters.JobExecutionStates = append(
			ap.RuntimeParameters.JobExecutionStates,
			parser.ExecutionState{
				Initialized:       JOB_STAGE_NOT_STARTED,
				DatabaseConnected: JOB_STAGE_NOT_STARTED,
				QueryExecuted:     JOB_STAGE_NOT_STARTED,
				FileExported:      JOB_STAGE_NOT_STARTED,
				FileCompressed:    JOB_STAGE_NOT_STARTED,
				FileUploaded:      JOB_STAGE_NOT_STARTED,
				CleanupDone:       JOB_STAGE_NOT_STARTED,
				ThreadSuccess:     JOB_STAGE_NOT_STARTED,
			},
		)

	case EXECUTION_MODE_RESUMED:
		// Resumed execution, do nothing as the states are already loaded from the statefile
		fmt.Println("Resuming execution, existing job states loaded")
	default:
		log.Fatal("Invalid execution mode:", ap.RuntimeParameters.ExecutionMode)
	}
}

func updateJobExecutionState(ap *parser.ArchivalPlan, workerID int, stateKey, stateValue string) error {
	// There is no need to lock the mutex here because this function is always called
	// from within launchArchivalWorker, which updates only its own workerID's state.

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
	case "ThreadSuccess":
		ap.RuntimeParameters.JobExecutionStates[workerID].ThreadSuccess = stateValue
	default:
		return fmt.Errorf("invalid state key: %s", stateKey)
	}
	return nil
}

// Function that will carry out the archival process. To be used in a Go-routine.
func launchArchivalWorker(workerID int, ap *parser.ArchivalPlan) error {
	// Initialize worker-specific variables
	threadIndex := workerID
	dryRun := ap.RuntimeParameters.DryRun
	fmt.Println("WorkerID: ", workerID)
	query := ap.RuntimeParameters.Queries[threadIndex]

	workingDir := ap.RuntimeParameters.WorkingDir

	fmt.Println("Executing query: ", query)
	uncompressedFilepath := fmt.Sprintf(
		"%s/output-%s.txt",
		workingDir,
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

	// JobStateMutexLock is used only for createJobExecutionState.
	// Initialize (new execution) or retrieve (resumed execution) the execution state for this worker
	getOrCreateJobExecutionState(ap)

	// Set worker initialized state value
	updateJobExecutionState(ap, workerID, "Initialized", JOB_STAGE_COMPLETED)
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
			updateJobExecutionState(ap, workerID, "DatabaseConnected", JOB_STAGE_FAILED)
			log.Fatal("Failed to connect to DB:", err)
		} else {
			updateJobExecutionState(ap, workerID, "DatabaseConnected", JOB_STAGE_COMPLETED)
		}
		defer db.Close() // Ensure DB connection is closed when done

		// Run the query on the DB connection
		data, columns, err := extract.QueryDynamic(db, query)
		if err != nil {
			updateJobExecutionState(ap, workerID, "QueryExecuted", JOB_STAGE_FAILED)
			log.Fatal("Query failed:", err)
		} else {
			updateJobExecutionState(ap, workerID, "QueryExecuted", JOB_STAGE_COMPLETED)
		}

		// Export to CSV
		err = extract.WriteCSV(uncompressedFilepath, data, columns)
		if err != nil {
			updateJobExecutionState(ap, workerID, "FileExported", JOB_STAGE_FAILED)
			log.Fatal("Failed to write CSV:", err)
		} else {
			updateJobExecutionState(ap, workerID, "FileExported", JOB_STAGE_COMPLETED)
		}
		fmt.Println("CSV Exported Successfully!")

		err = compress.CompressFile(uncompressedFilepath, compressedFilePath)
		if err != nil {
			updateJobExecutionState(ap, workerID, "FileCompressed", JOB_STAGE_FAILED)
			fmt.Println("Compression failed:", err)
		} else {
			fmt.Println("File compressed successfully!")
			updateJobExecutionState(ap, workerID, "FileCompressed", JOB_STAGE_COMPLETED)
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
				updateJobExecutionState(ap, workerID, "FileUploaded", JOB_STAGE_FAILED)
				log.Fatal("Error uploading file", err)
			} else {
				updateJobExecutionState(ap, workerID, "FileUploaded", JOB_STAGE_COMPLETED)
			}

		} else {
			updateJobExecutionState(ap, workerID, "FileUploaded", JOB_STAGE_SKIPPED_BY_USER)
			fmt.Println("Archival storage not enabled, skipping upload. Retaining data files.")
		}

		// Cleanup: delete the uncompressed and compressed files if cleanup is enabled
		// Cleanup is inside this block because we only want to delete if upload was enabled (?)
		if ap.Cleanup.Enabled {
			fmt.Printf("Deleting file: %s \n", uncompressedFilepath)
			err = os.Remove(uncompressedFilepath)
			if err != nil {
				updateJobExecutionState(ap, workerID, "CleanupDone", JOB_STAGE_FAILED)
				fmt.Println("Error deleting uncompressed file:", err)
			}

			fmt.Printf("Deleting file: %s\n", compressedFilePath)
			err = os.Remove(compressedFilePath)
			if err != nil {
				updateJobExecutionState(ap, workerID, "CleanupDone", JOB_STAGE_FAILED)
				fmt.Println("Error deleting compressed file:", err)
			}
			updateJobExecutionState(ap, workerID, "CleanupDone", JOB_STAGE_COMPLETED)

		} else {
			updateJobExecutionState(ap, workerID, "CleanupDone", JOB_STAGE_SKIPPED_BY_USER)
			fmt.Println("Cleanup not enabled, data files retained")
		}
	} else {
		// In dry run mode, skip all steps after initialization
		updateJobExecutionState(ap, workerID, "DatabaseConnected", JOB_STAGE_SKIPPED_ON_DRY_RUN)
		updateJobExecutionState(ap, workerID, "QueryExecuted", JOB_STAGE_SKIPPED_ON_DRY_RUN)
		updateJobExecutionState(ap, workerID, "FileExported", JOB_STAGE_SKIPPED_ON_DRY_RUN)
		updateJobExecutionState(ap, workerID, "FileCompressed", JOB_STAGE_SKIPPED_ON_DRY_RUN)
		updateJobExecutionState(ap, workerID, "FileUploaded", JOB_STAGE_SKIPPED_ON_DRY_RUN)
		updateJobExecutionState(ap, workerID, "CleanupDone", JOB_STAGE_SKIPPED_ON_DRY_RUN)
		fmt.Println("Dry run enabled, skipping database operations and file handling.")
	}

	updateJobExecutionState(ap, workerID, "ThreadSuccess", JOB_STAGE_COMPLETED)
	return nil
}
