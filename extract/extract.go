package extract

import (
    "database/sql"
	"encoding/csv"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func QueryDynamic(db *sql.DB, query string) ([]map[string]interface{}, []string, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	var results []map[string]interface{}

	for rows.Next() {
		columnValues := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))

		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, nil, err
		}

		rowMap := make(map[string]interface{})
		for i, colName := range columns {
			val := columnValues[i]
			// Convert []byte to string for text fields
			b, ok := val.([]byte)
			if ok {
				rowMap[colName] = string(b)
			} else if val == nil {
				rowMap[colName] = "" // Handle NULL values
			} else {
				rowMap[colName] = val
			}
		}

		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return results, columns, nil
}

func WriteCSV(filename string, data []map[string]interface{}, columns []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write Header
	if err := writer.Write(columns); err != nil {
		return err
	}

	// Write Rows
	for _, row := range data {
		record := make([]string, len(columns))
		for i, col := range columns {
			val := fmt.Sprintf("%v", row[col])
			record[i] = val
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}