package sqlgen


import (
    "fmt"
    "db-archive/parser"
    "strings"
)

func GenerateSQL(ap *parser.ArchivalPlan) string {

    var sql string;
    where_conditions := strings.Join(ap.Query.FilterConditions, " AND ")

    sql = fmt.Sprintf("SELECT * from %s WHERE %s;", ap.Query.Table, where_conditions)
    return sql;
    //return "SELECT * FROM flights;"
}