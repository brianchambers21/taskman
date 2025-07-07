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

// ProjectResources handles project-related MCP resources
type ProjectResources struct {
	apiClient *client.APIClient
}

// NewProjectResources creates a new project resources handler
func NewProjectResources(apiClient *client.APIClient) *ProjectResources {
	return &ProjectResources{
		apiClient: apiClient,
	}
}

// HandleProjectResource handles individual project resource requests
func (pr *ProjectResources) HandleProjectResource(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.ReadResourceParams,
) (*mcp.ReadResourceResult, error) {
	slog.Info("Reading project resource", "uri", params.URI)

	// Extract project ID from URI: taskman://project/{project_id}
	parts := strings.Split(params.URI, "/")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid project resource URI format: %s", params.URI)
	}
	projectID := parts[len(parts)-1]

	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	// Get project details
	projectResp, err := pr.apiClient.Get(ctx, fmt.Sprintf("/api/v1/projects/%s", url.PathEscape(projectID)))
	if err != nil {
		slog.Error("Failed to get project", "error", err, "project_id", projectID)
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	var project Project
	if err := json.Unmarshal(projectResp, &project); err != nil {
		slog.Error("Failed to parse project", "error", err)
		return nil, fmt.Errorf("failed to parse project: %w", err)
	}

	// Get project tasks
	tasksResp, err := pr.apiClient.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/tasks", url.PathEscape(projectID)))
	if err != nil {
		slog.Warn("Failed to get project tasks", "error", err, "project_id", projectID)
		// Continue without tasks if they can't be retrieved
	}

	var tasks []Task
	if tasksResp != nil {
		if err := json.Unmarshal(tasksResp, &tasks); err != nil {
			slog.Warn("Failed to parse project tasks", "error", err)
		}
	}

	// Build formatted response
	response := buildProjectResourceResponse(project, tasks)

	slog.Info("Project resource retrieved", "project_id", projectID, "task_count", len(tasks))

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

// HandleProjectsOverviewResource handles projects overview resource requests
func (pr *ProjectResources) HandleProjectsOverviewResource(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.ReadResourceParams,
) (*mcp.ReadResourceResult, error) {
	slog.Info("Reading projects overview resource", "uri", params.URI)

	// Get all projects
	projectsResp, err := pr.apiClient.Get(ctx, "/api/v1/projects")
	if err != nil {
		slog.Error("Failed to get projects", "error", err)
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	var projects []Project
	if err := json.Unmarshal(projectsResp, &projects); err != nil {
		slog.Error("Failed to parse projects", "error", err)
		return nil, fmt.Errorf("failed to parse projects: %w", err)
	}

	// Build formatted response
	response := buildProjectsOverviewResponse(projects)

	slog.Info("Projects overview resource retrieved", "project_count", len(projects))

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

// HandleProjectTasksResource handles project tasks resource requests
func (pr *ProjectResources) HandleProjectTasksResource(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.ReadResourceParams,
) (*mcp.ReadResourceResult, error) {
	slog.Info("Reading project tasks resource", "uri", params.URI)

	// Extract project ID from URI: taskman://project/{project_id}/tasks
	parts := strings.Split(params.URI, "/")
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid project tasks resource URI format: %s", params.URI)
	}
	projectID := parts[len(parts)-2] // Second to last part before "tasks"

	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	// Get project details for context
	projectResp, err := pr.apiClient.Get(ctx, fmt.Sprintf("/api/v1/projects/%s", url.PathEscape(projectID)))
	if err != nil {
		slog.Error("Failed to get project", "error", err, "project_id", projectID)
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	var project Project
	if err := json.Unmarshal(projectResp, &project); err != nil {
		slog.Error("Failed to parse project", "error", err)
		return nil, fmt.Errorf("failed to parse project: %w", err)
	}

	// Get project tasks
	tasksResp, err := pr.apiClient.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/tasks", url.PathEscape(projectID)))
	if err != nil {
		slog.Error("Failed to get project tasks", "error", err, "project_id", projectID)
		return nil, fmt.Errorf("failed to get project tasks: %w", err)
	}

	var tasks []Task
	if err := json.Unmarshal(tasksResp, &tasks); err != nil {
		slog.Error("Failed to parse project tasks", "error", err)
		return nil, fmt.Errorf("failed to parse project tasks: %w", err)
	}

	// Build formatted response
	response := buildProjectTasksResponse(project, tasks)

	slog.Info("Project tasks resource retrieved", "project_id", projectID, "task_count", len(tasks))

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

// buildProjectResourceResponse formats individual project data
func buildProjectResourceResponse(project Project, tasks []Task) string {
	var response strings.Builder

	response.WriteString(fmt.Sprintf("# Project: %s\n\n", project.ProjectName))
	response.WriteString(fmt.Sprintf("**ID:** %s\n", project.ProjectID))
	response.WriteString(fmt.Sprintf("**Created By:** %s\n", project.CreatedBy))
	response.WriteString(fmt.Sprintf("**Created:** %s\n", project.CreationDate))

	if project.ProjectDescription != nil && *project.ProjectDescription != "" {
		response.WriteString(fmt.Sprintf("\n**Description:**\n%s\n", *project.ProjectDescription))
	}

	response.WriteString(fmt.Sprintf("\n## Tasks Summary\n"))
	response.WriteString(fmt.Sprintf("**Total Tasks:** %d\n\n", len(tasks)))

	if len(tasks) > 0 {
		// Status breakdown
		statusCounts := make(map[string]int)
		priorityCounts := make(map[string]int)
		completedCount := 0

		for _, task := range tasks {
			statusCounts[task.Status]++

			if task.Priority != nil {
				priorityCounts[*task.Priority]++
			} else {
				priorityCounts["None"]++
			}

			if task.Status == "Complete" {
				completedCount++
			}
		}

		// Calculate completion percentage
		completionPercentage := float64(completedCount) / float64(len(tasks)) * 100
		response.WriteString(fmt.Sprintf("**Completion:** %.1f%%\n\n", completionPercentage))

		response.WriteString("### Status Breakdown\n")
		for status, count := range statusCounts {
			response.WriteString(fmt.Sprintf("- %s: %d\n", status, count))
		}

		response.WriteString("\n### Priority Breakdown\n")
		for priority, count := range priorityCounts {
			response.WriteString(fmt.Sprintf("- %s: %d\n", priority, count))
		}

		// Recent tasks
		response.WriteString("\n### Tasks\n")
		displayCount := len(tasks)
		if displayCount > 15 {
			displayCount = 15
		}

		for i := 0; i < displayCount; i++ {
			task := tasks[i]
			assignee := "Unassigned"
			if task.AssignedTo != nil {
				assignee = *task.AssignedTo
			}

			priority := "None"
			if task.Priority != nil {
				priority = *task.Priority
			}

			response.WriteString(fmt.Sprintf("- **%s** (%s, %s) - %s\n", task.TaskName, task.Status, priority, assignee))
		}

		if len(tasks) > 15 {
			response.WriteString(fmt.Sprintf("... and %d more tasks\n", len(tasks)-15))
		}
	} else {
		response.WriteString("No tasks in this project yet.\n")
	}

	return response.String()
}

// buildProjectsOverviewResponse formats projects overview data
func buildProjectsOverviewResponse(projects []Project) string {
	var response strings.Builder

	response.WriteString("# Projects Overview\n\n")
	response.WriteString(fmt.Sprintf("**Total Projects:** %d\n\n", len(projects)))

	if len(projects) == 0 {
		response.WriteString("No projects found.\n")
		return response.String()
	}

	// Group by creator
	creatorCounts := make(map[string]int)
	for _, project := range projects {
		creatorCounts[project.CreatedBy]++
	}

	response.WriteString("## Projects by Creator\n")
	for creator, count := range creatorCounts {
		response.WriteString(fmt.Sprintf("- %s: %d\n", creator, count))
	}

	// List all projects
	response.WriteString("\n## All Projects\n")
	for _, project := range projects {
		response.WriteString(fmt.Sprintf("- **%s** (%s) - Created by %s on %s\n",
			project.ProjectName, project.ProjectID, project.CreatedBy, project.CreationDate))

		if project.ProjectDescription != nil && *project.ProjectDescription != "" {
			// Truncate long descriptions
			desc := *project.ProjectDescription
			if len(desc) > 100 {
				desc = desc[:100] + "..."
			}
			response.WriteString(fmt.Sprintf("  *%s*\n", desc))
		}
		response.WriteString("\n")
	}

	return response.String()
}

// buildProjectTasksResponse formats project tasks data
func buildProjectTasksResponse(project Project, tasks []Task) string {
	var response strings.Builder

	response.WriteString(fmt.Sprintf("# Tasks in Project: %s\n\n", project.ProjectName))
	response.WriteString(fmt.Sprintf("**Project ID:** %s\n", project.ProjectID))
	response.WriteString(fmt.Sprintf("**Total Tasks:** %d\n\n", len(tasks)))

	if len(tasks) == 0 {
		response.WriteString("No tasks in this project.\n")
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
				assignee := "Unassigned"
				if task.AssignedTo != nil {
					assignee = *task.AssignedTo
				}

				priority := "None"
				if task.Priority != nil {
					priority = *task.Priority
				}

				dueDate := "No due date"
				if task.DueDate != nil {
					dueDate = *task.DueDate
				}

				response.WriteString(fmt.Sprintf("- **%s** (%s) - %s - Due: %s\n",
					task.TaskName, priority, assignee, dueDate))

				if task.TaskDescription != nil && *task.TaskDescription != "" {
					// Truncate long descriptions
					desc := *task.TaskDescription
					if len(desc) > 150 {
						desc = desc[:150] + "..."
					}
					response.WriteString(fmt.Sprintf("  *%s*\n", desc))
				}
				response.WriteString("\n")
			}
		}
	}

	return response.String()
}
