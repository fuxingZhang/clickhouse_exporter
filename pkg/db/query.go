package db

type KeyValueData struct {
	Key string
	Val float64
}

func GetKeyValueData(sql string, args ...any) (result []KeyValueData, err error) {
	rows, err := DB.Raw(sql).Rows()
	defer func() {
		err = rows.Close()
	}()
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
