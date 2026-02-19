package enviar

import (
	"os"
	"testing"
)

func TestLoadEnv_Defaults(t *testing.T) {
	// Clear any existing env vars.
	for _, key := range []string{
		"ENVIAR_NAMESPACE",
		"ENVIAR_REDIS_URL",
		"ENVIAR_REDIS_DB",
		"ENVIAR_CONCURRENCY",
	} {
		t.Setenv(key, "")
		os.Unsetenv(key)
	}

	cfg := LoadEnv()

	if cfg.Namespace != "enviar" {
		t.Errorf("Namespace = %q, want %q", cfg.Namespace, "enviar")
	}
	if cfg.RedisURL != "localhost:6379" {
		t.Errorf("RedisURL = %q, want %q", cfg.RedisURL, "localhost:6379")
	}
	if cfg.RedisDB != 0 {
		t.Errorf("RedisDB = %d, want 0", cfg.RedisDB)
	}
	if cfg.Concurrency != 10 {
		t.Errorf("Concurrency = %d, want 10", cfg.Concurrency)
	}
}

func TestLoadEnv_CustomValues(t *testing.T) {
	t.Setenv("ENVIAR_NAMESPACE", "myapp")
	t.Setenv("ENVIAR_REDIS_URL", "redis://:secret@redis.example.com:6380")
	t.Setenv("ENVIAR_REDIS_DB", "3")
	t.Setenv("ENVIAR_CONCURRENCY", "25")

	cfg := LoadEnv()

	if cfg.Namespace != "myapp" {
		t.Errorf("Namespace = %q, want %q", cfg.Namespace, "myapp")
	}
	if cfg.RedisURL != "redis://:secret@redis.example.com:6380" {
		t.Errorf("RedisURL = %q, want custom URL", cfg.RedisURL)
	}
	if cfg.RedisDB != 3 {
		t.Errorf("RedisDB = %d, want 3", cfg.RedisDB)
	}
	if cfg.Concurrency != 25 {
		t.Errorf("Concurrency = %d, want 25", cfg.Concurrency)
	}
}

func TestLoadEnv_InvalidInt_FallsBackToDefault(t *testing.T) {
	t.Setenv("ENVIAR_REDIS_DB", "not-a-number")
	t.Setenv("ENVIAR_CONCURRENCY", "abc")

	cfg := LoadEnv()

	if cfg.RedisDB != 0 {
		t.Errorf("RedisDB = %d, want 0 (fallback)", cfg.RedisDB)
	}
	if cfg.Concurrency != 10 {
		t.Errorf("Concurrency = %d, want 10 (fallback)", cfg.Concurrency)
	}
}

func TestEnvOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		envVal   string
		fallback string
		want     string
	}{
		{"env set", "TEST_KEY_1", "value", "default", "value"},
		{"env empty", "TEST_KEY_2", "", "default", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVal != "" {
				t.Setenv(tt.key, tt.envVal)
			} else {
				os.Unsetenv(tt.key)
			}
			got := envOrDefault(tt.key, tt.fallback)
			if got != tt.want {
				t.Errorf("envOrDefault(%q, %q) = %q, want %q",
					tt.key, tt.fallback, got, tt.want)
			}
		})
	}
}

func TestEnvOrDefaultInt(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		envVal   string
		fallback int
		want     int
	}{
		{"valid int", "TEST_INT_1", "42", 0, 42},
		{"invalid int", "TEST_INT_2", "xyz", 5, 5},
		{"empty env", "TEST_INT_3", "", 7, 7},
		{"negative int", "TEST_INT_4", "-3", 0, -3},
		{"zero", "TEST_INT_5", "0", 99, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVal != "" {
				t.Setenv(tt.key, tt.envVal)
			} else {
				os.Unsetenv(tt.key)
			}
			got := envOrDefaultInt(tt.key, tt.fallback)
			if got != tt.want {
				t.Errorf("envOrDefaultInt(%q, %d) = %d, want %d",
					tt.key, tt.fallback, got, tt.want)
			}
		})
	}
}

func TestConfig_NewPool_PlainAddress(t *testing.T) {
	cfg := Config{
		Namespace:   "test",
		RedisURL:    "localhost:6379",
		RedisDB:     0,
		Concurrency: 5,
	}
	pool := cfg.NewPool()
	if pool == nil {
		t.Fatal("NewPool returned nil")
	}
	if pool.MaxActive != 20 {
		t.Errorf("MaxActive = %d, want 20", pool.MaxActive)
	}
	if pool.MaxIdle != 10 {
		t.Errorf("MaxIdle = %d, want 10", pool.MaxIdle)
	}
	if !pool.Wait {
		t.Error("Wait should be true")
	}
}

func TestConfig_NewPool_RedisURL(t *testing.T) {
	cfg := Config{
		Namespace:   "test",
		RedisURL:    "redis://:password@redis.example.com:6380/2",
		Concurrency: 5,
	}
	pool := cfg.NewPool()
	if pool == nil {
		t.Fatal("NewPool returned nil")
	}
}

func TestConfig_NewPool_RedissURL(t *testing.T) {
	cfg := Config{
		Namespace:   "test",
		RedisURL:    "rediss://:password@redis.example.com:6380/2",
		Concurrency: 5,
	}
	pool := cfg.NewPool()
	if pool == nil {
		t.Fatal("NewPool returned nil")
	}
}
