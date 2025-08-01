package sqlgen

import (
    "db-archive/parser"
)

type SQLGeneratorDriver interface {
    GenerateSQL(*parser.ArchivalPlan) []string
}