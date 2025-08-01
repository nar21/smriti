package main

import (
	"db-archive/parser"
	"db-archive/sqlgen"
	"fmt"
    "log"
	"os"
	"flag"
)

// Declare global variables
var workDir string

func main() {
    // Parse command line flags
    dryRun := flag.Bool("dry-run", false, "Dry run without executing queries or creating files")
	flag.Parse()

    // Create working directory if it does not exist
    workDir = "./workDir"
    CreateDirIfNotExist(workDir)

    archivalPlanName := os.Getenv("ARCHIVAL_PLAN")
    if archivalPlanName == "" {
        log.Fatal("Could not find ARCHIVAL_PLAN env variable")
    }
    archivalPlanFilepath := fmt.Sprintf("archival-plan/%s.yaml", archivalPlanName)
    fmt.Println("Loading archival plan at ", archivalPlanFilepath)

    // Load the archival plan object
    var ap, err2 = parser.LoadArchivalPlan(archivalPlanFilepath)
	if err2 != nil {
		log.Fatal("Could not load archival plan")
	}

    // Overwrite archival plan with CLI values
    ap.RuntimeParameters.DryRun = *dryRun
    executionID, err := GenerateExecutionID()
    if err != nil {
        log.Fatal("Could not generate Execution ID", err)
    }
    ap.RuntimeParameters.ExecutionID = executionID
    fmt.Println("ExecutionID: ", ap.RuntimeParameters.ExecutionID)

    // Call the appropriate DB plugin to generate SQL queries
    var sqlGenerator sqlgen.SQLGeneratorDriver
    sqlGenName := "postgresql"

    switch sqlGenName {
        case "postgresql":
            sqlGenerator = &sqlgen.PostgresqlSQLGenerator{}
    }

    // Generic function call to the interface
    ap.RuntimeParameters.Queries = sqlGenerator.GenerateSQL(ap)

    // Execute all the queries through workers
    for i := 0; i < len(ap.RuntimeParameters.Queries); i++ {

        launchArchivalWorker(i, ap)
    }

}
