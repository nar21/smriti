package parser

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"db-archive/objectStorage"
	"gopkg.in/yaml.v3"
)

type ArchivalPlanQuery struct {
	Table            string   `yaml:"table"`
	BatchingEnabled  bool     `yaml:"batchingEnabled"`
	BatchColumn      string   `yaml:"batchColumn"`
	BatchColumnType  string   `yaml:"batchColumnType"`  // "int" or "date"
	BatchStep        int      `yaml:"batchStep"`        // number if FilterColumnType is int, days if date
	BatchColumnMin   string   `yaml:"batchColumnMin"`   // Parse this variable according to FilterColumnType
	BatchColumnMax   string   `yaml:"batchColumnMax"`   // Parse this variable according to FilterColumnType
	FilterConditions []string `yaml:"filterConditions"` // can this be replaced with a struct, key/operator/value?
}
type DBCredential struct {
	Engine   string `yaml:"engine"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

type CredentialFile struct {
	Databases map[string]DBCredential
}

type QueryExecutionState struct {
	QueryString string
	Executed    bool
}

type ExecutionState struct {
	// TODO: Some of these stages can be skipped by the user, e.g., compression, upload, cleanup,
	// or during dry run
	// A bool may not suffice to represent the other states: not started, completed, skipped_by_user, skipped_due_to_dry_run, skipped_already_completed
	// Currently, if a stage is not run, the flag remains false.

	Initialized       string
	DatabaseConnected string
	QueryExecuted     string
	FileExported      string
	FileCompressed    string
	FileUploaded      string
	CleanupDone       string
	ThreadSuccess     string
}

type RuntimeParams struct {
	DryRun              bool
	ExecutionMode       string // "new" or "resume"
	Workers             int
	ExecutionID         string
	DatabaseConnections []*sql.DB
	Queries             []string
	JobExecutionStates  []ExecutionState
	WorkingDir          string
	MutexLocks          map[string]*sync.Mutex
}

type ArchiveParameters struct {
	Enabled bool                         `yaml:"enabled"`
	Type    string                       `yaml:"type"`
	S3      *objectStorage.S3Driver      `yaml:"s3,omitempty"`
	Local   *objectStorage.LocalFSDriver `yaml:"local,omitempty"`
}
type CleanupParameters struct {
	Enabled bool `yaml:"enabled"`
}

type ArchivalPlan struct {
	DatabaseID string            `yaml:"databaseID"`
	Query      ArchivalPlanQuery `yaml:"query"`

	Workers            int `yaml:"workers"` // number of parallel workers
	DatabaseCredential DBCredential
	RuntimeParameters  RuntimeParams
	ArchiveStorage     ArchiveParameters `yaml:"archiveStorage"`
	Cleanup            CleanupParameters `yaml:"cleanup"`
}

func (a ArchiveParameters) Validate() error {
	switch a.Type {
	case "s3":
		if a.S3 == nil || a.S3.Bucket == "" || a.S3.Region == "" {
			return fmt.Errorf("S3 configuration missing required fields")
		}
	case "local":
		if a.Local == nil || a.Local.Path == "" {
			return fmt.Errorf("local configuration missing required fields")
		}
	default:
		return fmt.Errorf("unknown storage type: %s", a.Type)
	}
	return nil
}

func LoadArchivalPlan(filepath string) (*ArchivalPlan, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	// Unmarshal YAML into struct
	var plan ArchivalPlan
	err = yaml.Unmarshal(data, &plan)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	credfile, err := os.ReadFile("db_credentials.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	var dbcredfile CredentialFile
	err = yaml.Unmarshal(credfile, &dbcredfile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	plan.DatabaseCredential = dbcredfile.Databases[plan.DatabaseID]

	// Validate the archival storage configuration
	if err := plan.ArchiveStorage.Validate(); err != nil {
		return nil, fmt.Errorf("invalid archival storage configuration: %w", err)
	}

	return &plan, nil
}

func (ap ArchivalPlan) SaveExecutionState(stateFilePath string) error {
	//Create a copy of the archival plan to avoid modifying the original
	apTemp := ap

	//Remove the database connection from the copy
	apTemp.RuntimeParameters.DatabaseConnections = nil
	//Remove the database credentials from the copy
	apTemp.DatabaseCredential = DBCredential{}
	//Remove the mutex locks from the copy
	apTemp.RuntimeParameters.MutexLocks = nil

	// Convert the struct back to YAML
	data, err := yaml.Marshal(apTemp)
	if err != nil {
		return fmt.Errorf("failed to marshal archival plan: %w", err)
	}

	// Write the YAML data to a file
	err = os.WriteFile(stateFilePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write archival plan to file: %w", err)
	}

	fmt.Printf("Archival plan saved to %s\n", stateFilePath)
	return nil
}
