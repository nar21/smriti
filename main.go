package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"smriti/parser"
	"smriti/sqlgen"
	"sync"
	"time"
)

// Declare global variables
var workDirBasePath string

func main() {
	// Parse command line flags
	dryRun := flag.Bool("dry-run", false, "Dry run without executing queries or creating files")
	executionIDArg := flag.String("execution-id", "", "Execution ID that has to be resumed")
	flag.Parse()

	// Create working directory if it does not exist
	workDirBasePath = "./workDir"
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
	workingDir := fmt.Sprintf(
		"%s/%s",
		path.Join(workDirBasePath),
		executionID,
	)
	fmt.Println("Execution Directory: ", workingDir)

	archivalPlanName := os.Getenv("ARCHIVAL_PLAN")
	if archivalPlanName == "" {
		log.Fatal("Could not find ARCHIVAL_PLAN env variable")
	}

	stateFilePath := fmt.Sprintf(
		"%s/%s",
		workingDir,
		statefileName,
	)

	switch executionMode {
	case EXECUTION_MODE_NEW:
		archivalPlanName := os.Getenv("ARCHIVAL_PLAN")
		if archivalPlanName == "" {
			log.Fatal("Could not find ARCHIVAL_PLAN env variable")
		}
		archivalPlanFilepath = fmt.Sprintf("archival-plan/%s.yaml", archivalPlanName)
	case EXECUTION_MODE_RESUMED:
		// Load the archival plan from the saved state file
		archivalPlanFilepath = stateFilePath
	}

	fmt.Println("Loading archival plan at ", archivalPlanFilepath)

	// Load the archival plan object
	var ap, err2 = parser.LoadArchivalPlan(archivalPlanFilepath)
	if err2 != nil {
		log.Fatal("Could not load archival plan")
	}

	// Overwrite archival plan with CLI values
	ap.RuntimeParameters.DryRun = *dryRun
	ap.RuntimeParameters.ExecutionMode = executionMode
	ap.RuntimeParameters.ExecutionID = executionID
	fmt.Println("ExecutionID: ", ap.RuntimeParameters.ExecutionID)
	ap.RuntimeParameters.WorkingDir = workingDir

	// Create the working directory if it does not exist
	CreateDirIfNotExist(workingDir)

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
