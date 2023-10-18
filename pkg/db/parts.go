package db

type PartsData struct {
	Database string
	Table    string
	Bytes    float64
	Parts    float64
	Rows     float64
}

func GetPartsData(query string, args ...any) (result []PartsData, err error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item PartsData
		err = rows.Scan(&item.Database, &item.Table, &item.Bytes, &item.Parts, &item.Rows)
		if err != nil {
			return
		}
		result = append(result, item)
	}

	return
}
