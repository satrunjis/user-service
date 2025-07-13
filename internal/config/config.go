package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type ElasticConfig struct {
	URL   string `yaml:"url" env:"ES_URL" env-default:"http://localhost:9200"`
	Index string `yaml:"index" env:"ES_INDEX" env-default:"users"`
}
type HTTPServerConfig struct {
	Address      string        `yaml:"address" env:"HTTP_ADDRESS" env-default:"0.0.0.0:8080"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env:"HTTP_READ_TIMEOUT" env-default:"30s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env:"HTTP_WRITE_TIMEOUT" env-default:"30s"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
}
type CacheConfig struct {
	URL string        `yaml:"url" env:"REDIS_URL" env-default:"localhost:6379"`
	TTL time.Duration `yaml:"ttl" env:"CACHE_TTL" env-default:"168h"`
}
type OpenStreetMapConfig struct {
	BaseURL   string        `yaml:"base_url" env:"OSM_BASE_URL" env-default:"https://tile.openstreetmap.org"`
	Timeout   time.Duration `yaml:"timeout" env:"OSM_TIMEOUT" env-default:"10s"`
	UserAgent string        `yaml:"user_agent" env:"OSM_USER_AGENT" env-default:"UserService/1.0 (satrunjis@mail.ru)"`
	ZoomLevel int           `yaml:"zoom_level" env:"OSM_ZOOM_LEVEL" env-default:"15"`
}
type Config struct {
	Env                 string              `yaml:"env" env:"ENV" env-default:"development"`
	ElasticConfig       ElasticConfig       `yaml:"elastic"`
	HTTPServerConfig    HTTPServerConfig    `yaml:"http_server"`
	CacheConfig         CacheConfig         `yaml:"cache"`
	OpenStreetMapConfig OpenStreetMapConfig `yaml:"openstreetmap"`
}

func Load() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config.yaml"
		log.Printf("CONFIG_PATH not set, using default: %s", configPath)
	}

	var cfg Config
	var err error

	if _, err = os.Stat(configPath); err == nil {
		log.Printf("Loading config from: %s", configPath)
		if err = cleanenv.ReadConfig(configPath, &cfg); err != nil {
			log.Fatalf("Error reading config file: %v", err)
		}
	} else {
		log.Printf("Config file not found with path: %s", configPath)
	}

	if err = cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("Error reading env vars: %v", err)
	}

	if cfg.Env != "production" {
		log.Printf("Configuration: %+v", cfg)
	}

	return &cfg
}
