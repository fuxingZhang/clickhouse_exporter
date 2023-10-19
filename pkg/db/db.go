package db

import (
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
)

var DB *gorm.DB

func NewClient(dsn string) (err error) {
	DB, err = gorm.Open(clickhouse.Open(dsn), &gorm.Config{})
	return
}
