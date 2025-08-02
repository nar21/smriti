package parser

import (
	"database/sql"
	"fmt"
	"os"

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
	Type   string `yaml:"type"`
	Bucket struct {
		BucketName string `yaml:"bucketName"`
		Region     string `yaml:"region"`
	}
}

type ArchivalPlan struct {
	DatabaseID string            `yaml:"databaseID"`
	Query      ArchivalPlanQuery `yaml:"query"`

	Workers            int `yaml:"workers"` // number of parallel workers
	DatabaseCredential DBCredential
	RuntimeParameters  RuntimeParams
	Archive            ArchiveParameters
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

	return &plan, nil
}
