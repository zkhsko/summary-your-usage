package persistent

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"

	"summary-your-usage/internal/entity"
)

type User struct {
	db *sql.DB
}

func NewUser(db *sql.DB) *User {
	return &User{db: db}
}

func (r *User) List(ctx context.Context) ([]entity.User, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, email, created_at, updated_at FROM users ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := []entity.User{}
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(&user.Id, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (r *User) Get(ctx context.Context, id int64) (entity.User, error) {
	var user entity.User
	err := r.db.QueryRowContext(ctx, "SELECT id, name, email, created_at, updated_at FROM users WHERE id = ?", id).
		Scan(&user.Id, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	return user, userError(err)
}

func (r *User) Create(ctx context.Context, user entity.User) (entity.User, error) {
	result, err := r.db.ExecContext(ctx, "INSERT INTO users (name, email, created_at, updated_at) VALUES (?, ?, ?, ?)",
		user.Name, user.Email, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return entity.User{}, userError(err)
	}
	user.Id, err = result.LastInsertId()
	return user, err
}

func (r *User) Update(ctx context.Context, user entity.User) (entity.User, error) {
	_, err := r.db.ExecContext(ctx, "UPDATE users SET name = ?, email = ?, updated_at = ? WHERE id = ?",
		user.Name, user.Email, user.UpdatedAt, user.Id)
	if err != nil {
		return entity.User{}, userError(err)
	}
	// 回读也能正确处理 MySQL 更新相同值时 RowsAffected 为 0 的情况。
	return r.Get(ctx, user.Id)
}

func (r *User) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return entity.ErrUserNotFound
	}
	return nil
}

func userError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return entity.ErrUserNotFound
	}
	var sqliteErr *sqlite.Error
	if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
		return entity.ErrEmailExists
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return entity.ErrEmailExists
	}
	return err
}
