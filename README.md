# Bubble
Web front-end for [Poppit](https://github.com/its-the-vibe/Poppit)

Bubble provides a simple web interface that displays command buttons. When clicked, these buttons send notifications to Poppit via Redis for execution.

## Features

- Clean web UI with configurable command buttons
- Sends notifications to Poppit via Redis
- Configurable via YAML file
- Docker support with readonly container
- Built with Go 1.24

## Quick Start

### Prerequisites

- Go 1.24 or later (for building from source)
- Redis server running and accessible
- [Poppit](https://github.com/its-the-vibe/Poppit) service running

### Configuration

1. Copy the example configuration file:
```bash
cp config.example.yml config.yml
```

2. Edit `config.yml` to configure your commands and Redis connection:
```yaml
redis:
  addr: "localhost:6379"
  password: ""
  list_name: "poppit:notifications"

server:
  port: "8080"

commands:
  - name: "Deploy Production"
    repo: "example/repo"
    branch: "refs/heads/main"
    type: "manual-trigger"
    dir: "/opt/deployments/production"
    commands:
      - "git pull origin main"
      - "./deploy.sh production"
```

### Running with Go

1. Build the application:
```bash
go build -o bubble .
```

2. Run with default configuration (config.yml):
```bash
./bubble
```

3. Run with custom configuration:
```bash
BUBBLE_CONFIG=/path/to/config.yml ./bubble
```

The web interface will be available at http://localhost:8080

### Running with Docker

1. Create your `config.yml` file (see Configuration section above)

2. Build the Docker image:
```bash
docker compose build
```

3. Run the container:
```bash
docker compose up
```

The container runs in read-only mode for security.

### Docker Image

The Dockerfile uses a multi-stage build:
- Build stage: golang:1.24-alpine
- Runtime stage: scratch (minimal image)

This results in a small, secure container image.

## How It Works

1. User accesses the web interface at http://localhost:8080
2. The UI displays buttons for each configured command
3. When a button is clicked:
   - The command details are sent to the backend
   - A JSON notification is pushed to the Redis list configured in `config.yml`
   - Poppit monitors this Redis list and executes the commands

## Environment Variables

- `BUBBLE_CONFIG`: Path to the configuration file (default: `config.yml`)

## Development

### Building

```bash
go build -o bubble .
```

### Testing Locally

You'll need a Redis server running. You can use Docker:
```bash
docker run -d -p 6379:6379 redis:latest
```

Then start Bubble:
```bash
./bubble
```

## Security

- The Docker container runs in read-only mode (`read_only: true` in docker-compose.yml)
- The scratch base image provides minimal attack surface
- Configuration file should not be committed to source control (excluded in .gitignore)
