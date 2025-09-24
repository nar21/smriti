package sqlgen

import (
	"fmt"
	"smriti/logging"
	"smriti/parser"
	"strconv"
	"strings"
	"time"
)

type PostgresqlSQLGenerator struct {
	Logger *logging.JobLogger
}

func (s *PostgresqlSQLGenerator) GenerateSQL(ap *parser.ArchivalPlan) []string {
	var sql []string

	if ap.Query.BatchingEnabled {
		// Batching enabled, generate multiple queries
		batchColumn := ap.Query.BatchColumn
		batchColumnType := ap.Query.BatchColumnType
		batchStep := ap.Query.BatchStep
		var batchWhereConditions []string

		if batchColumnType == "int" {
			minVal, err := strconv.Atoi(getMinBatchColumnValue(ap))
			if err != nil {
				s.Logger.Log("Could not cast string to int")
			}
			maxVal, err := strconv.Atoi(getMaxBatchColumnValue(ap))
			if err != nil {
				s.Logger.Log("Could not cast string to int")
			}
			minTmp := minVal
			for {
				whereConTmp := fmt.Sprintf(
					"%s >= %d AND %s < %d",
					batchColumn,
					minTmp,
					batchColumn,
					min(minTmp+batchStep, maxVal), // Do not exceed maxVal
				)
				batchWhereConditions = append(batchWhereConditions, whereConTmp)
				minTmp = minTmp + batchStep
				if minTmp > maxVal {
					break
				}
			}
		} else if batchColumnType == "datetime" {
			dateFormat := "2006-01-02T15:04:05" // Golang reference time format

			minValUser, err := time.Parse(dateFormat, getMinBatchColumnValue(ap))
			if err != nil {
				s.Logger.Log("Could not cast string to datetime")
			}
			maxValUser, err := time.Parse(dateFormat, getMaxBatchColumnValue(ap))
			if err != nil {
				s.Logger.Log("Could not cast string to datetime")
			}

			minTmp := minValUser
			for {
				// Find the minimum value between (minTmp + batchStep) and maxValUser
				// This is to ensure that the batch does not exceed the user-provided max
				// value.
				// e.g., if minTmp is 2023-01-01, batchStep is 10 days, and maxValUser is
				// 2023-01-05, then the next batch should end at 2023-01-05 and not
				// 2023-01-11.
				// This is important to ensure that the last batch does not exceed the
				// user-provided max value.

				maxTmp := minTmp.AddDate(0, 0, batchStep) // Add batchStep days
				var maxValBatch time.Time
				var breakSignal bool = false
				if maxTmp.Before(maxValUser) {
					maxValBatch = maxTmp
				} else {
					maxValBatch = maxValUser
					breakSignal = true
				}

				whereConTmp := fmt.Sprintf(
					"%s >= '%s' AND %s < '%s'",
					batchColumn,
					minTmp.Format(dateFormat),
					batchColumn,
					maxValBatch.Format(dateFormat),
				)
				batchWhereConditions = append(batchWhereConditions, whereConTmp)
				minTmp = maxTmp

				// if the upper limit of this batch is the user-provided max value,
				// then we are done.
				if breakSignal {
					break
				}
			}
		}

		for i := 0; i < len(batchWhereConditions); i++ {
			whereConditionsStr := strings.Join(
				append(ap.Query.FilterConditions, batchWhereConditions[i]),
				" AND ")
			sqlTmp := fmt.Sprintf("SELECT * from %s WHERE %s;", ap.Query.Table, whereConditionsStr)
			//fmt.Println(sqlTmp)

			sql = append(sql, sqlTmp)
		}
	} else {
		// No batching, single query
		var sqlTmp string

		// if there are filter conditions, include them in the query
		if len(ap.Query.FilterConditions) > 0 {
			whereConditionsStr := strings.Join(ap.Query.FilterConditions, " AND ")
			sqlTmp = fmt.Sprintf("SELECT * from %s WHERE %s;", ap.Query.Table, whereConditionsStr)
		} else {
			// no filter conditions
			sqlTmp = fmt.Sprintf("SELECT * from %s;", ap.Query.Table)
		}
		sql = append(sql, sqlTmp)
	}
	return sql
}

func getMinBatchColumnValue(ap *parser.ArchivalPlan) string { //TODO: return value should also support date
	// This function returns the minimum value of the batch column, which will
	// be used as the starting value to generate the ranges for every batch.
	// Currently, this value is to be provided by the user in the archival plan.

	return ap.Query.BatchColumnMin
}

func getMaxBatchColumnValue(ap *parser.ArchivalPlan) string { //TODO: return value should also support date
	// This function returns the maximum value of the batch column, which will
	// be used as the maximum value, beyond which archival should not be done.
	// Currently, this value is to be provided by the user in the archival plan.

	return ap.Query.BatchColumnMax
}
