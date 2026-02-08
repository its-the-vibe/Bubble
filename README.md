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

### Running as a systemd Service

For production deployments on Linux systems, you can run Bubble as a systemd service for automatic startup and management.

#### Prerequisites

- A Linux system with systemd (most modern distributions)
- Bubble binary built and installed
- Redis service running
- A dedicated user account for running Bubble (recommended for security)

#### Installation Steps

1. **Create a dedicated user for Bubble** (recommended):
```bash
sudo useradd -r -s /bin/false bubble
```

2. **Create necessary directories**:
```bash
sudo mkdir -p /opt/bubble
sudo mkdir -p /etc/bubble
sudo mkdir -p /var/log/bubble
```

3. **Build and install the Bubble binary**:
```bash
go build -o bubble .
sudo cp bubble /opt/bubble/
sudo chown bubble:bubble /opt/bubble/bubble
sudo chmod 755 /opt/bubble/bubble
```

4. **Install the configuration file**:
```bash
sudo cp config.yml /etc/bubble/config.yml
sudo chown bubble:bubble /etc/bubble/config.yml
sudo chmod 640 /etc/bubble/config.yml
```

5. **Install the systemd service file**:
```bash
sudo cp contrib/bubble.service /etc/systemd/system/
sudo chmod 644 /etc/systemd/system/bubble.service
```

6. **Configure the service** (optional):

Edit `/etc/systemd/system/bubble.service` if you need to customize:
- Installation paths (default: `/opt/bubble`)
- Configuration file location (default: `/etc/bubble/config.yml`)
- User/group (default: `bubble`)
- Redis password (uncomment and set `REDIS_PASSWORD` environment variable)

7. **Set proper permissions**:
```bash
sudo chown -R bubble:bubble /var/log/bubble
```

8. **Reload systemd and enable the service**:
```bash
sudo systemctl daemon-reload
sudo systemctl enable bubble.service
```

9. **Start the service**:
```bash
sudo systemctl start bubble.service
```

#### Managing the Service

**Check service status**:
```bash
sudo systemctl status bubble.service
```

**View logs**:
```bash
sudo journalctl -u bubble.service -f
```

**Restart the service**:
```bash
sudo systemctl restart bubble.service
```

**Stop the service**:
```bash
sudo systemctl stop bubble.service
```

**Disable automatic startup**:
```bash
sudo systemctl disable bubble.service
```

#### Troubleshooting

If the service fails to start:

1. Check the service status and logs:
```bash
sudo systemctl status bubble.service
sudo journalctl -u bubble.service -n 50
```

2. Verify configuration file exists and is valid:
```bash
sudo -u bubble cat /etc/bubble/config.yml
```

3. Ensure Redis is running:
```bash
sudo systemctl status redis.service
```

4. Check file permissions:
```bash
ls -la /opt/bubble/bubble
ls -la /etc/bubble/config.yml
```

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
- `REDIS_PASSWORD`: Redis password (overrides the password set in config.yml). This is the recommended way to set the Redis password for security reasons.

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
