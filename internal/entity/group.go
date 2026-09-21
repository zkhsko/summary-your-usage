package entity

import (
	"errors"
	"time"
)

var (
	ErrGroupNotFound   = errors.New("group not found")
	ErrGroupNameExists = errors.New("group name already exists")
	ErrInvalidGroup    = errors.New("invalid group")
)

type Group struct {
	Id                int64     `json:"id"`
	Name              string    `json:"name"`
	BillingMultiplier float64   `json:"billing_multiplier"`
	VisibleOtherGroup bool      `json:"visible_other_group"`
	Description       string    `json:"description"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type GroupInput struct {
	Name              string   `json:"name"`
	BillingMultiplier *float64 `json:"billing_multiplier"`
	VisibleOtherGroup bool     `json:"visible_other_group"`
	Description       string   `json:"description"`
}
