package logging

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"smriti/parser"
	"strings"
)

type JobLogger struct {
	jobID  string
	file   *os.File
	buffer bytes.Buffer
}

// NewJobLogger initializes a logger for a specific job ID
func NewJobLogger(jobID string, ap *parser.ArchivalPlan) (*JobLogger, error) {
	filepath := fmt.Sprintf(
		"%s/job-%s.log",
		ap.RuntimeParameters.WorkingDir,
		jobID,
	)

	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	return &JobLogger{jobID: jobID, file: f, buffer: bytes.Buffer{}}, nil
}

// Log writes to stdout and the job log file
func (jl *JobLogger) Log(msg ...string) error {
	// Create a logger that writes to a buffer
	logger := log.New(&jl.buffer, "", log.Ldate|log.Ltime)

	// Use a string builder to concatenate messages with spaces
	var sb strings.Builder
	for _, m := range msg {
		sb.WriteString(m + " ")
	}
	// Join all messages with a space, like Println
	logger.Println(fmt.Sprintf("[Job %s] ", jl.jobID) + sb.String() + "\n")

	// Get the log string from the buffer and write to both stdout and file
	logString := jl.buffer.String()

	fmt.Println(logString)
	if _, err := jl.file.WriteString(logString); err != nil {
		return fmt.Errorf("failed to write log: %w", err)
	}
	return nil
}

// Close closes the file
func (jl *JobLogger) Close() error {
	return jl.file.Close()
}
