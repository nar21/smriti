package archival-plan


import (
    "fmt"
)

type ArchivalPlanQuery struct {
    FilterColumn string
    FilterColumnType string // "int" or "date"
    FilterColumnMin string // Parse this variable according to FilterColumnType
    FilterColumnMax string // Parse this variable according to FilterColumnType
}

type ArchivalPlan struct {
    DatabaseID string,
    Query ArchivalPlanQuery,
    BatchStep int // number if FilterColumnType is int, days if FilterColumnType is date
    Workers int // number of workers that will parallely read from database
}


func LoadArchivalPlan(path string) (ArchivalPlan) {

}