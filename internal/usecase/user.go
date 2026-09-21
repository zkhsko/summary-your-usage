package usecase

import (
	"context"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"summary-your-usage/internal/entity"
)

// UserRepository 由业务层定义，持久化实现通过构造函数注入。
type UserRepository interface {
	List(context.Context) ([]entity.User, error)
	Get(context.Context, int64) (entity.User, error)
	Create(context.Context, entity.User) (entity.User, error)
	Update(context.Context, entity.User) (entity.User, error)
	Delete(context.Context, int64) error
}

type User struct {
	repo UserRepository
}

func NewUser(repo UserRepository) *User {
	return &User{repo: repo}
}

func (u *User) List(ctx context.Context) ([]entity.User, error) {
	return u.repo.List(ctx)
}

func (u *User) Get(ctx context.Context, id int64) (entity.User, error) {
	return u.repo.Get(ctx, id)
}

func (u *User) Create(ctx context.Context, input entity.UserInput) (entity.User, error) {
	if err := normalize(&input); err != nil {
		return entity.User{}, err
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	return u.repo.Create(ctx, entity.User{
		Name:      input.Name,
		Email:     input.Email,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (u *User) Update(ctx context.Context, id int64, input entity.UserInput) (entity.User, error) {
	if err := normalize(&input); err != nil {
		return entity.User{}, err
	}
	return u.repo.Update(ctx, entity.User{
		Id:        id,
		Name:      input.Name,
		Email:     input.Email,
		UpdatedAt: time.Now().UTC().Truncate(time.Microsecond),
	})
}

func (u *User) Delete(ctx context.Context, id int64) error {
	return u.repo.Delete(ctx, id)
}

func normalize(input *entity.UserInput) error {
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	address, err := mail.ParseAddress(input.Email)
	if input.Name == "" || utf8.RuneCountInString(input.Name) > 100 || len(input.Email) > 254 ||
		err != nil || address.Address != input.Email {
		return entity.ErrInvalidUser
	}
	return nil
}
