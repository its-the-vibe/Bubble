package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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
	testConfigPath := t.TempDir() + "/config.yml"
	if err := os.WriteFile(testConfigPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

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

func TestLoadConfigDefaults(t *testing.T) {
	originalConfig := os.Getenv("BUBBLE_CONFIG")
	defer os.Setenv("BUBBLE_CONFIG", originalConfig)

	testConfig := `redis:
  addr: "localhost:6379"
`
	testConfigPath := t.TempDir() + "/config.yml"
	if err := os.WriteFile(testConfigPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	os.Setenv("BUBBLE_CONFIG", testConfigPath)
	os.Unsetenv("REDIS_PASSWORD")

	if err := loadConfig(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.Server.Port != "8080" {
		t.Errorf("Expected default port '8080', got '%s'", config.Server.Port)
	}
	if config.Redis.ListName != "poppit:notifications" {
		t.Errorf("Expected default list name 'poppit:notifications', got '%s'", config.Redis.ListName)
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	originalConfig := os.Getenv("BUBBLE_CONFIG")
	defer os.Setenv("BUBBLE_CONFIG", originalConfig)

	os.Setenv("BUBBLE_CONFIG", "/tmp/nonexistent_bubble_config.yml")

	if err := loadConfig(); err == nil {
		t.Error("Expected error for missing config file, got nil")
	}
}

func TestSendJSONResponse(t *testing.T) {
	tests := []struct {
		name    string
		success bool
		message string
	}{
		{"success response", true, "Command executed successfully"},
		{"error response", false, "Something went wrong"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			sendJSONResponse(w, tc.success, tc.message)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}
			if ct := w.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Expected Content-Type 'application/json', got '%s'", ct)
			}

			var resp map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}
			if resp["success"] != tc.success {
				t.Errorf("Expected success=%v, got %v", tc.success, resp["success"])
			}
			if resp["message"] != tc.message {
				t.Errorf("Expected message=%q, got %q", tc.message, resp["message"])
			}
		})
	}
}

func TestHandleExecuteMethodNotAllowed(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/execute", nil)
			w := httptest.NewRecorder()
			handleExecute(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected 405 for %s, got %d", method, w.Code)
			}
		})
	}
}

func TestHandleExecuteInvalidJSON(t *testing.T) {
	config = Config{}

	req := httptest.NewRequest(http.MethodPost, "/execute", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handleExecute(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if resp["success"] != false {
		t.Error("Expected success=false for invalid JSON")
	}
}

func TestHandleExecuteCommandNotFound(t *testing.T) {
	config = Config{
		Commands: []CommandButton{
			{Name: "existing-cmd", Repo: "r", Branch: "b", Type: "t"},
		},
	}

	body := `{"name":"unknown-cmd"}`
	req := httptest.NewRequest(http.MethodPost, "/execute", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handleExecute(w, req)

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if resp["success"] != false {
		t.Error("Expected success=false for unknown command")
	}
}

func TestHandleIndex(t *testing.T) {
	config = Config{
		Commands: []CommandButton{
			{Name: "Deploy", Repo: "example/repo", Branch: "refs/heads/main", Type: "manual-trigger"},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handleIndex(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Bubble") {
		t.Error("Expected page to contain 'Bubble'")
	}
	if !strings.Contains(body, "Deploy") {
		t.Error("Expected page to contain configured command name 'Deploy'")
	}
}
