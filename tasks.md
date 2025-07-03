# Task Management System - Development Tasks

## Task 1: Project Setup and Environment Configuration

**Description:** Set up the initial project structure and environment configuration files.

**Requirements:**

- Create the exact folder structure specified in architecture.md
- Create `taskman/.env.example` with all required `TASKMAN_` environment variables
- Create `taskman/README.md` with setup instructions
- Create `taskman-api/go.mod` with module name and Go version 1.21+
- Create `taskman-mcp/go.mod` with module name and Go version 1.21+
- Create `docker-compose.yml` in root with PostgreSQL configuration using environment variables

**Success Criteria:**

- All folders exist as specified
- Environment variables follow `TASKMAN_` naming pattern
- Docker Compose starts PostgreSQL successfully
- Go modules initialize without errors

---

## Task 2: Database Schema and Migration System

**Description:** Implement the complete database schema and migration infrastructure.

**Requirements:**

- Create migration files in `taskman-api/internal/repository/postgres/migrations/`
- Implement the exact database schema from architecture.md (projects, tasks, task_notes tables)
- Create database connection logic using `lib/pq` driver
- Implement migration runner in `taskman-api/cmd/migrate/`
- Use `TASKMAN_` environment variables for database credentials
- Add structured logging for all database operations

**Success Criteria:**

- All tables created with correct schema, constraints, and indexes
- Migration system can run up/down migrations
- Database connection works with environment variables
- All database operations are logged

---

## Task 3: Core Models and Repository Interfaces

**Description:** Create Go structs for data models and repository interface definitions.

**Requirements:**

- Create models in `taskman-api/internal/models/` matching database schema exactly
- Use proper JSON and database tags
- Create repository interfaces in `taskman-api/internal/repository/interfaces.go`
- Implement PostgreSQL repository in `taskman-api/internal/repository/postgres/`
- Add structured logging for all function inputs and outputs
- Include proper error handling

**Success Criteria:**

- Models match database schema exactly
- Repository interfaces define all CRUD operations
- PostgreSQL implementation works with real database
- All functions have comprehensive unit tests
- Integration tests pass with test database

---

## Task 4: Basic Task Management API Endpoints

**Description:** Implement the 5 basic task management REST endpoints.

**Requirements:**

- `GET /api/v1/tasks` - List tasks with filtering (status, assignee, project, etc.)
- `POST /api/v1/tasks` - Create new task
- `GET /api/v1/tasks/{task_id}` - Get task details
- `PUT /api/v1/tasks/{task_id}` - Update task
- `DELETE /api/v1/tasks/{task_id}` - Delete task
- Use only Go standard library HTTP server (`net/http`)
- Implement proper HTTP status codes
- Add structured logging for all requests/responses
- Create service layer for business logic

**Success Criteria:**

- All 5 endpoints work correctly
- Proper HTTP status codes returned
- Request/response logging implemented
- Unit tests for all handlers and services
- Integration tests for all endpoints with real database

---

## Task 5: Project Management API Endpoints

**Description:** Implement the 6 project management REST endpoints.

**Requirements:**

- `GET /api/v1/projects` - List projects
- `POST /api/v1/projects` - Create project
- `GET /api/v1/projects/{project_id}` - Get project details
- `PUT /api/v1/projects/{project_id}` - Update project
- `DELETE /api/v1/projects/{project_id}` - Delete project
- `GET /api/v1/projects/{project_id}/tasks` - Get tasks in project
- Use same patterns as Task 4
- Add structured logging for all operations

**Success Criteria:**

- All 6 endpoints work correctly
- Proper HTTP status codes and error handling
- Request/response logging implemented
- Unit tests for all handlers and services
- Integration tests for all endpoints with real database

---

## Task 6: Task Notes API Endpoints

**Description:** Implement the 4 task notes REST endpoints.

**Requirements:**

- `GET /api/v1/tasks/{task_id}/notes` - Get task notes
- `POST /api/v1/tasks/{task_id}/notes` - Add note to task
- `PUT /api/v1/tasks/{task_id}/notes/{note_id}` - Update note
- `DELETE /api/v1/tasks/{task_id}/notes/{note_id}` - Delete note
- Use same patterns as previous tasks
- Add structured logging for all operations

**Success Criteria:**

- All 4 endpoints work correctly
- Proper HTTP status codes and error handling
- Request/response logging implemented
- Unit tests for all handlers and services
- Integration tests for all endpoints with real database

---

## Task 7: HTTP Server and Routing Infrastructure

**Description:** Create the main HTTP server, routing, and middleware infrastructure.

**Requirements:**

- Implement HTTP server in `taskman-api/internal/server/`
- Create routing logic in `taskman-api/internal/routes/`
- Add middleware for logging, CORS, and error handling
- Create main application entry point in `taskman-api/cmd/api/`
- Use only Go standard library (`net/http`, `http.ServeMux`)
- Add graceful shutdown handling

**Success Criteria:**

- HTTP server starts and serves all endpoints
- Middleware functions correctly
- Graceful shutdown works
- All routes properly registered
- Integration tests pass for complete API

---

## Task 8: MCP Server Infrastructure

**Description:** Set up the MCP server foundation using the official Go SDK.

**Requirements:**

- Set up MCP server in `taskman-mcp/internal/server/`
- Use Go SDK from https://github.com/modelcontextprotocol/go-sdk
- Implement stdio transport for MCP communication
- Create HTTP client in `taskman-mcp/internal/client/` to communicate with REST API
- Add configuration management for API endpoint URLs
- Follow MCP specification exactly

**Success Criteria:**

- MCP server initializes and responds to initialization requests
- HTTP client can successfully call REST API endpoints
- MCP protocol compliance verified
- Basic server structure in place
- Unit tests for server initialization and HTTP client

---

## Task 9: MCP Task Management Tools

**Description:** Implement the 4 task management MCP tools.

**Requirements:**

- Implement `get_task_overview` tool combining GET /tasks and GET /projects
- Implement `create_task_with_context` tool (POST /tasks + POST /tasks/{id}/notes)
- Implement `get_task_details` tool (GET /tasks/{id} + GET /notes + GET /projects/{id})
- Implement `update_task_progress` tool (PUT /tasks/{id} + POST /notes)
- Each tool should make multiple API calls as specified in architecture.md
- Return structured, contextual responses
- Add comprehensive error handling and logging

**Success Criteria:**

- All 4 tools work correctly via MCP protocol
- Tools make appropriate API calls as specified
- Responses include contextual information and insights
- Unit tests for each tool
- Integration tests with real API server

---

## Task 10: MCP Project and Query Tools

**Description:** Implement the remaining 4 MCP tools for projects and querying.

**Requirements:**

- Implement `get_project_status` tool (GET /projects/{id} + GET /projects/{id}/tasks)
- Implement `create_project_with_initial_tasks` tool (POST /projects + multiple POST /tasks)
- Implement `search_tasks` tool (GET /tasks with complex filtering)
- Implement `get_my_work` tool (GET /tasks with specific filters)
- Follow same patterns as Task 9
- Add comprehensive error handling and logging

**Success Criteria:**

- All 4 tools work correctly via MCP protocol
- Tools provide workflow-oriented responses as specified
- Smart filtering and defaults implemented
- Unit tests for each tool
- Integration tests with real API server

---

## Task 11: MCP Resources Implementation

**Description:** Implement MCP resources for read-only data access.

**Requirements:**

- Implement task resources in `taskman-mcp/internal/resources/task_resources.go`
- Implement project resources in `taskman-mcp/internal/resources/project_resources.go`
- Implement dashboard resources in `taskman-mcp/internal/resources/dashboard_resources.go`
- Resources should provide read-only access to data via API calls
- Follow MCP specification for resource implementation
- Add structured logging

**Success Criteria:**

- All resource types work correctly
- Resources return appropriate data from API
- MCP resource protocol compliance verified
- Unit tests for all resources
- Integration tests with real API server

---

## Task 12: MCP Prompts Implementation

**Description:** Implement MCP prompts for workflow templates.

**Requirements:**

- Create prompt templates in `taskman-mcp/templates/prompts/`
- Implement task prompts in `taskman-mcp/internal/prompts/task_prompts.go`
- Implement project prompts in `taskman-mcp/internal/prompts/project_prompts.go`
- Implement workflow prompts in `taskman-mcp/internal/prompts/workflow_prompts.go`
- Templates should guide users through common task management workflows
- Follow MCP specification for prompt implementation

**Success Criteria:**

- All prompt types work correctly
- Templates provide useful workflow guidance
- MCP prompt protocol compliance verified
- Unit tests for all prompt handlers
- Templates are easily customizable

---

## Task 13: Final Integration and End-to-End Testing

**Description:** Complete system integration and comprehensive testing.

**Requirements:**

- Create end-to-end tests that verify complete workflows
- Test MCP server communication with REST API
- Verify all MCP tools work with real data
- Add comprehensive error handling throughout system
- Create deployment scripts and documentation
- Verify all unit and integration tests pass

**Success Criteria:**

- Complete system works end-to-end
- All MCP tools function correctly with REST API
- All tests pass (unit, integration, end-to-end)
- System can be deployed and run successfully
- Documentation is complete and accurate


## Task 14: MCP client 

**Description:** Client of MCP server that is able to take an intent from a model and execute the appropriate MCP calls (list_tools, execute_tool())

**Requirements:** 

- Create MCP client in separate folder from the root of the project named `mcp-client`
- Integration tests must be present to test the features of the client including listing tools and executing tools, as well as fetching prompts. 
- The client should expect an input message that is compatible with the MCP specification. 
- The client should use HTTP for transport since we are interacting with a web-deployed MCP server.
- The client must fully comply with the MCP specification 
- The client should use the MCP golang SDK
- The client should be executable by the end user for testing purposes. 