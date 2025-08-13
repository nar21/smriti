package parser

import (
	"database/sql"
	"fmt"
	"os"

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
type RuntimeParams struct {
	DryRun              bool
	ExecutionID         string
	DatabaseConnections []*sql.DB
	Queries             []string
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
			return fmt.Errorf("Local configuration missing required fields")
		}
	default:
		return fmt.Errorf("Unknown storage type: %s", a.Type)
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
