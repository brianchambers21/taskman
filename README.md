# Task Management System

A task management system built with Go, PostgreSQL, and Model Context Protocol (MCP).

## Action Items 
[ ] No authN / authZ currently


## Architecture

- **REST API**: Golang-based REST API for task management
- **Database**: PostgreSQL for persistent storage
- **MCP Server**: Model Context Protocol server for AI integration

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- PostgreSQL (via Docker)

## Project Structure

```
taskman/
├── taskman-api/     # REST API server
└── taskman-mcp/     # MCP server
```

## Setup Instructions

### 1. Clone the Repository

```bash
git clone <repository-url>
cd taskman
```

### 2. Environment Configuration

Copy the example environment file and configure your settings:

```bash
cp .env.example .env
```

Edit `.env` and update the following variables:
- `TASKMAN_DB_PASSWORD`: Set a secure password for the database
- Other variables can be left as defaults for local development

### 3. Start PostgreSQL Database

```bash
docker-compose up -d
```

This will start PostgreSQL with the configured environment variables.

Note: The default port is 5433 to avoid conflicts with existing PostgreSQL installations.

### 4. Initialize Go Modules

```bash
# Initialize API module
cd taskman-api
go mod download

# Initialize MCP module
cd ../taskman-mcp
go mod download
```

### 5. Run Database Migrations

```bash
cd taskman-api
go run cmd/migrate/main.go up
```

### 6. Start the API Server

```bash
cd taskman-api
go run cmd/api/main.go
```

The API server will start on `http://localhost:8080` (or the port specified in TASKMAN_API_PORT).

### 7. Start the MCP Server

In a new terminal:

```bash
cd taskman-mcp
go run cmd/server/main.go
```

## API Endpoints

### Task Management
- `GET /api/v1/tasks` - List all tasks
- `POST /api/v1/tasks` - Create a new task
- `GET /api/v1/tasks/{task_id}` - Get task details
- `PUT /api/v1/tasks/{task_id}` - Update a task
- `DELETE /api/v1/tasks/{task_id}` - Delete a task

### Project Management
- `GET /api/v1/projects` - List all projects
- `POST /api/v1/projects` - Create a new project
- `GET /api/v1/projects/{project_id}` - Get project details
- `PUT /api/v1/projects/{project_id}` - Update a project
- `DELETE /api/v1/projects/{project_id}` - Delete a project
- `GET /api/v1/projects/{project_id}/tasks` - Get tasks in a project

### Task Notes
- `GET /api/v1/tasks/{task_id}/notes` - Get task notes
- `POST /api/v1/tasks/{task_id}/notes` - Add a note to a task
- `PUT /api/v1/tasks/{task_id}/notes/{note_id}` - Update a note
- `DELETE /api/v1/tasks/{task_id}/notes/{note_id}` - Delete a note

## Development

### Running Tests

```bash
# Run API tests
cd taskman-api
go test ./...

# Run MCP tests
cd taskman-mcp
go test ./...
```

### Building for Production

```bash
# Build API server
cd taskman-api
go build -o bin/api cmd/api/main.go

# Build MCP server
cd taskman-mcp
go build -o bin/mcp cmd/server/main.go
```

## License

[Your License Here]
