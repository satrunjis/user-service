package service

import (
	"context"
)

type CacheService interface {
	Get(ctx context.Context, key *string) (*[]byte, error)
	Set(ctx context.Context, key *string, value *[]byte) error
}
