package service

import (
	"context"
	"fmt"
	"log"
)

type MapService interface {
	GetMapTile(ctx context.Context, lat, lon float64, zoom int) (*[]byte, error)
}

func (s *UserService) GetMapTile(ctx context.Context, userID *string, zoom int) (*[]byte, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, mapRepositoryError(err, "getmaptile")
	}

	if user.Location == nil {
		return nil, NewServiceError(ErrCodeInvalidInput, "User location is required")
	}

	cacheKey := fmt.Sprintf("tile_%f_%f_%d", user.Location.Lat, user.Location.Lon, zoom)

	if tile, err := s.mapCache.Get(ctx, &cacheKey); err == nil {
		return tile, nil
	}

	tileData, err := s.mapService.GetMapTile(ctx, user.Location.Lat, user.Location.Lon, zoom)
	if err != nil {
		return nil, NewServiceError(ErrCodeInternal, "Failed to get map tile from map service")
	}

	if err := s.mapCache.Set(ctx, &cacheKey, tileData); err != nil {
		log.Printf("Cache set failed: %v", err)
	}

	return tileData, nil
}
