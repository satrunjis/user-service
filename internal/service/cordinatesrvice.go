package service

import (
	"github.com/satrunjis/user-service/internal/domain"
)

type UserService struct {
	userRepo   domain.UserRepository
	mapCache   CacheService
	mapService MapService
}

func NewUserService(repo domain.UserRepository, cache CacheService, maps MapService) *UserService {
	return &UserService{
		userRepo:   repo,
		mapCache:   cache,
		mapService: maps,
	}
}
