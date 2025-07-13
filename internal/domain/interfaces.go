package domain

import (
	"context"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) error

	GetByID(ctx context.Context, id *string) (*User, error)
	Search(ctx context.Context, filters *UserFilter) ([]*User, error)

	Replace(ctx context.Context, user *User) error
	UpdatePartial(ctx context.Context, user *User) error

	Delete(ctx context.Context, id *string) error
}
