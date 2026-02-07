package main

import (
"os"
"testing"
)

func TestLoadConfigRedisPasswordEnvOverride(t *testing.T) {
// Save original BUBBLE_CONFIG and REDIS_PASSWORD
originalConfig := os.Getenv("BUBBLE_CONFIG")
originalPassword := os.Getenv("REDIS_PASSWORD")
defer func() {
os.Setenv("BUBBLE_CONFIG", originalConfig)
os.Setenv("REDIS_PASSWORD", originalPassword)
}()

// Create a temporary test config file
testConfig := `redis:
  addr: "localhost:6379"
  password: "config-password"
  list_name: "poppit:notifications"

server:
  port: "8080"

commands:
  - name: "Test Command"
    repo: "test/repo"
    branch: "refs/heads/main"
    type: "manual-trigger"
    dir: "/tmp"
    commands:
      - "echo test"
`
testConfigPath := "/tmp/test_config_bubble.yml"
if err := os.WriteFile(testConfigPath, []byte(testConfig), 0644); err != nil {
t.Fatalf("Failed to create test config: %v", err)
}
defer os.Remove(testConfigPath)

// Set test config
os.Setenv("BUBBLE_CONFIG", testConfigPath)

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
