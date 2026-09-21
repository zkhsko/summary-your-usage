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

type Group struct {
	db *sql.DB
}

func NewGroup(db *sql.DB) *Group {
	return &Group{db: db}
}

func (r *Group) List(ctx context.Context) ([]entity.Group, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, billing_multiplier, visible_other_group, description, created_at, updated_at FROM `groups` ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	groups := []entity.Group{}
	for rows.Next() {
		var group entity.Group
		if err := rows.Scan(&group.Id, &group.Name, &group.BillingMultiplier, &group.VisibleOtherGroup, &group.Description, &group.CreatedAt, &group.UpdatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, rows.Err()
}

func (r *Group) Get(ctx context.Context, id int64) (entity.Group, error) {
	var group entity.Group
	err := r.db.QueryRowContext(ctx, "SELECT id, name, billing_multiplier, visible_other_group, description, created_at, updated_at FROM `groups` WHERE id = ?", id).
		Scan(&group.Id, &group.Name, &group.BillingMultiplier, &group.VisibleOtherGroup, &group.Description, &group.CreatedAt, &group.UpdatedAt)
	return group, groupError(err)
}

func (r *Group) Create(ctx context.Context, group entity.Group) (entity.Group, error) {
	result, err := r.db.ExecContext(ctx, "INSERT INTO `groups` (name, billing_multiplier, visible_other_group, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		group.Name, group.BillingMultiplier, group.VisibleOtherGroup, group.Description, group.CreatedAt, group.UpdatedAt)
	if err != nil {
		return entity.Group{}, groupError(err)
	}
	group.Id, err = result.LastInsertId()
	return group, err
}

func (r *Group) Update(ctx context.Context, group entity.Group) (entity.Group, error) {
	_, err := r.db.ExecContext(ctx, "UPDATE `groups` SET name = ?, billing_multiplier = ?, visible_other_group = ?, description = ?, updated_at = ? WHERE id = ?",
		group.Name, group.BillingMultiplier, group.VisibleOtherGroup, group.Description, group.UpdatedAt, group.Id)
	if err != nil {
		return entity.Group{}, groupError(err)
	}
	return r.Get(ctx, group.Id)
}

func (r *Group) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM `groups` WHERE id = ?", id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return entity.ErrGroupNotFound
	}
	return nil
}

func groupError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return entity.ErrGroupNotFound
	}
	var sqliteErr *sqlite.Error
	if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
		return entity.ErrGroupNameExists
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return entity.ErrGroupNameExists
	}
	return err
}
