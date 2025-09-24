package main

import (
	"fmt"
	"os"
	"path"
	"smriti/compress"
	"smriti/database"
	"smriti/extract"
	"smriti/logging"
	"smriti/objectStorage"
	"smriti/parser"
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
		// Resumed execution, initialization is not required as the states are already loaded from the statefile
		fmt.Println("Resuming execution, existing job states loaded")
	}
}

func updateJobStageState(ap *parser.ArchivalPlan, jobID int, stateKey string, stateValue string) {
	// There is no need to lock the mutex here because this function is always called
	// from within launchArchivalWorker, which updates only a specific job's state.

	// Update the specific state field based on the stateKey
	switch stateKey {
	case "Initialized":
		ap.RuntimeParameters.JobExecutionStates[jobID].Initialized = stateValue
	case "DatabaseConnected":
		ap.RuntimeParameters.JobExecutionStates[jobID].DatabaseConnected = stateValue
	case "QueryExecuted":
		ap.RuntimeParameters.JobExecutionStates[jobID].QueryExecuted = stateValue
	case "FileExported":
		ap.RuntimeParameters.JobExecutionStates[jobID].FileExported = stateValue
	case "FileCompressed":
		ap.RuntimeParameters.JobExecutionStates[jobID].FileCompressed = stateValue
	case "FileUploaded":
		ap.RuntimeParameters.JobExecutionStates[jobID].FileUploaded = stateValue
	case "CleanupDone":
		ap.RuntimeParameters.JobExecutionStates[jobID].CleanupDone = stateValue
	case "ThreadSuccess":
		ap.RuntimeParameters.JobExecutionStates[jobID].ThreadSuccess = stateValue
	}
}

func finalizeJobExecution(ap *parser.ArchivalPlan, jobID int) error {
	//
	jobState := ap.RuntimeParameters.JobExecutionStates[jobID]

	if jobState.Initialized == JOB_STAGE_COMPLETED &&
		jobState.DatabaseConnected == JOB_STAGE_COMPLETED &&
		jobState.QueryExecuted == JOB_STAGE_COMPLETED &&
		jobState.FileExported == JOB_STAGE_COMPLETED &&
		(jobState.FileCompressed == JOB_STAGE_COMPLETED || jobState.FileCompressed == JOB_STAGE_SKIPPED_BY_USER) &&
		(jobState.FileUploaded == JOB_STAGE_COMPLETED || jobState.FileUploaded == JOB_STAGE_SKIPPED_BY_USER) &&
		(jobState.CleanupDone == JOB_STAGE_COMPLETED || jobState.CleanupDone == JOB_STAGE_SKIPPED_BY_USER) {

		// Mark the thread as successful only if all mandatory stages are completed or optional fields are opted out
		updateJobStageState(ap, jobID, "ThreadSuccess", JOB_STAGE_COMPLETED)
	} else {
		updateJobStageState(ap, jobID, "ThreadSuccess", JOB_STAGE_FAILED)
	}
	return nil
}

func getJobStageState(ap *parser.ArchivalPlan, jobID int, stateKey string) string {
	var stateValue string
	switch stateKey {
	case "Initialized":
		stateValue = ap.RuntimeParameters.JobExecutionStates[jobID].Initialized
	case "DatabaseConnected":
		stateValue = ap.RuntimeParameters.JobExecutionStates[jobID].DatabaseConnected
	case "QueryExecuted":
		stateValue = ap.RuntimeParameters.JobExecutionStates[jobID].QueryExecuted
	case "FileExported":
		stateValue = ap.RuntimeParameters.JobExecutionStates[jobID].FileExported
	case "FileCompressed":
		stateValue = ap.RuntimeParameters.JobExecutionStates[jobID].FileCompressed
	case "FileUploaded":
		stateValue = ap.RuntimeParameters.JobExecutionStates[jobID].FileUploaded
	case "CleanupDone":
		stateValue = ap.RuntimeParameters.JobExecutionStates[jobID].CleanupDone
	case "ThreadSuccess":
		stateValue = ap.RuntimeParameters.JobExecutionStates[jobID].ThreadSuccess
	}
	return stateValue
}

// Function that will carry out the archival process. To be used in a Go-routine.
func launchArchivalWorker(jobID int, ap *parser.ArchivalPlan) error {
	// Initialize worker-specific variables
	threadIndex := jobID
	dryRun := ap.RuntimeParameters.DryRun
	fmt.Println("JobID: ", jobID)
	query := ap.RuntimeParameters.Queries[threadIndex]
	workingDir := ap.RuntimeParameters.WorkingDir

	
	logger, err := logging.NewJobLogger(strconv.Itoa(jobID), ap)
	if err != nil {
		return fmt.Errorf("could not create job logger: %s", err)
	}
	defer logger.Close()

	logger.Log("Executing query:", query)
	uncompressedFilepath := fmt.Sprintf(
		"%s/output-%s.txt",
		workingDir,
		strconv.Itoa(jobID),
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
	updateJobStageState(ap, jobID, "Initialized", JOB_STAGE_COMPLETED)
	// COMPLETED: initialization of worker-specific variables

	if !dryRun {
		// Execute a stage only if not already completed successfully
		if getJobStageState(ap, jobID, "ThreadSuccess") != JOB_STAGE_COMPLETED {
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
				return fmt.Errorf("unsupported database engine: %s", dbEngine)
			}

			// Establish a connection to the database in both the "new" and "resumed" execution modes
			db, err := dbDriver.GetDatabaseConnection(host, port, user, password, dbname)
			if err != nil {
				updateJobStageState(ap, jobID, "DatabaseConnected", JOB_STAGE_FAILED)
				return fmt.Errorf("failed to connect to DB: %s", err)
			} else {
				updateJobStageState(ap, jobID, "DatabaseConnected", JOB_STAGE_COMPLETED)
			}
			defer db.Close() // Ensure DB connection is closed when done

			// Run the query on the DB connection
			var data []map[string]interface{}
			var columns []string

			// Execute the query only if it hasn't been executed successfully before
			if getJobStageState(ap, jobID, "QueryExecuted") != JOB_STAGE_COMPLETED {
				data, columns, err = extract.QueryDynamic(db, query)
				if err != nil {
					updateJobStageState(ap, jobID, "QueryExecuted", JOB_STAGE_FAILED)
					return fmt.Errorf("query failed: %s", err)
				} else {
					updateJobStageState(ap, jobID, "QueryExecuted", JOB_STAGE_COMPLETED)
				}
			}

			// Export to CSV if not already done
			if getJobStageState(ap, jobID, "FileExported") != JOB_STAGE_COMPLETED {
				err = extract.WriteCSV(uncompressedFilepath, data, columns)
				if err != nil {
					updateJobStageState(ap, jobID, "FileExported", JOB_STAGE_FAILED)
					return fmt.Errorf("failed to write CSV: %s", err)
				} else {
					updateJobStageState(ap, jobID, "FileExported", JOB_STAGE_COMPLETED)
				}
				fmt.Println("CSV Exported Successfully!")
			}

			// Compress the exported file if not already done
			if getJobStageState(ap, jobID, "FileCompressed") != JOB_STAGE_COMPLETED {
				err = compress.CompressFile(uncompressedFilepath, compressedFilePath)
				if err != nil {
					updateJobStageState(ap, jobID, "FileCompressed", JOB_STAGE_FAILED)
					return fmt.Errorf("compression failed: %s", err)
				} else {
					fmt.Println("File compressed successfully!")
					updateJobStageState(ap, jobID, "FileCompressed", JOB_STAGE_COMPLETED)
				}
			}

			// If archival storage is enabled, upload the compressed file if not already done
			if getJobStageState(ap, jobID, "FileUploaded") != JOB_STAGE_COMPLETED {
				if ap.ArchiveStorage.Enabled {
					// Upload compressed file to object storage
					fmt.Println("Uploading file to object storage...")

					// Initialize the object storage driver based on the archival plan
					var driver objectStorage.ObjectStorageDriver
					storageDriver := ap.ArchiveStorage.Type
					if storageDriver == "" {
						return fmt.Errorf("no archival storage type specified in archival plan")
					}

					// Create the appropriate storage driver based on the type specified
					switch storageDriver {
					case "s3":
						s3Driver, err := objectStorage.NewS3Driver(
							ap.ArchiveStorage.S3.Bucket,
						)
						if err != nil {
							return fmt.Errorf("could not create S3 driver: %s", err)
						}
						driver = s3Driver
					case "local":
						localFSDriver, err := objectStorage.NewLocalFSDriver(
							ap.ArchiveStorage.Local.Path,
						)
						if err != nil {
							return fmt.Errorf("could not create LocalFS driver: %s", err)
						}
						driver = localFSDriver
					default:
						return fmt.Errorf("unsupported storage type: %s", storageDriver)
					}

					// Upload the compressed file to the configured storage
					err = driver.Upload(compressedFilePath, archiveFilePath)
					if err != nil {
						updateJobStageState(ap, jobID, "FileUploaded", JOB_STAGE_FAILED)
						return fmt.Errorf("error uploading file: %s", err)
					} else {
						updateJobStageState(ap, jobID, "FileUploaded", JOB_STAGE_COMPLETED)
					}

				} else {
					updateJobStageState(ap, jobID, "FileUploaded", JOB_STAGE_SKIPPED_BY_USER)
					fmt.Println("Archival storage not enabled, skipping upload.")
				}
			}

			// Cleanup: delete the uncompressed and compressed files if cleanup is enabled
			// Cleanup if enabled can delete files even if upload failed. Can make troubleshooting harder.
			// Run cleanup only if not already done
			if getJobStageState(ap, jobID, "CleanupDone") != JOB_STAGE_COMPLETED {
				if ap.Cleanup.Enabled {
					fmt.Printf("Deleting file: %s \n", uncompressedFilepath)
					err = os.Remove(uncompressedFilepath)
					if err != nil {
						updateJobStageState(ap, jobID, "CleanupDone", JOB_STAGE_FAILED)
						return fmt.Errorf("error deleting uncompressed file:", err)
					}

					fmt.Printf("Deleting file: %s\n", compressedFilePath)
					err = os.Remove(compressedFilePath)
					if err != nil {
						updateJobStageState(ap, jobID, "CleanupDone", JOB_STAGE_FAILED)
						return fmt.Errorf("error deleting compressed file: %s", err)
					}
					updateJobStageState(ap, jobID, "CleanupDone", JOB_STAGE_COMPLETED)

				} else {
					updateJobStageState(ap, jobID, "CleanupDone", JOB_STAGE_SKIPPED_BY_USER)
					fmt.Println("Cleanup not enabled, data files retained")
				}
			}
		} else {
			fmt.Printf("Thread %d already completed successfully, skipping all stages.\n", jobID)
		}
	} else {
		// In dry run mode, skip all steps after initialization
		updateJobStageState(ap, jobID, "DatabaseConnected", JOB_STAGE_SKIPPED_ON_DRY_RUN)
		updateJobStageState(ap, jobID, "QueryExecuted", JOB_STAGE_SKIPPED_ON_DRY_RUN)
		updateJobStageState(ap, jobID, "FileExported", JOB_STAGE_SKIPPED_ON_DRY_RUN)
		updateJobStageState(ap, jobID, "FileCompressed", JOB_STAGE_SKIPPED_ON_DRY_RUN)
		updateJobStageState(ap, jobID, "FileUploaded", JOB_STAGE_SKIPPED_ON_DRY_RUN)
		updateJobStageState(ap, jobID, "CleanupDone", JOB_STAGE_SKIPPED_ON_DRY_RUN)
		fmt.Println("Dry run enabled, skipping database operations and file handling.")
	}

	// Confirm and update the overall job completion status
	finalizeJobExecution(ap, jobID)

	return nil
}
