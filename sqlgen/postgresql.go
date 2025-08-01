package sqlgen


import (
    "fmt"
    "db-archive/parser"
    "strings"
    "strconv"
)



func GenerateSQL(ap *parser.ArchivalPlan) []string {

    var sql []string;
    whereConditionsStr := strings.Join(ap.Query.FilterConditions, " AND ")

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
                min(minTmp + batchStep, maxVal), // Do not exceed maxVal
            )
            batchWhereConditions = append(batchWhereConditions, whereConTmp)
            minTmp = minTmp + batchStep
            //break
            if minTmp > maxVal{
                break
            }

        }

        for i := 0; i < len(batchWhereConditions); i++ {
            sqlTmp := fmt.Sprintf("SELECT * from %s WHERE %s AND %s;", ap.Query.Table, whereConditionsStr, batchWhereConditions[i])
            //fmt.Println(sqlTmp)
            sql = append(sql, sqlTmp)
        }
    }

    return sql;
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