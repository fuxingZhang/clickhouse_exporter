package db

import (
	"database/sql"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type Option struct {
	MaxExecutionTime int
	MaxIdleConns     int
	MaxOpenConns     int
	ConnMaxLifetime  time.Duration
	DialTimeout      time.Duration
}

var db *sql.DB

func InitTCPClient(addr, username, password string, opt Option) {
	db = clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: username,
			Password: password,
		},
		// TLS: &tls.Config{
		// 	InsecureSkipVerify: true,
		// },
		Settings: clickhouse.Settings{
			"max_execution_time": opt.MaxExecutionTime,
		},
		DialTimeout: opt.DialTimeout,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		Debug:                false,
		BlockBufferSize:      10,
		MaxCompressionBuffer: 10240,
		ClientInfo: clickhouse.ClientInfo{ // optional, please see Client info section in the README.md
			Products: []struct {
				Name    string
				Version string
			}{
				{Name: "clickhouse_exporter", Version: "0.1"},
			},
		},
	})
	db.SetMaxIdleConns(opt.MaxIdleConns)
	db.SetMaxOpenConns(opt.MaxOpenConns)
	db.SetConnMaxLifetime(opt.ConnMaxLifetime)
}

func InitHTTPClient(addr, username, password string, opt Option) {
	db = clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: username,
			Password: password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": opt.MaxExecutionTime,
		},
		DialTimeout: opt.DialTimeout,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		Protocol: clickhouse.HTTP,
	})
}

func Ping() error {
	return db.Ping()
}
