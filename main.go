package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Redis    RedisConfig     `yaml:"redis"`
	Commands []CommandButton `yaml:"commands"`
	Server   ServerConfig    `yaml:"server"`
}

// RedisConfig holds Redis connection details
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	ListName string `yaml:"list_name"`
}

// CommandButton represents a button in the UI
type CommandButton struct {
	Name     string   `yaml:"name"`
	Repo     string   `yaml:"repo"`
	Branch   string   `yaml:"branch"`
	Type     string   `yaml:"type"`
	Dir      string   `yaml:"dir"`
	Commands []string `yaml:"commands"`
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Port string `yaml:"port"`
}

// PoppitNotification represents the message sent to Poppit
type PoppitNotification struct {
	Repo     string   `json:"repo"`
	Branch   string   `json:"branch"`
	Type     string   `json:"type"`
	Dir      string   `json:"dir"`
	Commands []string `json:"commands"`
}

var (
	config      Config
	redisClient *redis.Client
	ctx         = context.Background()
)

func main() {
	// Load configuration
	if err := loadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Redis client
	redisClient = redis.NewClient(&redis.Options{
		Addr:     config.Redis.Addr,
		Password: config.Redis.Password,
		DB:       0,
	})

	// Test Redis connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis successfully")

	// Setup HTTP handlers
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/execute", handleExecute)

	// Start HTTP server
	server := &http.Server{
		Addr:         ":" + config.Server.Port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Starting Bubble server on port %s", config.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	if err := redisClient.Close(); err != nil {
		log.Printf("Redis close error: %v", err)
	}

	log.Println("Server stopped")
}

func loadConfig() error {
	configPath := os.Getenv("BUBBLE_CONFIG")
	if configPath == "" {
		configPath = "config.yml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if config.Server.Port == "" {
		config.Server.Port = "8080"
	}
	if config.Redis.ListName == "" {
		config.Redis.ListName = "poppit:notifications"
	}

	// Override Redis password from environment variable if set
	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		config.Redis.Password = redisPassword
	}

	return nil
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Bubble - Poppit Frontend</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            padding: 40px;
            max-width: 800px;
            width: 100%;
        }
        h1 {
            color: #333;
            margin-bottom: 10px;
            font-size: 2.5em;
        }
        .subtitle {
            color: #666;
            margin-bottom: 30px;
            font-size: 1.1em;
        }
        .buttons-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
            gap: 15px;
            margin-top: 20px;
        }
        .command-button {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            padding: 20px;
            border-radius: 8px;
            cursor: pointer;
            font-size: 16px;
            font-weight: 600;
            transition: all 0.3s ease;
            box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4);
        }
        .command-button:hover {
            transform: translateY(-2px);
            box-shadow: 0 6px 20px rgba(102, 126, 234, 0.6);
        }
        .command-button:active {
            transform: translateY(0);
        }
        .message {
            margin-top: 20px;
            padding: 15px;
            border-radius: 8px;
            display: none;
        }
        .message.success {
            background-color: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }
        .message.error {
            background-color: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
        }
        .message.show {
            display: block;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🫧 Bubble</h1>
        <p class="subtitle">Web Frontend for Poppit</p>
        
        <div class="buttons-grid">
            {{range .Commands}}
            <button class="command-button" onclick="executeCommand('{{.Name}}')">
                {{.Name}}
            </button>
            {{end}}
        </div>
        
        <div id="message" class="message"></div>
    </div>

    <script>
        function executeCommand(name) {
            const messageDiv = document.getElementById('message');
            
            fetch('/execute', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ name: name })
            })
            .then(response => response.json())
            .then(data => {
                messageDiv.textContent = data.message;
                messageDiv.className = 'message ' + (data.success ? 'success' : 'error') + ' show';
                setTimeout(() => {
                    messageDiv.className = 'message';
                }, 5000);
            })
            .catch(error => {
                messageDiv.textContent = 'Error: ' + error.message;
                messageDiv.className = 'message error show';
                setTimeout(() => {
                    messageDiv.className = 'message';
                }, 5000);
            });
        }
    </script>
</body>
</html>
`

	t := template.Must(template.New("index").Parse(tmpl))
	if err := t.Execute(w, config); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONResponse(w, false, "Invalid request")
		return
	}

	// Find the command button
	var cmdButton *CommandButton
	for i := range config.Commands {
		if config.Commands[i].Name == req.Name {
			cmdButton = &config.Commands[i]
			break
		}
	}

	if cmdButton == nil {
		sendJSONResponse(w, false, "Command not found")
		return
	}

	// Create notification for Poppit
	notification := PoppitNotification{
		Repo:     cmdButton.Repo,
		Branch:   cmdButton.Branch,
		Type:     cmdButton.Type,
		Dir:      cmdButton.Dir,
		Commands: cmdButton.Commands,
	}

	// Convert to JSON
	notificationJSON, err := json.Marshal(notification)
	if err != nil {
		sendJSONResponse(w, false, "Failed to create notification")
		return
	}

	// Push to Redis list
	if err := redisClient.RPush(ctx, config.Redis.ListName, string(notificationJSON)).Err(); err != nil {
		log.Printf("Redis error: %v", err)
		sendJSONResponse(w, false, "Failed to send command to Poppit")
		return
	}

	log.Printf("Command '%s' sent to Poppit successfully", req.Name)
	sendJSONResponse(w, true, fmt.Sprintf("Command '%s' sent to Poppit successfully!", req.Name))
}

func sendJSONResponse(w http.ResponseWriter, success bool, message string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": success,
		"message": message,
	})
}
