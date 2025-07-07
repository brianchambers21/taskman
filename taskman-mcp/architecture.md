# Taskman MCP Server Architecture

## Overview

The Taskman MCP (Model Context Protocol) Server provides a comprehensive interface for task and project management through Claude Desktop. It acts as a bridge between Claude and the Taskman REST API, offering tools, resources, and prompts for efficient task management workflows.

## Architecture Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Claude        │◄──►│  Taskman MCP    │◄──►│  Taskman API    │
│   Desktop       │    │    Server       │    │   (REST)        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   PostgreSQL    │
                       │   Database      │
                       └─────────────────┘
```

## Transport & Communication

- **Protocol**: MCP (Model Context Protocol)
- **Transport**: stdio (standard input/output)
- **Data Format**: JSON-RPC 2.0
- **Configuration**: Via environment variables and `claude_desktop_config.json`

## Core Components

### 1. Server (`internal/server/server.go`)
- **Main MCP server** implementing the MCP protocol
- **Transport management** (stdio/HTTP)
- **Tool registration** and middleware setup
- **Comprehensive request/response logging**
- **Keep-alive and ping handling**

### 2. API Client (`internal/client/api_client.go`)
- **HTTP client** for Taskman REST API communication
- **Request/response handling** with structured logging
- **Error handling** and timeout management
- **JSON serialization/deserialization**

### 3. Configuration (`internal/config/config.go`)
- **Environment-based configuration**
- **API connection settings**
- **Transport mode selection** (stdio/HTTP/both)
- **Logging level configuration**

## MCP Tools (12 Total)

### Core System Tools

#### `health_check`
- **Purpose**: Check API server health and connectivity
- **Parameters**: None
- **Returns**: API health status and service information

### Task Management Tools

#### `get_task_overview`
- **Purpose**: Dashboard overview with task statistics and insights
- **Parameters**: Optional filters (status, assignee, project)
- **Returns**: Task breakdown, overdue tasks, and recent activity

#### `get_all_tasks`
- **Purpose**: Complete list of all tasks with comprehensive analysis
- **Parameters**: None
- **Returns**: All tasks with status/priority breakdowns, overdue analysis, and insights

#### `create_task_with_context`
- **Purpose**: Create new tasks with initial notes and context
- **Parameters**: task_name, task_description, status, priority, assigned_to, project_id, due_date, initial_note, created_by
- **Returns**: Created task details and suggested next steps

#### `get_task_details`
- **Purpose**: Complete task information including notes and project context
- **Parameters**: task_id
- **Returns**: Full task details, notes, project info, and insights

#### `update_task_progress`
- **Purpose**: Update task status with progress notes
- **Parameters**: task_id, new_status, progress_note, updated_by
- **Returns**: Updated task and change summary

#### `search_tasks`
- **Purpose**: Advanced task search with multiple filters
- **Parameters**: status, priority, assigned_to, project_id, created_by, due_date_from, due_date_to, search_text, archived, sort_by, sort_order, limit
- **Returns**: Filtered tasks with search insights and suggestions

#### `add_task_note`
- **Purpose**: Add a note/comment to an existing task without requiring status changes
- **Parameters**: task_id, note, created_by
- **Returns**: Created note details and task context

### Project Management Tools

#### `get_project_status`
- **Purpose**: Project overview with task breakdown and progress metrics
- **Parameters**: project_id
- **Returns**: Project details, task statistics, and progress insights

#### `get_all_projects`
- **Purpose**: Complete list of all projects in the system
- **Parameters**: None
- **Returns**: All projects with creation details and metadata

#### `create_project_with_initial_tasks`
- **Purpose**: Create project and populate with initial tasks
- **Parameters**: project_name, project_description, created_by, initial_tasks[]
- **Returns**: Created project and task creation results

### User-Focused Tools

#### `get_my_work`
- **Purpose**: Personalized work queue with prioritized tasks
- **Parameters**: user_id, priority_filter, status_filter, limit
- **Returns**: User's assigned tasks with workload insights

## MCP Resources (10 Total)

### Dashboard Resources
- `taskman://dashboard/overview` - System-wide dashboard
- `taskman://dashboard/metrics` - Performance metrics and KPIs

### Task Resources
- `taskman://tasks/overview` - Task statistics and breakdowns
- `taskman://tasks/{task_id}` - Individual task details
- `taskman://tasks/{task_id}/notes` - Task notes and history
- `taskman://tasks/{task_id}/project` - Task's project context

### Project Resources
- `taskman://projects/overview` - All projects with statistics
- `taskman://projects/{project_id}` - Individual project details
- `taskman://projects/{project_id}/tasks` - Project's task breakdown
- `taskman://projects/{project_id}/metrics` - Project performance metrics

## MCP Prompts (11 Total)

### Task Management Prompts
- `create_task` - Template for creating new tasks
- `plan_task` - Comprehensive task planning guide
- `task_breakdown` - Break complex tasks into subtasks
- `update_task_status` - Task status update template
- `task_review` - Task completion review template
- `task_handoff` - Task transfer between team members

### Project Management Prompts
- `create_project_plan` - Comprehensive project planning
- `project_status_review` - Regular project health checks
- `project_retrospective` - Post-project analysis template

### Workflow Prompts
- `daily_standup` - Daily work summaries and planning
- `weekly_planning` - Weekly priority and capacity planning

## Configuration

### Environment Variables
```bash
TASKMAN_API_BASE_URL=http://localhost:8080    # API endpoint
TASKMAN_MCP_TRANSPORT=stdio                   # Transport mode
TASKMAN_LOG_LEVEL=INFO                        # Logging level
TASKMAN_API_TIMEOUT=30s                       # API request timeout
TASKMAN_MCP_SERVER_NAME=taskman-mcp          # Server name
TASKMAN_MCP_SERVER_VERSION=1.0.0             # Server version
```

### Claude Desktop Configuration
```json
{
  "mcpServers": {
    "taskman": {
      "command": "/path/to/taskman-mcp/bin/taskman-mcp-wrapper.sh",
      "env": {
        "TASKMAN_API_BASE_URL": "http://localhost:8080"
      }
    }
  }
}
```

## Logging & Monitoring

### Request/Response Logging
- **Comprehensive middleware** logging all MCP requests/responses
- **Structured logging** with timestamps and session information
- **Performance metrics** with request duration tracking
- **Error tracking** with detailed error information

### Log Levels
- **DEBUG**: Detailed request/response bodies and API calls
- **INFO**: Request summaries and completion status
- **WARN**: Non-critical issues and warnings
- **ERROR**: Failed requests and system errors

### Log Output
- **stderr redirection** to timestamped log files
- **Real-time monitoring** capabilities
- **Log rotation** with timestamped files

## Development & Testing

### Build System
- **Makefile** with comprehensive build targets
- **Development builds** (`make build-dev`)
- **Production builds** (`make build`)
- **Docker support** (`make build-docker`)

### Testing
- **Unit tests** for all components
- **Integration tests** for API connectivity
- **MCP compliance tests** for protocol adherence
- **Tool-specific tests** for functionality validation

### Code Quality
- **Go formatting** (`make fmt`)
- **Linting** (`make lint`)
- **Vet checks** (`make vet`)
- **Security scanning** (`make security-scan`)

## Deployment

### Local Development
1. Build: `make build-dev`
2. Configure Claude Desktop with appropriate config
3. Start Taskman API server
4. Launch Claude Desktop

### Production Deployment
1. Build: `make build`
2. Deploy binary to target environment
3. Configure environment variables
4. Set up monitoring and logging
5. Deploy via scripts: `make deploy-staging` / `make deploy-prod`

## Error Handling

### API Error Handling
- **HTTP status code** interpretation
- **Structured error responses** from Taskman API
- **Graceful degradation** for non-critical failures
- **Retry logic** for transient failures

### MCP Error Handling
- **Protocol-compliant** error responses
- **Detailed error context** in responses
- **Logging** of all error conditions
- **Client-friendly** error messages

## Security Considerations

### Authentication
- **No built-in authentication** (relies on API server)
- **Environment-based** configuration
- **No credential storage** in MCP server

### Data Handling
- **No data persistence** in MCP server
- **Pass-through** for all data operations
- **Logging controls** for sensitive information
- **API proxy** pattern for security isolation

## Performance Characteristics

### Response Times
- **Health check**: < 50ms
- **Simple queries**: < 200ms
- **Complex searches**: < 500ms
- **Data creation**: < 1s

### Concurrency
- **Single session** per Claude Desktop instance
- **Concurrent API requests** supported
- **Connection pooling** for API client
- **Graceful shutdown** handling

### Resource Usage
- **Memory footprint**: ~10-20MB
- **CPU usage**: Minimal (I/O bound)
- **Network**: HTTP requests to API only
- **Disk**: Log files only

## Future Enhancements

### Planned Features
- **Caching layer** for improved performance
- **Bulk operations** for efficiency
- **Real-time notifications** via WebSocket
- **Advanced analytics** and reporting
- **Team collaboration** features

### Protocol Extensions
- **File attachment** support
- **Rich media** handling
- **Custom resource types**
- **Extended prompt capabilities**