package enviar

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
)

// Config holds the Redis connection and worker pool configuration.
type Config struct {
	Namespace   string // Redis key namespace (e.g. "myapp-work").
	RedisURL    string // "host:port" or "redis://:password@host:port".
	RedisDB     int    // Redis database number (0–15); ignored for redis:// URLs.
	Concurrency uint   // Number of concurrent workers (default 10).
}

// LoadEnv builds a Config from environment variables:
//
//   - ENVIAR_NAMESPACE   (default "enviar")
//   - ENVIAR_REDIS_URL   (default "localhost:6379") — supports redis:// URLs
//   - ENVIAR_REDIS_DB    (default "0")
//   - ENVIAR_CONCURRENCY (default "10")
func LoadEnv() Config {
	return Config{
		Namespace:   envOrDefault("ENVIAR_NAMESPACE", "enviar"),
		RedisURL:    envOrDefault("ENVIAR_REDIS_URL", "localhost:6379"),
		RedisDB:     envOrDefaultInt("ENVIAR_REDIS_DB", 0),
		Concurrency: uint(envOrDefaultInt("ENVIAR_CONCURRENCY", 10)),
	}
}

// NewPool creates a *redis.Pool from the configuration.
// It supports both plain "host:port" and full "redis://" URLs.
func (c Config) NewPool() *redis.Pool {
	isURL := strings.HasPrefix(c.RedisURL, "redis://") ||
		strings.HasPrefix(c.RedisURL, "rediss://")

	return &redis.Pool{
		MaxActive:   20,
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			var conn redis.Conn
			var err error

			if isURL {
				conn, err = redis.DialURL(c.RedisURL)
			} else {
				conn, err = redis.Dial("tcp", c.RedisURL)
			}
			if err != nil {
				return nil, fmt.Errorf("redis dial %s: %w", c.RedisURL, err)
			}

			// SELECT db only for plain host:port — redis:// URLs carry it in the path.
			if !isURL && c.RedisDB != 0 {
				if _, err := conn.Do("SELECT", c.RedisDB); err != nil {
					conn.Close()
					return nil, fmt.Errorf("redis select db %d: %w", c.RedisDB, err)
				}
			}
			return conn, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envOrDefaultInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
