package sqlgen

import (
	"smriti/parser"
)

type SQLGeneratorDriver interface {
	GenerateSQL(*parser.ArchivalPlan) []string
}
