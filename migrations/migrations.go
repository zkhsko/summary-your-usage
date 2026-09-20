package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
)

//go:embed sqlite/*.sql mysql/*.sql
var files embed.FS

// New 为每个数据库创建独立的 goose Provider，避免修改全局方言或文件系统。
func New(db *sql.DB, driver string) (*goose.Provider, error) {
	var dialect goose.Dialect
	switch driver {
	case "sqlite":
		dialect = goose.DialectSQLite3
	case "mysql":
		dialect = goose.DialectMySQL
	default:
		return nil, fmt.Errorf("unsupported migration driver %q", driver)
	}
	fsys, err := fs.Sub(files, driver)
	if err != nil {
		return nil, fmt.Errorf("open migrations: %w", err)
	}
	return goose.NewProvider(dialect, db, fsys)
}
