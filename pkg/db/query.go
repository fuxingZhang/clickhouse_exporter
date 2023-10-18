package db

type KeyValueData struct {
	Key string
	Val float64
}

func GetKeyValueData(query string, args ...any) (result []KeyValueData, err error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item KeyValueData
		err = rows.Scan(&item.Key, &item.Val)
		if err != nil {
			return
		}
		result = append(result, item)
	}

	return
}
