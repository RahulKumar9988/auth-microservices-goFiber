package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type DBConfig struct {
	URL            string
	MaxOpenConns   int
	MaxIdleConns   int
	ConnMaxLife    time.Duration
	ConnectTimeout time.Duration
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type Config struct {
	AppPort  string
	DB       DBConfig
	RedisURL RedisConfig
	JWT      JWTConfig
}

func Load() *Config {
	cfg := &Config{}

	// LOAD APP ENV
	cfg.AppPort = getEnv("APP_PORT", "8080")

	// LOAD DB ENV
	cfg.DB.URL = mustGetEnv("DB_URL")
	cfg.DB.MaxOpenConns = getEnvInt("DB_MAX_OPEN", 25)
	cfg.DB.MaxIdleConns = getEnvInt("DB_MAX_IDLE", 10)
	cfg.DB.ConnMaxLife = time.Minute * 30
	cfg.DB.ConnectTimeout = time.Second * 5

	// LOAD REDIS ENV
	cfg.RedisURL.Addr = mustGetEnv("REDIS_ADDR")
	cfg.RedisURL.Password = getEnv("REDIS_PASSWORD", "")
	cfg.RedisURL.DB = getEnvInt("REDIS_DB", 0)

	// LOAD JWT ENV
	cfg.JWT.AccessSecret = mustGetEnv("JWT_ACCESS_SECRET")
	cfg.JWT.RefreshSecret = mustGetEnv("JWT_REFRESH_SECRET")
	cfg.JWT.AccessTTL = mustGetEnvDuration("ACCESS_TOKEN_TTL")
	cfg.JWT.RefreshTTL = mustGetEnvDuration("REFRESH_TOKEN_TTL")

	return cfg
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)

	if val == "" {
		log.Fatal("missing requied env: ", key)
	}
	return val
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func mustGetEnvDuration(key string) time.Duration {
	val := os.Getenv(key)

	if val == "" {
		log.Fatal("missing required env", key)
	}

	d, err := time.ParseDuration(val)

	if err != nil {
		log.Fatal("invalid duration", err)
	}

	return d
}

func getEnvInt(key string, defaultVal int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		log.Printf("invalid int value for env %s: %v", key, err)
	}

	return val
}
