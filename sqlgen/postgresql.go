package sqlgen

import (
	"fmt"
	"smriti/parser"
	"strconv"
	"strings"
)

type PostgresqlSQLGenerator struct{}

func (s *PostgresqlSQLGenerator) GenerateSQL(ap *parser.ArchivalPlan) []string {
	var sql []string
	// var whereConditionsStr string
	// if len(ap.Query.FilterConditions) > 0 {
	// 	whereConditionsStr = strings.Join(ap.Query.FilterConditions, " AND ")
	// 	whereConditionsStr = whereConditionsStr + " AND "
	// }

	if ap.Query.BatchingEnabled {
		// Batching enabled, generate multiple queries
		batchColumn := ap.Query.BatchColumn
		batchColumnType := ap.Query.BatchColumnType
		batchStep := ap.Query.BatchStep

		if batchColumnType == "int" {
			minVal, err := strconv.Atoi(getMinBatchColumnValue(ap))
			if err != nil {
				fmt.Println("Could not cast string to int")
			}
			maxVal, err := strconv.Atoi(getMaxBatchColumnValue(ap))
			if err != nil {
				fmt.Println("Could not cast string to int")
			}

			var batchWhereConditions []string
			minTmp := minVal
			for true {
				//fmt.Println(minTmp, minTmp + batchStep)
				whereConTmp := fmt.Sprintf(
					"%s >= %d AND %s < %d",
					batchColumn,
					minTmp,
					batchColumn,
					min(minTmp+batchStep, maxVal), // Do not exceed maxVal
				)
				batchWhereConditions = append(batchWhereConditions, whereConTmp)
				minTmp = minTmp + batchStep
				//break
				if minTmp > maxVal {
					break
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
