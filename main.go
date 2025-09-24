package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"smriti/config"
	"smriti/logging"
	"smriti/parser"
	"smriti/sqlgen"
	"sync"
	"time"
)

func main() {
	// Parse command line flags
	dryRun := flag.Bool("dry-run", false, "Dry run without executing queries or creating files")
	init := flag.Bool("init", false, "Create settings.yaml with default values and exit")
	executionIDArg := flag.String("execution-id", "", "Execution ID that has to be resumed")
	flag.Parse()

	if init != nil && *init {
		err := initSmritiDirsAndFiles()
		if err != nil {
			log.Fatal("Could not initialize smriti directories and files:", err)
		}
		fmt.Println("Initialized smriti directories and files. Edit settings.yaml and db-credentials.yaml as needed, then run again without --init flag.")
		os.Exit(0)
	}

	globalSettings, err := config.LoadGlobalSettings()
	if err != nil {
		log.Fatal("Could not load global settings. Try running with --init flag to create a default settings.yaml file")
	}

	// Create working directory if it does not exist
	workDirBasePath, err := filepath.Abs(globalSettings.WorkingDir)
	if err != nil {
		log.Fatal("Could not get absolute path of working directory")
	}
	fmt.Println("Working Directory: ", workDirBasePath)
	CreateDirIfNotExist(workDirBasePath)

	// Initialize static variables
	executionID := *executionIDArg
	statefileName := "statefile.yaml"

	// Declare dynamic variables
	var executionMode string
	var archivalPlanFilepath string

	if executionID == "" {
		var err error
		executionID, err = GenerateExecutionID()
		if err != nil {
			log.Fatal("Could not generate Execution ID", err)
		}
		executionMode = "new"
	} else {
		executionMode = "resumed"
	}

	// Generate the working directory for the current execution
	executionDir := fmt.Sprintf(
		"%s/%s",
		path.Join(workDirBasePath),
		executionID,
	)
	fmt.Println("Execution Directory: ", executionDir)

	// Set the archival plans directory path
	archivalPlansDir, err := filepath.Abs(globalSettings.ArchivalPlansDir)
	if err != nil {
		log.Fatal("Could not get absolute path of archival plans directory")
	}
	fmt.Println("Archival Plans Directory: ", archivalPlansDir)
	// Get the archival plan name from env variable
	archivalPlanName := os.Getenv("ARCHIVAL_PLAN")
	if archivalPlanName == "" {
		log.Fatal("Could not find ARCHIVAL_PLAN env variable")
	}

	stateFilePath := fmt.Sprintf(
		"%s/%s",
		executionDir,
		statefileName,
	)

	switch executionMode {
	case EXECUTION_MODE_NEW:
		archivalPlanName := os.Getenv("ARCHIVAL_PLAN")
		if archivalPlanName == "" {
			log.Fatal("Could not find ARCHIVAL_PLAN env variable")
		}
		archivalPlanFilepath = fmt.Sprintf("%s/%s.yaml", archivalPlansDir, archivalPlanName)
	case EXECUTION_MODE_RESUMED:
		// Load the archival plan from the saved state file
		archivalPlanFilepath = stateFilePath
	}

	fmt.Println("Loading archival plan at ", archivalPlanFilepath)

	// Load the archival plan object
	var ap, err2 = parser.LoadArchivalPlan(archivalPlanFilepath, *globalSettings)
	if err2 != nil {
		log.Fatal("Could not load archival plan")
	}

	// Set runtime values with CLI values
	ap.RuntimeParameters.DryRun = *dryRun
	ap.RuntimeParameters.ExecutionMode = executionMode
	ap.RuntimeParameters.ExecutionID = executionID
	ap.RuntimeParameters.WorkingDir = executionDir

	// Create the working directory if it does not exist
	CreateDirIfNotExist(executionDir)

	// Initialize logger after basic initilization (like creation of working directory) is done
	logger, err := logging.NewJobLogger("main", ap)
	if err != nil {
		log.Fatal("Could not create main logger: ", err)
	}
	logger.Log("Starting execution with ExecutionID: ", ap.RuntimeParameters.ExecutionID)

	// Call the appropriate DB plugin to generate SQL queries
	var sqlGenerator sqlgen.SQLGeneratorDriver
	sqlGenName := "postgresql"

	switch sqlGenName {
	case "postgresql":
		sqlGenerator = &sqlgen.PostgresqlSQLGenerator{}
	}

	// Generic function call to the interface
	ap.RuntimeParameters.Queries = sqlGenerator.GenerateSQL(ap)

	//Create a global mutex for synchronizing access to shared resources
	ap.RuntimeParameters.MutexLocks = make(map[string]*sync.Mutex)
	ap.RuntimeParameters.MutexLocks["JobStateMutexLock"] = &sync.Mutex{}

	// Execute all the queries using workers
	for i := 0; i < len(ap.RuntimeParameters.Queries); i++ {
		if err := launchArchivalWorker(i, ap); err != nil {
			fmt.Printf("Archival worker %d failed: %s \n", i, err)
		}

		if i == 2 {
			fmt.Println("Sleeping for 10 seconds")
			time.Sleep(10) // * time.Second)
		}
	}

	// Save the final execution state
	if err := ap.SaveExecutionState(stateFilePath); err != nil {
		log.Fatal("Failed to save execution state:", err)
	}

}
