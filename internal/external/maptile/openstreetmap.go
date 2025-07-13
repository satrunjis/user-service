package maptile

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"time"
)

type OpenStreetMapService struct {
	logger    *slog.Logger
	client    *http.Client
	baseURL   string
	userAgent string
}

func Init(logger *slog.Logger) *OpenStreetMapService {
	return &OpenStreetMapService{
		client: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:    10,
				IdleConnTimeout: 30 * time.Second,
			},
		},
		baseURL:   "https://tile.openstreetmap.org",
		userAgent: "user-service/1.0",
		logger:    logger,
	}
}
func (s *OpenStreetMapService) GetMapTile(ctx context.Context, lat, lon float64, zoom int) (*[]byte, error) {
	const op = "maptile.GetMapTile"
	log := s.logger.With("operation", op)
	log.Debug("getting map tile", "lat", lat, "lon", lon, "zoom", zoom)
	if err := s.validateParameters(lat, lon, zoom); err != nil {
		log.WarnContext(ctx, "params is not valid")
		return nil, err
	}
	x, y := s.latLonToTileXY(lat, lon, zoom)
	url := fmt.Sprintf("%s/%d/%d/%d.png", s.baseURL, zoom, x, y)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.ErrorContext(ctx, "failed to create request", "error", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", s.userAgent)
	resp, err := s.client.Do(req)
	if err != nil {
		log.ErrorContext(ctx, "failed to create request", "error", err)
		return nil, fmt.Errorf("failed to fetch tile: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.ErrorContext(ctx, "failed to fetch tile", "error", err)
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	tileData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.ErrorContext(ctx, "failed to read tile data", "error", err)
		return nil, fmt.Errorf("failed to read tile data: %w", err)
	}
	log.InfoContext(ctx, "successfully get map tile")
	return &tileData, nil
}
func (s *OpenStreetMapService) validateParameters(lat, lon float64, zoom int) error {
	if lat < -90 || lat > 90 {
		return fmt.Errorf("invalid latitude: %f (must be between -90 and 90)", lat)
	}
	if lon < -180 || lon > 180 {
		return fmt.Errorf("invalid longitude: %f (must be between -180 and 180)", lon)
	}
	if zoom < 0 || zoom > 19 {
		return fmt.Errorf("invalid zoom level: %d (must be between 0 and 19)", zoom)
	}
	return nil
}
func (s *OpenStreetMapService) latLonToTileXY(lat, lon float64, zoom int) (int, int) {
	// Используется проекция Web Mercator (EPSG:3857)

	n := math.Pow(2, float64(zoom))
	x := int(math.Floor((lon + 180) / 360 * n))
	latRad := lat * math.Pi / 180
	y := int(math.Floor((1 - math.Log(math.Tan(latRad)+1/math.Cos(latRad))/math.Pi) / 2 * n))

	return x, y
}
