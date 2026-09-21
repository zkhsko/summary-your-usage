package usecase

import (
	"context"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"summary-your-usage/internal/entity"
)

type GroupRepository interface {
	List(context.Context) ([]entity.Group, error)
	Get(context.Context, int64) (entity.Group, error)
	Create(context.Context, entity.Group) (entity.Group, error)
	Update(context.Context, entity.Group) (entity.Group, error)
	Delete(context.Context, int64) error
}

type Group struct {
	repo GroupRepository
}

func NewGroup(repo GroupRepository) *Group {
	return &Group{repo: repo}
}

func (g *Group) List(ctx context.Context) ([]entity.Group, error) {
	return g.repo.List(ctx)
}

func (g *Group) Get(ctx context.Context, id int64) (entity.Group, error) {
	return g.repo.Get(ctx, id)
}

func (g *Group) Create(ctx context.Context, input entity.GroupInput) (entity.Group, error) {
	if err := normalizeGroup(&input); err != nil {
		return entity.Group{}, err
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	return g.repo.Create(ctx, entity.Group{
		Name:              input.Name,
		BillingMultiplier: *input.BillingMultiplier,
		VisibleOtherGroup: input.VisibleOtherGroup,
		Description:       input.Description,
		CreatedAt:         now,
		UpdatedAt:         now,
	})
}

func (g *Group) Update(ctx context.Context, id int64, input entity.GroupInput) (entity.Group, error) {
	if err := normalizeGroup(&input); err != nil {
		return entity.Group{}, err
	}
	return g.repo.Update(ctx, entity.Group{
		Id:                id,
		Name:              input.Name,
		BillingMultiplier: *input.BillingMultiplier,
		VisibleOtherGroup: input.VisibleOtherGroup,
		Description:       input.Description,
		UpdatedAt:         time.Now().UTC().Truncate(time.Microsecond),
	})
}

func (g *Group) Delete(ctx context.Context, id int64) error {
	return g.repo.Delete(ctx, id)
}

func normalizeGroup(input *entity.GroupInput) error {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if input.Name == "" || utf8.RuneCountInString(input.Name) > 100 ||
		utf8.RuneCountInString(input.Description) > 1000 || input.BillingMultiplier == nil ||
		math.IsNaN(*input.BillingMultiplier) || math.IsInf(*input.BillingMultiplier, 0) || *input.BillingMultiplier < 0 {
		return entity.ErrInvalidGroup
	}
	return nil
}
