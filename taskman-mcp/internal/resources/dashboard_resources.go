package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/bchamber/taskman-mcp/internal/client"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// DashboardResources handles dashboard-related MCP resources
type DashboardResources struct {
	apiClient *client.APIClient
}

// NewDashboardResources creates a new dashboard resources handler
func NewDashboardResources(apiClient *client.APIClient) *DashboardResources {
	return &DashboardResources{
		apiClient: apiClient,
	}
}

// HandleSystemDashboardResource handles system dashboard resource requests
func (dr *DashboardResources) HandleSystemDashboardResource(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.ReadResourceParams,
) (*mcp.ReadResourceResult, error) {
	slog.Info("Reading system dashboard resource", "uri", params.URI)

	// Get all tasks
	tasksResp, err := dr.apiClient.Get(ctx, "/api/v1/tasks")
	if err != nil {
		slog.Error("Failed to get tasks", "error", err)
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	var tasks []Task
	if err := json.Unmarshal(tasksResp, &tasks); err != nil {
		slog.Error("Failed to parse tasks", "error", err)
		return nil, fmt.Errorf("failed to parse tasks: %w", err)
	}

	// Get all projects
	projectsResp, err := dr.apiClient.Get(ctx, "/api/v1/projects")
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
	response := buildSystemDashboardResponse(tasks, projects)

	slog.Info("System dashboard resource retrieved", "task_count", len(tasks), "project_count", len(projects))

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

// HandleUserDashboardResource handles user dashboard resource requests
func (dr *DashboardResources) HandleUserDashboardResource(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.ReadResourceParams,
) (*mcp.ReadResourceResult, error) {
	slog.Info("Reading user dashboard resource", "uri", params.URI)

	// Extract user ID from URI: taskman://dashboard/user/{user_id}
	parts := strings.Split(params.URI, "/")
	if len(parts) != 5 || parts[0] != "taskman:" || parts[1] != "" || parts[2] != "dashboard" || parts[3] != "user" {
		return nil, fmt.Errorf("invalid user dashboard resource URI format: %s", params.URI)
	}
	userID := parts[4]

	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Get tasks assigned to user
	tasksResp, err := dr.apiClient.Get(ctx, fmt.Sprintf("/api/v1/tasks?assigned_to=%s", url.QueryEscape(userID)))
	if err != nil {
		slog.Error("Failed to get user tasks", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to get user tasks: %w", err)
	}

	var tasks []Task
	if err := json.Unmarshal(tasksResp, &tasks); err != nil {
		slog.Error("Failed to parse user tasks", "error", err)
		return nil, fmt.Errorf("failed to parse user tasks: %w", err)
	}

	// Get tasks created by user
	createdTasksResp, err := dr.apiClient.Get(ctx, fmt.Sprintf("/api/v1/tasks?created_by=%s", url.QueryEscape(userID)))
	if err != nil {
		slog.Warn("Failed to get user created tasks", "error", err, "user_id", userID)
	}

	var createdTasks []Task
	if createdTasksResp != nil {
		if err := json.Unmarshal(createdTasksResp, &createdTasks); err != nil {
			slog.Warn("Failed to parse user created tasks", "error", err)
		}
	}

	// Build formatted response
	response := buildUserDashboardResponse(userID, tasks, createdTasks)

	slog.Info("User dashboard resource retrieved", "user_id", userID, "assigned_tasks", len(tasks), "created_tasks", len(createdTasks))

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

// HandleProjectDashboardResource handles project dashboard resource requests
func (dr *DashboardResources) HandleProjectDashboardResource(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.ReadResourceParams,
) (*mcp.ReadResourceResult, error) {
	slog.Info("Reading project dashboard resource", "uri", params.URI)

	// Extract project ID from URI: taskman://dashboard/project/{project_id}
	parts := strings.Split(params.URI, "/")
	if len(parts) != 5 || parts[0] != "taskman:" || parts[1] != "" || parts[2] != "dashboard" || parts[3] != "project" {
		return nil, fmt.Errorf("invalid project dashboard resource URI format: %s", params.URI)
	}
	projectID := parts[4]

	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	// Get project details
	projectResp, err := dr.apiClient.Get(ctx, fmt.Sprintf("/api/v1/projects/%s", url.PathEscape(projectID)))
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
	tasksResp, err := dr.apiClient.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/tasks", url.PathEscape(projectID)))
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
	response := buildProjectDashboardResponse(project, tasks)

	slog.Info("Project dashboard resource retrieved", "project_id", projectID, "task_count", len(tasks))

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

// buildSystemDashboardResponse formats system dashboard data
func buildSystemDashboardResponse(tasks []Task, projects []Project) string {
	var response strings.Builder

	response.WriteString("# System Dashboard\n\n")
	response.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	// Overall statistics
	response.WriteString("## Overview\n")
	response.WriteString(fmt.Sprintf("- **Total Projects:** %d\n", len(projects)))
	response.WriteString(fmt.Sprintf("- **Total Tasks:** %d\n", len(tasks)))

	if len(tasks) > 0 {
		// Task statistics
		statusCounts := make(map[string]int)
		priorityCounts := make(map[string]int)
		assigneeCounts := make(map[string]int)
		overdueTasks := 0
		completedTasks := 0

		now := time.Now()

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

			if task.Status == "Complete" {
				completedTasks++
			}

			// Check if overdue
			if task.DueDate != nil && task.Status != "Complete" {
				if dueDate, err := time.Parse(time.RFC3339, *task.DueDate); err == nil {
					if dueDate.Before(now) {
						overdueTasks++
					}
				}
			}
		}

		completionRate := float64(completedTasks) / float64(len(tasks)) * 100
		response.WriteString(fmt.Sprintf("- **Completion Rate:** %.1f%%\n", completionRate))
		response.WriteString(fmt.Sprintf("- **Overdue Tasks:** %d\n\n", overdueTasks))

		// Task status breakdown
		response.WriteString("## Task Status Distribution\n")
		for status, count := range statusCounts {
			percentage := float64(count) / float64(len(tasks)) * 100
			response.WriteString(fmt.Sprintf("- **%s:** %d (%.1f%%)\n", status, count, percentage))
		}

		// Priority breakdown
		response.WriteString("\n## Priority Distribution\n")
		for priority, count := range priorityCounts {
			percentage := float64(count) / float64(len(tasks)) * 100
			response.WriteString(fmt.Sprintf("- **%s:** %d (%.1f%%)\n", priority, count, percentage))
		}

		// Top assignees
		response.WriteString("\n## Top Assignees\n")
		assigneeList := make([]struct {
			name  string
			count int
		}, 0, len(assigneeCounts))
		for name, count := range assigneeCounts {
			assigneeList = append(assigneeList, struct {
				name  string
				count int
			}{name, count})
		}

		// Sort by count (simple bubble sort for minimal code)
		for i := 0; i < len(assigneeList); i++ {
			for j := i + 1; j < len(assigneeList); j++ {
				if assigneeList[j].count > assigneeList[i].count {
					assigneeList[i], assigneeList[j] = assigneeList[j], assigneeList[i]
				}
			}
		}

		displayCount := len(assigneeList)
		if displayCount > 5 {
			displayCount = 5
		}

		for i := 0; i < displayCount; i++ {
			assignee := assigneeList[i]
			response.WriteString(fmt.Sprintf("- **%s:** %d tasks\n", assignee.name, assignee.count))
		}

		// Recent activity
		if len(tasks) > 0 {
			response.WriteString("\n## Recent Tasks\n")
			recentCount := len(tasks)
			if recentCount > 5 {
				recentCount = 5
			}

			for i := 0; i < recentCount; i++ {
				task := tasks[i]
				assignee := "Unassigned"
				if task.AssignedTo != nil {
					assignee = *task.AssignedTo
				}
				response.WriteString(fmt.Sprintf("- **%s** (%s) - %s - %s\n",
					task.TaskName, task.Status, assignee, task.CreationDate))
			}
		}
	}

	// Project information
	if len(projects) > 0 {
		response.WriteString("\n## Recent Projects\n")
		projectCount := len(projects)
		if projectCount > 5 {
			projectCount = 5
		}

		for i := 0; i < projectCount; i++ {
			project := projects[i]
			response.WriteString(fmt.Sprintf("- **%s** - Created by %s on %s\n",
				project.ProjectName, project.CreatedBy, project.CreationDate))
		}
	}

	return response.String()
}

// buildUserDashboardResponse formats user dashboard data
func buildUserDashboardResponse(userID string, assignedTasks []Task, createdTasks []Task) string {
	var response strings.Builder

	response.WriteString(fmt.Sprintf("# Dashboard for %s\n\n", userID))
	response.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	// User statistics
	response.WriteString("## My Statistics\n")
	response.WriteString(fmt.Sprintf("- **Assigned Tasks:** %d\n", len(assignedTasks)))
	response.WriteString(fmt.Sprintf("- **Created Tasks:** %d\n", len(createdTasks)))

	if len(assignedTasks) > 0 {
		// Analyze assigned tasks
		statusCounts := make(map[string]int)
		priorityCounts := make(map[string]int)
		overdueTasks := 0
		completedTasks := 0

		now := time.Now()

		for _, task := range assignedTasks {
			statusCounts[task.Status]++

			if task.Priority != nil {
				priorityCounts[*task.Priority]++
			} else {
				priorityCounts["None"]++
			}

			if task.Status == "Complete" {
				completedTasks++
			}

			// Check if overdue
			if task.DueDate != nil && task.Status != "Complete" {
				if dueDate, err := time.Parse(time.RFC3339, *task.DueDate); err == nil {
					if dueDate.Before(now) {
						overdueTasks++
					}
				}
			}
		}

		completionRate := float64(completedTasks) / float64(len(assignedTasks)) * 100
		response.WriteString(fmt.Sprintf("- **My Completion Rate:** %.1f%%\n", completionRate))
		response.WriteString(fmt.Sprintf("- **My Overdue Tasks:** %d\n\n", overdueTasks))

		// My task status breakdown
		response.WriteString("## My Task Status\n")
		for status, count := range statusCounts {
			response.WriteString(fmt.Sprintf("- **%s:** %d\n", status, count))
		}

		// My priority breakdown
		response.WriteString("\n## My Task Priorities\n")
		for priority, count := range priorityCounts {
			response.WriteString(fmt.Sprintf("- **%s:** %d\n", priority, count))
		}

		// Current workload
		response.WriteString("\n## Current Workload\n")
		activeTasks := make([]Task, 0)
		for _, task := range assignedTasks {
			if task.Status == "In Progress" || task.Status == "Review" || task.Status == "Blocked" {
				activeTasks = append(activeTasks, task)
			}
		}

		if len(activeTasks) > 0 {
			response.WriteString(fmt.Sprintf("**Active Tasks:** %d\n\n", len(activeTasks)))

			for _, task := range activeTasks {
				priority := "None"
				if task.Priority != nil {
					priority = *task.Priority
				}

				dueDate := "No due date"
				if task.DueDate != nil {
					dueDate = *task.DueDate
				}

				response.WriteString(fmt.Sprintf("- **%s** (%s, %s) - Due: %s\n",
					task.TaskName, task.Status, priority, dueDate))
			}
		} else {
			response.WriteString("No active tasks assigned.\n")
		}

		// Upcoming deadlines
		upcomingTasks := make([]Task, 0)
		for _, task := range assignedTasks {
			if task.DueDate != nil && task.Status != "Complete" {
				if dueDate, err := time.Parse(time.RFC3339, *task.DueDate); err == nil {
					// Tasks due in the next 7 days
					if dueDate.After(now) && dueDate.Before(now.AddDate(0, 0, 7)) {
						upcomingTasks = append(upcomingTasks, task)
					}
				}
			}
		}

		if len(upcomingTasks) > 0 {
			response.WriteString("\n## Upcoming Deadlines (Next 7 Days)\n")
			for _, task := range upcomingTasks {
				priority := "None"
				if task.Priority != nil {
					priority = *task.Priority
				}
				response.WriteString(fmt.Sprintf("- **%s** (%s, %s) - Due: %s\n",
					task.TaskName, task.Status, priority, *task.DueDate))
			}
		}
	} else {
		response.WriteString("\nNo tasks currently assigned.\n")
	}

	return response.String()
}

// buildProjectDashboardResponse formats project dashboard data
func buildProjectDashboardResponse(project Project, tasks []Task) string {
	var response strings.Builder

	response.WriteString(fmt.Sprintf("# Project Dashboard: %s\n\n", project.ProjectName))
	response.WriteString(fmt.Sprintf("**Project ID:** %s\n", project.ProjectID))
	response.WriteString(fmt.Sprintf("**Created by:** %s on %s\n", project.CreatedBy, project.CreationDate))
	response.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	if project.ProjectDescription != nil && *project.ProjectDescription != "" {
		response.WriteString(fmt.Sprintf("**Description:** %s\n\n", *project.ProjectDescription))
	}

	// Project statistics
	response.WriteString("## Project Statistics\n")
	response.WriteString(fmt.Sprintf("- **Total Tasks:** %d\n", len(tasks)))

	if len(tasks) > 0 {
		// Analyze tasks
		statusCounts := make(map[string]int)
		priorityCounts := make(map[string]int)
		assigneeCounts := make(map[string]int)
		overdueTasks := 0
		completedTasks := 0

		now := time.Now()

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

			if task.Status == "Complete" {
				completedTasks++
			}

			// Check if overdue
			if task.DueDate != nil && task.Status != "Complete" {
				if dueDate, err := time.Parse(time.RFC3339, *task.DueDate); err == nil {
					if dueDate.Before(now) {
						overdueTasks++
					}
				}
			}
		}

		completionRate := float64(completedTasks) / float64(len(tasks)) * 100
		response.WriteString(fmt.Sprintf("- **Completion Rate:** %.1f%%\n", completionRate))
		response.WriteString(fmt.Sprintf("- **Overdue Tasks:** %d\n\n", overdueTasks))

		// Status breakdown
		response.WriteString("## Task Status Distribution\n")
		for status, count := range statusCounts {
			percentage := float64(count) / float64(len(tasks)) * 100
			response.WriteString(fmt.Sprintf("- **%s:** %d (%.1f%%)\n", status, count, percentage))
		}

		// Team workload
		response.WriteString("\n## Team Workload\n")
		for assignee, count := range assigneeCounts {
			response.WriteString(fmt.Sprintf("- **%s:** %d tasks\n", assignee, count))
		}

		// Priority distribution
		response.WriteString("\n## Priority Distribution\n")
		for priority, count := range priorityCounts {
			response.WriteString(fmt.Sprintf("- **%s:** %d tasks\n", priority, count))
		}

		// Critical tasks (high priority or overdue)
		criticalTasks := make([]Task, 0)
		for _, task := range tasks {
			if task.Status != "Complete" {
				isHighPriority := task.Priority != nil && *task.Priority == "High"
				isOverdue := false

				if task.DueDate != nil {
					if dueDate, err := time.Parse(time.RFC3339, *task.DueDate); err == nil {
						isOverdue = dueDate.Before(now)
					}
				}

				if isHighPriority || isOverdue {
					criticalTasks = append(criticalTasks, task)
				}
			}
		}

		if len(criticalTasks) > 0 {
			response.WriteString("\n## Critical Tasks (High Priority or Overdue)\n")
			for _, task := range criticalTasks {
				assignee := "Unassigned"
				if task.AssignedTo != nil {
					assignee = *task.AssignedTo
				}

				priority := "None"
				if task.Priority != nil {
					priority = *task.Priority
				}

				dueInfo := "No due date"
				if task.DueDate != nil {
					dueInfo = *task.DueDate
				}

				response.WriteString(fmt.Sprintf("- **%s** (%s, %s) - %s - Due: %s\n",
					task.TaskName, task.Status, priority, assignee, dueInfo))
			}
		}

		// Recent activity
		response.WriteString("\n## Recent Activity\n")
		recentCount := len(tasks)
		if recentCount > 8 {
			recentCount = 8
		}

		for i := 0; i < recentCount; i++ {
			task := tasks[i]
			assignee := "Unassigned"
			if task.AssignedTo != nil {
				assignee = *task.AssignedTo
			}
			response.WriteString(fmt.Sprintf("- **%s** (%s) - %s - %s\n",
				task.TaskName, task.Status, assignee, task.CreationDate))
		}

		if len(tasks) > 8 {
			response.WriteString(fmt.Sprintf("... and %d more tasks\n", len(tasks)-8))
		}
	} else {
		response.WriteString("\nNo tasks in this project yet.\n")
	}

	return response.String()
}
