package main

import (
"os"
"testing"
)

func TestRedisPasswordEnvOverride(t *testing.T) {
// Save original BUBBLE_CONFIG and REDIS_PASSWORD
originalConfig := os.Getenv("BUBBLE_CONFIG")
originalPassword := os.Getenv("REDIS_PASSWORD")
defer func() {
os.Setenv("BUBBLE_CONFIG", originalConfig)
os.Setenv("REDIS_PASSWORD", originalPassword)
}()

// Set test config
os.Setenv("BUBBLE_CONFIG", "/tmp/test_config.yml")

// Test 1: Without REDIS_PASSWORD env var, should use config value
os.Unsetenv("REDIS_PASSWORD")
if err := loadConfig(); err != nil {
t.Fatalf("Failed to load config: %v", err)
}
if config.Redis.Password != "config-password" {
t.Errorf("Expected password 'config-password', got '%s'", config.Redis.Password)
}

// Test 2: With REDIS_PASSWORD env var, should override config value
os.Setenv("REDIS_PASSWORD", "env-password")
if err := loadConfig(); err != nil {
t.Fatalf("Failed to load config: %v", err)
}
if config.Redis.Password != "env-password" {
t.Errorf("Expected password 'env-password', got '%s'", config.Redis.Password)
}
}
