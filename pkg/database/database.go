package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"
)

const DefaultSQLiteDSN = "file:app.db?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)"

func Open(ctx context.Context, driver, dsn string) (*sql.DB, error) {
	switch driver {
	case "sqlite":
		if dsn == "" {
			dsn = DefaultSQLiteDSN
		}
	case "mysql":
		if dsn == "" {
			return nil, fmt.Errorf("database.dsn is required for mysql")
		}
		cfg, err := mysql.ParseDSN(dsn)
		if err != nil {
			return nil, fmt.Errorf("parse mysql DSN: %w", err)
		}
		cfg.ParseTime = true
		cfg.Loc = time.UTC
		dsn = cfg.FormatDSN()
	default:
		return nil, fmt.Errorf("unsupported database.driver %q (expected sqlite or mysql)", driver)
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if driver == "sqlite" {
		// SQLite 串行写入，单连接也能保证内存数据库与连接级配置一致。
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)
	} else {
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(3 * time.Minute)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return db, nil
}
