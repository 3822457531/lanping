package repo

import (
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func Open(dsn string) (*sqlx.DB, error) {
	// phpstudy / MySQL 8 本地：缩短空闲连接寿命，并补齐超时参数，避免 invalid connection
	if dsn != "" && !strings.Contains(dsn, "timeout=") {
		if strings.Contains(dsn, "?") {
			dsn += "&timeout=5s&readTimeout=10s&writeTimeout=10s"
		} else {
			dsn += "?timeout=5s&readTimeout=10s&writeTimeout=10s"
		}
	}
	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetConnMaxIdleTime(60 * time.Second)
	for i := 0; i < 30; i++ {
		if err = db.Ping(); err == nil {
			return db, nil
		}
		time.Sleep(time.Second)
	}
	return nil, fmt.Errorf("mysql ping failed: %w", err)
}
