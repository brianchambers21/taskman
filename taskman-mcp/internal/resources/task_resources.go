package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/bchamber/taskman-mcp/internal/client"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TaskResources handles task-related MCP resources
type TaskResources struct {
	apiClient *client.APIClient
}

// NewTaskResources creates a new task resources handler
func NewTaskResources(apiClient *client.APIClient) *TaskResources {
	return &TaskResources{
		apiClient: apiClient,
	}
}

// Task represents a task from the API
type Task struct {
	TaskID          string   `json:"task_id"`
	TaskName        string   `json:"task_name"`
	TaskDescription *string  `json:"task_description"`
	Status          string   `json:"status"`
	Priority        *string  `json:"priority"`
	AssignedTo      *string  `json:"assigned_to"`
	ProjectID       *string  `json:"project_id"`
	DueDate         *string  `json:"due_date"`
	StartDate       *string  `json:"start_date"`
	CompletionDate  *string  `json:"completion_date"`
	Tags            []string `json:"tags"`
	Archived        bool     `json:"archived"`
	CreatedBy       string   `json:"created_by"`
	CreationDate    string   `json:"creation_date"`
	LastUpdatedBy   *string  `json:"last_updated_by"`
	LastUpdateDate  *string  `json:"last_update_date"`
}

// TaskNote represents a task note from the API
type TaskNote struct {
	NoteID         string  `json:"note_id"`
	TaskID         string  `json:"task_id"`
	Note           string  `json:"note"`
	CreatedBy      string  `json:"created_by"`
	CreationDate   string  `json:"creation_date"`
	LastUpdatedBy  *string `json:"last_updated_by"`
	LastUpdateDate *string `json:"last_update_date"`
}

// Project represents a project from the API
type Project struct {
	ProjectID          string  `json:"project_id"`
	ProjectName        string  `json:"project_name"`
	ProjectDescription *string `json:"project_description"`
	CreatedBy          string  `json:"created_by"`
	CreationDate       string  `json:"creation_date"`
}

// HandleTaskResource handles individual task resource requests
func (tr *TaskResources) HandleTaskResource(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.ReadResourceParams,
) (*mcp.ReadResourceResult, error) {
	slog.Info("Reading task resource", "uri", params.URI)

	// Extract task ID from URI: taskman://task/{task_id}
	parts := strings.Split(params.URI, "/")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid task resource URI format: %s", params.URI)
	}
	taskID := parts[len(parts)-1]

	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}

	// Get task details
	taskResp, err := tr.apiClient.Get(ctx, fmt.Sprintf("/api/v1/tasks/%s", url.PathEscape(taskID)))
	if err != nil {
		slog.Error("Failed to get task", "error", err, "task_id", taskID)
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	var task Task
	if err := json.Unmarshal(taskResp, &task); err != nil {
		slog.Error("Failed to parse task", "error", err)
		return nil, fmt.Errorf("failed to parse task: %w", err)
	}

	// Get task notes
	notesResp, err := tr.apiClient.Get(ctx, fmt.Sprintf("/api/v1/tasks/%s/notes", url.PathEscape(taskID)))
	if err != nil {
		slog.Warn("Failed to get task notes", "error", err, "task_id", taskID)
		// Continue without notes if they can't be retrieved
	}

	var notes []TaskNote
	if notesResp != nil {
		if err := json.Unmarshal(notesResp, &notes); err != nil {
			slog.Warn("Failed to parse task notes", "error", err)
		}
	}

	// Get project details if task has a project
	var project *Project
	if task.ProjectID != nil && *task.ProjectID != "" {
		projectResp, err := tr.apiClient.Get(ctx, fmt.Sprintf("/api/v1/projects/%s", url.PathEscape(*task.ProjectID)))
		if err != nil {
			slog.Warn("Failed to get project", "error", err, "project_id", *task.ProjectID)
		} else {
			var p Project
			if err := json.Unmarshal(projectResp, &p); err != nil {
				slog.Warn("Failed to parse project", "error", err)
			} else {
				project = &p
			}
		}
	}

	// Build formatted response
	response := buildTaskResourceResponse(task, notes, project)

	slog.Info("Task resource retrieved", "task_id", taskID, "note_count", len(notes), "has_project", project != nil)

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      params.URI,
				MIMEType: "text/plain",
				Text:     response,
			},
		},
	}, nil
}

// HandleTasksOverviewResource handles tasks overview resource requests
func (tr *TaskResources) HandleTasksOverviewResource(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.ReadResourceParams,
) (*mcp.ReadResourceResult, error) {
	slog.Info("Reading tasks overview resource", "uri", params.URI)

	// Get all tasks
	tasksResp, err := tr.apiClient.Get(ctx, "/api/v1/tasks")
	if err != nil {
		slog.Error("Failed to get tasks", "error", err)
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	var tasks []Task
	if err := json.Unmarshal(tasksResp, &tasks); err != nil {
		slog.Error("Failed to parse tasks", "error", err)
		return nil, fmt.Errorf("failed to parse tasks: %w", err)
	}

	// Build formatted response
	response := buildTasksOverviewResponse(tasks)

	slog.Info("Tasks overview resource retrieved", "task_count", len(tasks))

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      params.URI,
				MIMEType: "text/plain",
				Text:     response,
			},
		},
	}, nil
}

// HandleUserTasksResource handles user tasks resource requests
func (tr *TaskResources) HandleUserTasksResource(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.ReadResourceParams,
) (*mcp.ReadResourceResult, error) {
	slog.Info("Reading user tasks resource", "uri", params.URI)

	// Extract user ID from URI: taskman://tasks/user/{user_id}
	parts := strings.Split(params.URI, "/")
	if len(parts) != 5 || parts[0] != "taskman:" || parts[1] != "" || parts[2] != "tasks" || parts[3] != "user" {
		return nil, fmt.Errorf("invalid user tasks resource URI format: %s", params.URI)
	}
	userID := parts[4]

	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Get tasks assigned to user
	tasksResp, err := tr.apiClient.Get(ctx, fmt.Sprintf("/api/v1/tasks?assigned_to=%s", url.QueryEscape(userID)))
	if err != nil {
		slog.Error("Failed to get user tasks", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to get user tasks: %w", err)
	}

	var tasks []Task
	if err := json.Unmarshal(tasksResp, &tasks); err != nil {
		slog.Error("Failed to parse user tasks", "error", err)
		return nil, fmt.Errorf("failed to parse user tasks: %w", err)
	}

	// Build formatted response
	response := buildUserTasksResponse(userID, tasks)

	slog.Info("User tasks resource retrieved", "user_id", userID, "task_count", len(tasks))

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      params.URI,
				MIMEType: "text/plain",
				Text:     response,
			},
		},
	}, nil
}

// buildTaskResourceResponse formats individual task data
func buildTaskResourceResponse(task Task, notes []TaskNote, project *Project) string {
	var response strings.Builder

	response.WriteString(fmt.Sprintf("# Task: %s\n\n", task.TaskName))
	response.WriteString(fmt.Sprintf("**ID:** %s\n", task.TaskID))
	response.WriteString(fmt.Sprintf("**Status:** %s\n", task.Status))

	if task.Priority != nil {
		response.WriteString(fmt.Sprintf("**Priority:** %s\n", *task.Priority))
	}

	if task.AssignedTo != nil {
		response.WriteString(fmt.Sprintf("**Assigned To:** %s\n", *task.AssignedTo))
	}

	if task.DueDate != nil {
		response.WriteString(fmt.Sprintf("**Due Date:** %s\n", *task.DueDate))
	}

	response.WriteString(fmt.Sprintf("**Created By:** %s\n", task.CreatedBy))
	response.WriteString(fmt.Sprintf("**Created:** %s\n", task.CreationDate))

	if task.TaskDescription != nil && *task.TaskDescription != "" {
		response.WriteString(fmt.Sprintf("\n**Description:**\n%s\n", *task.TaskDescription))
	}

	if project != nil {
		response.WriteString(fmt.Sprintf("\n**Project:** %s (%s)\n", project.ProjectName, project.ProjectID))
	}

	if len(task.Tags) > 0 {
		response.WriteString(fmt.Sprintf("\n**Tags:** %s\n", strings.Join(task.Tags, ", ")))
	}

	if len(notes) > 0 {
		response.WriteString("\n## Notes\n\n")
		for _, note := range notes {
			response.WriteString(fmt.Sprintf("**%s** (%s):\n%s\n\n", note.CreatedBy, note.CreationDate, note.Note))
		}
	}

	return response.String()
}

// buildTasksOverviewResponse formats tasks overview data
func buildTasksOverviewResponse(tasks []Task) string {
	var response strings.Builder

	response.WriteString("# Tasks Overview\n\n")
	response.WriteString(fmt.Sprintf("**Total Tasks:** %d\n\n", len(tasks)))

	// Status breakdown
	statusCounts := make(map[string]int)
	priorityCounts := make(map[string]int)
	assigneeCounts := make(map[string]int)

	for _, task := range tasks {
		statusCounts[task.Status]++

		if task.Priority != nil {
			priorityCounts[*task.Priority]++
		} else {
			priorityCounts["None"]++
		}

		if task.AssignedTo != nil {
			assigneeCounts[*task.AssignedTo]++
		} else {
			assigneeCounts["Unassigned"]++
		}
	}

	response.WriteString("## Status Breakdown\n")
	for status, count := range statusCounts {
		response.WriteString(fmt.Sprintf("- %s: %d\n", status, count))
	}

	response.WriteString("\n## Priority Breakdown\n")
	for priority, count := range priorityCounts {
		response.WriteString(fmt.Sprintf("- %s: %d\n", priority, count))
	}

	response.WriteString("\n## Assignment Breakdown\n")
	for assignee, count := range assigneeCounts {
		response.WriteString(fmt.Sprintf("- %s: %d\n", assignee, count))
	}

	// Recent tasks
	if len(tasks) > 0 {
		response.WriteString("\n## Recent Tasks\n")
		displayCount := len(tasks)
		if displayCount > 10 {
			displayCount = 10
		}

		for i := 0; i < displayCount; i++ {
			task := tasks[i]
			assignee := "Unassigned"
			if task.AssignedTo != nil {
				assignee = *task.AssignedTo
			}
			response.WriteString(fmt.Sprintf("- **%s** (%s) - %s - %s\n", task.TaskName, task.Status, assignee, task.CreationDate))
		}

		if len(tasks) > 10 {
			response.WriteString(fmt.Sprintf("... and %d more tasks\n", len(tasks)-10))
		}
	}

	return response.String()
}

// buildUserTasksResponse formats user tasks data
func buildUserTasksResponse(userID string, tasks []Task) string {
	var response strings.Builder

	response.WriteString(fmt.Sprintf("# Tasks for %s\n\n", userID))
	response.WriteString(fmt.Sprintf("**Total Tasks:** %d\n\n", len(tasks)))

	if len(tasks) == 0 {
		response.WriteString("No tasks assigned to this user.\n")
		return response.String()
	}

	// Group by status
	tasksByStatus := make(map[string][]Task)
	for _, task := range tasks {
		tasksByStatus[task.Status] = append(tasksByStatus[task.Status], task)
	}

	// Display tasks by status
	statusOrder := []string{"In Progress", "Review", "Blocked", "Not Started", "Complete"}

	for _, status := range statusOrder {
		statusTasks := tasksByStatus[status]
		if len(statusTasks) > 0 {
			response.WriteString(fmt.Sprintf("## %s (%d)\n\n", status, len(statusTasks)))

			for _, task := range statusTasks {
				priority := "None"
				if task.Priority != nil {
					priority = *task.Priority
				}

				dueDate := "No due date"
				if task.DueDate != nil {
					dueDate = *task.DueDate
				}

				response.WriteString(fmt.Sprintf("- **%s** (%s) - Due: %s\n", task.TaskName, priority, dueDate))
				if task.TaskDescription != nil && *task.TaskDescription != "" {
					response.WriteString(fmt.Sprintf("  *%s*\n", *task.TaskDescription))
				}
				response.WriteString("\n")
			}
		}
	}

	return response.String()
}
