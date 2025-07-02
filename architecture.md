## Tech Stack

- REST API implemented in Golang
- PostgreSQL database for tasks 
- Model Context Protocol (MCP) server 
- Use structured logging for all inputs and outputs to functions across the code base.

## Coding Guardrails 

- Demonstrate an extreme bias for idiomatic Golang code.
- Write minimal code at all times.
- All functions must have unit tests.
- All functionality must have integration tests. 
- Unit and Integration tests must pass before the task is complete.

## Project Structure 

```
taskman/
├── README.md
└── .env.example

# REST API module (interacts with database)
taskman-api/
├── go.mod
├── go.sum
├── main.go
├── cmd/
│   ├── api/
│   └── migrate/
├── internal/
│   ├── handlers/
│   │   └── middleware/
│   ├── routes/
│   ├── server/
│   ├── models/
│   ├── repository/
│   │   ├── postgres/
│   │   │   └── migrations/
│   │   └── interfaces.go
│   ├── service/
│   └── config/
├── api/
│   └── openapi.yaml
└── scripts/

# MCP Server module (interacts with API via HTTP)
taskman-mcp/
├── go.mod
├── go.sum
├── main.go
├── cmd/
│   └── server/
├── internal/
│   ├── server/
│   │   ├── server.go
│   │   └── handlers.go
│   ├── tools/
│   │   ├── task_tools.go
│   │   ├── project_tools.go
│   │   └── note_tools.go
│   ├── resources/
│   │   ├── task_resources.go
│   │   ├── project_resources.go
│   │   └── dashboard_resources.go
│   ├── prompts/
│   │   ├── task_prompts.go
│   │   ├── project_prompts.go
│   │   └── workflow_prompts.go
│   ├── client/
│   │   └── api_client.go
│   └── config/
├── pkg/
│   └── types/
├── templates/
│   └── prompts/
└── scripts/
```

# System Design 
## Database 

- Use exactly the provided database schema. 
- Database credentials must leverage environment variables named `$TASKMAN_`.

```
-- Task Management System Database Schema

-- Projects table for task organization
CREATE TABLE projects (
    project_id VARCHAR(255) PRIMARY KEY,
    project_name VARCHAR(255) NOT NULL,
    project_description TEXT,
    created_by VARCHAR(255) NOT NULL,
    creation_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Main tasks table
CREATE TABLE tasks (
    task_id VARCHAR(255) PRIMARY KEY,
    task_name VARCHAR(255) NOT NULL,
    task_description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'Not Started',
    priority VARCHAR(20) DEFAULT 'Medium',
    assigned_to VARCHAR(255),
    project_id VARCHAR(255),
    due_date TIMESTAMP,
    start_date TIMESTAMP,
    completion_date TIMESTAMP,
    tags TEXT[], -- Array field for tags (PostgreSQL syntax)
    archived BOOLEAN DEFAULT FALSE,
    created_by VARCHAR(255) NOT NULL,
    creation_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_updated_by VARCHAR(255),
    last_update_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraints
    CONSTRAINT fk_tasks_project FOREIGN KEY (project_id) REFERENCES projects(project_id)
);

-- Task notes table for comments/updates
CREATE TABLE task_notes (
    note_id VARCHAR(255) PRIMARY KEY,
    task_id VARCHAR(255) NOT NULL,
    note TEXT NOT NULL,
    created_by VARCHAR(255) NOT NULL,
    creation_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_updated_by VARCHAR(255),
    last_update_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraint
    CONSTRAINT fk_task_notes_task FOREIGN KEY (task_id) REFERENCES tasks(task_id) ON DELETE CASCADE
);

-- Indexes for better query performance
CREATE INDEX idx_tasks_assigned_to ON tasks(assigned_to);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_project_id ON tasks(project_id);
CREATE INDEX idx_tasks_due_date ON tasks(due_date);
CREATE INDEX idx_tasks_archived ON tasks(archived);
CREATE INDEX idx_task_notes_task_id ON task_notes(task_id);

-- Optional: Add check constraints for data validation
ALTER TABLE tasks ADD CONSTRAINT chk_status 
    CHECK (status IN ('Not Started', 'In Progress', 'Blocked', 'Review', 'Complete'));

ALTER TABLE tasks ADD CONSTRAINT chk_priority 
    CHECK (priority IN ('Low', 'Medium', 'High'));

-- Optional: Trigger to automatically update last_update_date
CREATE OR REPLACE FUNCTION update_last_update_date()
RETURNS TRIGGER AS $$
BEGIN
    NEW.last_update_date = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_tasks_update_date
    BEFORE UPDATE ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION update_last_update_date();

CREATE TRIGGER trigger_task_notes_update_date
    BEFORE UPDATE ON task_notes
    FOR EACH ROW
    EXECUTE FUNCTION update_last_update_date();
```

## API

- The following API endpoints must be created using RESTful conventions.
- Use the golang HTTP server (not an external library)
- No authentication to start (TODO later)
- Use standard HTTP error codes for the API implementation
- Log all external calls (database) and all inputs and outputs to functions


```
**Basic Task Management:**

- `GET /api/v1/tasks` - List tasks (with filtering for status, assignee, project, etc.)
- `POST /api/v1/tasks` - Create new task
- `GET /api/v1/tasks/{task_id}` - Get task details
- `PUT /api/v1/tasks/{task_id}` - Update task (status, assignee, description, etc.)
- `DELETE /api/v1/tasks/{task_id}` - Delete task

**Project Management:**

- `GET /api/v1/projects` - List projects
- `POST /api/v1/projects` - Create project
- `GET /api/v1/projects/{project_id}` - Get project details
- `PUT /api/v1/projects/{project_id}` - Update project
- `DELETE /api/v1/projects/{project_id}` - Delete project
- `GET /api/v1/projects/{project_id}/tasks` - Get tasks in project

**Task Notes/Comments:**

- `GET /api/v1/tasks/{task_id}/notes` - Get task notes
- `POST /api/v1/tasks/{task_id}/notes` - Add note to task
- `PUT /api/v1/tasks/{task_id}/notes/{note_id}` - Update note
- `DELETE /api/v1/tasks/{task_id}/notes/{note_id}` - Delete note
```


## MCP

- Use the go MCP SDK available at https://github.com/modelcontextprotocol/go-sdk 
- Do NOT break the MCP specification under any circumstance. The details of the protocol are available here: https://github.com/modelcontextprotocol/modelcontextprotocol
- Trust the specification over the MCP GO SDK since it is early version

# ## Recommended MCP Tools

#### **Task Management Workflows**

```
get_task_overview
- GET /api/v1/tasks (with smart filtering)
- GET /api/v1/projects (for context)
- Returns: Dashboard view with task counts by status, overdue tasks, recent activity

create_task_with_context
- POST /api/v1/tasks
- POST /api/v1/tasks/{task_id}/notes (initial planning note)
- Returns: Created task with confirmation and next steps

get_task_details
- GET /api/v1/tasks/{task_id}
- GET /api/v1/tasks/{task_id}/notes
- GET /api/v1/projects/{project_id} (if task has project)
- Returns: Complete task context for decision-making

update_task_progress
- PUT /api/v1/tasks/{task_id} (status/progress updates)
- POST /api/v1/tasks/{task_id}/notes (progress note)
- Returns: Updated task with change summary
```

#### **Project-Focused Tools**

```
get_project_status
- GET /api/v1/projects/{project_id}
- GET /api/v1/projects/{project_id}/tasks
- Returns: Project overview with task breakdown, progress metrics

create_project_with_initial_tasks
- POST /api/v1/projects
- POST /api/v1/tasks (multiple tasks)
- Returns: Project created with task summary
```

#### **Query & Analysis Tools**

```
search_tasks
- GET /api/v1/tasks (with complex filtering)
- Smart parameter handling for status, assignee, project, date ranges
- Returns: Filtered results with summary statistics

get_my_work
- GET /api/v1/tasks?assigned_to={current_user}&status=In Progress
- GET /api/v1/tasks?assigned_to={current_user}&status=Review
- Returns: Personalized work queue with priorities
```

### Key Design Principles:

**1. Contextual Information:** Each tool provides enough context for the AI to make informed decisions without additional calls.
**2. Workflow-Oriented:** Tools match how humans actually think about task management (not database operations).
**3. Smart Defaults:** Tools handle common filtering and sorting automatically.
**4. Actionable Responses:** Each tool returns not just data, but insights and suggested next actions.
**5. Batch Operations:** Where sensible, tools combine multiple API calls to reduce back-and-forth.

## Development Order

- Set up environment variables for the database 
- Create and configure database using Docker compose 
- Create one task per group of API endpoints including unit and integration tests
- Complete all API endpoints first
- Create MCP server methods