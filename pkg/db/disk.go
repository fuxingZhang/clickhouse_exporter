package db

type DiskData struct {
	Disk       string
	FreeSpace  float64
	TotalSpace float64
}

func GetDiskData(query string, args ...any) (result []DiskData, err error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item DiskData
		err = rows.Scan(&item.Disk, &item.FreeSpace, &item.TotalSpace)
		if err != nil {
			return
		}
		result = append(result, item)
	}

	return
}
