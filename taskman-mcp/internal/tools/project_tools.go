package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/bchamber/taskman-mcp/internal/client"
)

// ProjectTools handles project management MCP tools
type ProjectTools struct {
	apiClient *client.APIClient
}

// NewProjectTools creates a new project tools handler
func NewProjectTools(apiClient *client.APIClient) *ProjectTools {
	return &ProjectTools{
		apiClient: apiClient,
	}
}

// GetProjectStatusParams defines input for get_project_status tool
type GetProjectStatusParams struct {
	ProjectID string `json:"project_id"`
}

// CreateProjectWithInitialTasksParams defines input for create_project_with_initial_tasks tool
type CreateProjectWithInitialTasksParams struct {
	ProjectName        string                 `json:"project_name"`
	ProjectDescription string                 `json:"project_description,omitempty"`
	CreatedBy          string                 `json:"created_by"`
	InitialTasks       []InitialTaskSpec      `json:"initial_tasks"`
}

// InitialTaskSpec defines a task to be created with the project
type InitialTaskSpec struct {
	TaskName        string `json:"task_name"`
	TaskDescription string `json:"task_description,omitempty"`
	Status          string `json:"status,omitempty"`
	Priority        string `json:"priority,omitempty"`
	AssignedTo      string `json:"assigned_to,omitempty"`
	DueDate         string `json:"due_date,omitempty"`
}

// HandleGetProjectStatus implements the get_project_status tool
func (p *ProjectTools) HandleGetProjectStatus(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.CallToolParamsFor[GetProjectStatusParams],
) (*mcp.CallToolResultFor[map[string]any], error) {
	slog.Info("Executing get_project_status tool", "params", params.Arguments)
	
	// Validate required fields
	if params.Arguments.ProjectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}
	
	// Get project details
	projectResp, err := p.apiClient.Get(ctx, fmt.Sprintf("/api/v1/projects/%s", url.PathEscape(params.Arguments.ProjectID)))
	if err != nil {
		slog.Error("Failed to get project", "error", err, "project_id", params.Arguments.ProjectID)
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	
	var project Project
	if err := json.Unmarshal(projectResp, &project); err != nil {
		slog.Error("Failed to parse project", "error", err)
		return nil, fmt.Errorf("failed to parse project: %w", err)
	}
	
	// Get project tasks
	tasksResp, err := p.apiClient.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/tasks", url.PathEscape(params.Arguments.ProjectID)))
	if err != nil {
		slog.Error("Failed to get project tasks", "error", err, "project_id", params.Arguments.ProjectID)
		return nil, fmt.Errorf("failed to get project tasks: %w", err)
	}
	
	var tasks []Task
	if err := json.Unmarshal(tasksResp, &tasks); err != nil {
		slog.Error("Failed to parse project tasks", "error", err)
		return nil, fmt.Errorf("failed to parse project tasks: %w", err)
	}
	
	// Analyze project metrics
	statusCounts := make(map[string]int)
	priorityCounts := make(map[string]int)
	overdueTasks := []Task{}
	completedTasks := []Task{}
	activeTasks := []Task{}
	
	totalTasks := len(tasks)
	
	for _, task := range tasks {
		// Count by status
		statusCounts[task.Status]++
		
		// Count by priority
		if task.Priority != nil {
			priorityCounts[*task.Priority]++
		} else {
			priorityCounts["None"]++
		}
		
		// Categorize tasks
		switch task.Status {
		case "Complete":
			completedTasks = append(completedTasks, task)
		case "In Progress", "Review":
			activeTasks = append(activeTasks, task)
		}
		
		// Check if overdue
		if isTaskOverdue(task) {
			overdueTasks = append(overdueTasks, task)
		}
	}
	
	// Calculate completion percentage
	completedCount := statusCounts["Complete"]
	var completionPercentage float64
	if totalTasks > 0 {
		completionPercentage = float64(completedCount) / float64(totalTasks) * 100
	}
	
	// Generate project insights
	var insights []string
	
	if len(overdueTasks) > 0 {
		insights = append(insights, fmt.Sprintf("⚠️ %d tasks are overdue and need attention", len(overdueTasks)))
	}
	
	if completionPercentage >= 90 {
		insights = append(insights, "🎉 Project is nearly complete!")
	} else if completionPercentage >= 75 {
		insights = append(insights, "📈 Project is in final stretch")
	} else if completionPercentage >= 50 {
		insights = append(insights, "🔄 Project is halfway complete")
	} else if completionPercentage < 25 && totalTasks > 0 {
		insights = append(insights, "🚀 Project is in early stages")
	}
	
	if len(activeTasks) > totalTasks/2 && totalTasks > 0 {
		insights = append(insights, "🔥 High activity - many tasks in progress")
	}
	
	notStartedCount := statusCounts["Not Started"]
	if notStartedCount > len(activeTasks) && totalTasks > 3 {
		insights = append(insights, "📋 Consider starting more tasks to increase momentum")
	}
	
	// Generate next actions
	var nextActions []string
	
	if len(overdueTasks) > 0 {
		nextActions = append(nextActions, "🚨 Address overdue tasks immediately")
	}
	
	if notStartedCount > 0 {
		nextActions = append(nextActions, fmt.Sprintf("▶️ Start work on %d pending tasks", notStartedCount))
	}
	
	reviewCount := statusCounts["Review"]
	if reviewCount > 0 {
		nextActions = append(nextActions, fmt.Sprintf("👀 Review %d tasks waiting for approval", reviewCount))
	}
	
	blockedCount := statusCounts["Blocked"]
	if blockedCount > 0 {
		nextActions = append(nextActions, fmt.Sprintf("🔓 Resolve %d blocked tasks", blockedCount))
	}
	
	if completionPercentage >= 90 {
		nextActions = append(nextActions, "🏁 Plan project closure activities")
	}
	
	// Build comprehensive response
	result := map[string]any{
		"project":              project,
		"tasks":                tasks,
		"total_tasks":          totalTasks,
		"completion_percentage": completionPercentage,
		"status_breakdown":     statusCounts,
		"priority_breakdown":   priorityCounts,
		"overdue_count":        len(overdueTasks),
		"overdue_tasks":        overdueTasks,
		"active_tasks":         activeTasks,
		"completed_tasks":      completedTasks,
		"insights":             insights,
		"next_actions":         nextActions,
	}
	
	// Build detailed response text
	responseText := fmt.Sprintf(`Project Status Report\n====================\n\nProject: %s\nID: %s\n`, 
		project.ProjectName, project.ProjectID)
	
	if project.ProjectDescription != nil && *project.ProjectDescription != "" {
		responseText += fmt.Sprintf("Description: %s\n", *project.ProjectDescription)
	}
	
	responseText += fmt.Sprintf("\nCreated by: %s\nCreated: %s\n", 
		project.CreatedBy, project.CreationDate)
	
	responseText += fmt.Sprintf("\n📊 Project Metrics:\n")
	responseText += fmt.Sprintf("Total Tasks: %d\n", totalTasks)
	responseText += fmt.Sprintf("Completion: %.1f%%\n", completionPercentage)
	
	responseText += fmt.Sprintf("\n📈 Status Breakdown:\n")
	for status, count := range statusCounts {
		responseText += fmt.Sprintf("- %s: %d\n", status, count)
	}
	
	if len(priorityCounts) > 0 {
		responseText += fmt.Sprintf("\n🎯 Priority Breakdown:\n")
		for priority, count := range priorityCounts {
			responseText += fmt.Sprintf("- %s: %d\n", priority, count)
		}
	}
	
	if len(overdueTasks) > 0 {
		responseText += fmt.Sprintf("\n⚠️ Overdue Tasks (%d):\n", len(overdueTasks))
		for _, task := range overdueTasks {
			responseText += fmt.Sprintf("- %s (Due: %s)\n", task.TaskName, *task.DueDate)
		}
	}
	
	if len(activeTasks) > 0 {
		responseText += fmt.Sprintf("\n🔄 Active Tasks (%d):\n", len(activeTasks))
		for i, task := range activeTasks {
			if i < 5 { // Show only first 5
				assignee := "Unassigned"
				if task.AssignedTo != nil {
					assignee = *task.AssignedTo
				}
				responseText += fmt.Sprintf("- %s (%s) - %s\n", task.TaskName, task.Status, assignee)
			}
		}
		if len(activeTasks) > 5 {
			responseText += fmt.Sprintf("... and %d more active tasks\n", len(activeTasks)-5)
		}
	}
	
	if len(insights) > 0 {
		responseText += "\n💡 Insights:\n"
		for _, insight := range insights {
			responseText += fmt.Sprintf("- %s\n", insight)
		}
	}
	
	if len(nextActions) > 0 {
		responseText += "\n📋 Recommended Actions:\n"
		for _, action := range nextActions {
			responseText += fmt.Sprintf("- %s\n", action)
		}
	}
	
	slog.Info("Project status generated", "project_id", project.ProjectID, "total_tasks", totalTasks, "completion", completionPercentage)
	
	return &mcp.CallToolResultFor[map[string]any]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: responseText,
			},
		},
		Meta: result,
	}, nil
}

// HandleCreateProjectWithInitialTasks implements the create_project_with_initial_tasks tool
func (p *ProjectTools) HandleCreateProjectWithInitialTasks(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.CallToolParamsFor[CreateProjectWithInitialTasksParams],
) (*mcp.CallToolResultFor[map[string]any], error) {
	slog.Info("Executing create_project_with_initial_tasks tool", "params", params.Arguments)
	
	// Validate required fields
	if params.Arguments.ProjectName == "" {
		return nil, fmt.Errorf("project_name is required")
	}
	if params.Arguments.CreatedBy == "" {
		return nil, fmt.Errorf("created_by is required")
	}
	if len(params.Arguments.InitialTasks) == 0 {
		return nil, fmt.Errorf("initial_tasks are required (at least one task)")
	}
	
	// Build project creation request
	projectRequest := map[string]interface{}{
		"project_name": params.Arguments.ProjectName,
		"created_by":   params.Arguments.CreatedBy,
	}
	
	if params.Arguments.ProjectDescription != "" {
		projectRequest["project_description"] = params.Arguments.ProjectDescription
	}
	
	// Create the project
	projectResp, err := p.apiClient.Post(ctx, "/api/v1/projects", projectRequest)
	if err != nil {
		slog.Error("Failed to create project", "error", err)
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	
	var createdProject Project
	if err := json.Unmarshal(projectResp, &createdProject); err != nil {
		slog.Error("Failed to parse created project", "error", err)
		return nil, fmt.Errorf("failed to parse created project: %w", err)
	}
	
	// Create initial tasks
	var createdTasks []Task
	var failedTasks []InitialTaskSpec
	
	for _, taskSpec := range params.Arguments.InitialTasks {
		taskRequest := map[string]interface{}{
			"task_name":  taskSpec.TaskName,
			"project_id": createdProject.ProjectID,
			"created_by": params.Arguments.CreatedBy,
		}
		
		if taskSpec.TaskDescription != "" {
			taskRequest["task_description"] = taskSpec.TaskDescription
		}
		if taskSpec.Status != "" {
			taskRequest["status"] = taskSpec.Status
		} else {
			taskRequest["status"] = "Not Started"
		}
		if taskSpec.Priority != "" {
			taskRequest["priority"] = taskSpec.Priority
		}
		if taskSpec.AssignedTo != "" {
			taskRequest["assigned_to"] = taskSpec.AssignedTo
		}
		if taskSpec.DueDate != "" {
			// Validate and format due date
			if dueDate, err := parseDueDate(taskSpec.DueDate); err == nil && dueDate != nil {
				taskRequest["due_date"] = dueDate.Format(time.RFC3339)
			} else {
				slog.Warn("Failed to parse due date for task", "task_name", taskSpec.TaskName, "due_date", taskSpec.DueDate, "error", err)
			}
		}
		
		taskResp, err := p.apiClient.Post(ctx, "/api/v1/tasks", taskRequest)
		if err != nil {
			slog.Error("Failed to create task", "error", err, "task_name", taskSpec.TaskName)
			failedTasks = append(failedTasks, taskSpec)
			continue
		}
		
		var createdTask Task
		if err := json.Unmarshal(taskResp, &createdTask); err != nil {
			slog.Error("Failed to parse created task", "error", err, "task_name", taskSpec.TaskName)
			failedTasks = append(failedTasks, taskSpec)
			continue
		}
		
		createdTasks = append(createdTasks, createdTask)
	}
	
	// Analyze task creation results
	var insights []string
	
	if len(failedTasks) == 0 {
		insights = append(insights, "✅ All initial tasks created successfully")
	} else {
		insights = append(insights, fmt.Sprintf("⚠️ %d tasks failed to create", len(failedTasks)))
	}
	
	if len(createdTasks) > 5 {
		insights = append(insights, "📋 Large project with many initial tasks")
	}
	
	// Count task priorities and assignments
	priorityCounts := make(map[string]int)
	assignedCount := 0
	
	for _, task := range createdTasks {
		if task.Priority != nil {
			priorityCounts[*task.Priority]++
		}
		if task.AssignedTo != nil && *task.AssignedTo != "" {
			assignedCount++
		}
	}
	
	if assignedCount == 0 {
		insights = append(insights, "👤 No tasks assigned yet - consider assigning team members")
	} else if assignedCount == len(createdTasks) {
		insights = append(insights, "👥 All tasks have been assigned")
	}
	
	highPriorityCount := priorityCounts["High"]
	if highPriorityCount > len(createdTasks)/2 {
		insights = append(insights, "🔥 Many high-priority tasks - ensure adequate resources")
	}
	
	// Generate next steps
	var nextSteps []string
	
	nextSteps = append(nextSteps, "🎉 Project created successfully")
	
	if assignedCount < len(createdTasks) {
		nextSteps = append(nextSteps, fmt.Sprintf("👤 Assign %d remaining tasks to team members", len(createdTasks)-assignedCount))
	}
	
	if len(failedTasks) > 0 {
		nextSteps = append(nextSteps, "🔄 Retry creating failed tasks manually")
	}
	
	nextSteps = append(nextSteps, "📅 Review and adjust task due dates as needed")
	nextSteps = append(nextSteps, "🚀 Begin work on the first tasks")
	nextSteps = append(nextSteps, "📊 Set up regular project status reviews")
	
	// Build comprehensive response
	result := map[string]any{
		"project":        createdProject,
		"created_tasks":  createdTasks,
		"failed_tasks":   failedTasks,
		"total_planned":  len(params.Arguments.InitialTasks),
		"total_created":  len(createdTasks),
		"total_failed":   len(failedTasks),
		"success_rate":   float64(len(createdTasks)) / float64(len(params.Arguments.InitialTasks)) * 100,
		"insights":       insights,
		"next_steps":     nextSteps,
	}
	
	// Build response text
	responseText := fmt.Sprintf(`Project Created with Initial Tasks\n=================================\n\nProject: %s\nID: %s\n`, 
		createdProject.ProjectName, createdProject.ProjectID)
	
	if createdProject.ProjectDescription != nil && *createdProject.ProjectDescription != "" {
		responseText += fmt.Sprintf("Description: %s\n", *createdProject.ProjectDescription)
	}
	
	responseText += fmt.Sprintf("Created by: %s\n", createdProject.CreatedBy)
	
	responseText += fmt.Sprintf("\n📊 Task Creation Summary:\n")
	responseText += fmt.Sprintf("Planned: %d tasks\n", len(params.Arguments.InitialTasks))
	responseText += fmt.Sprintf("Created: %d tasks\n", len(createdTasks))
	
	if len(failedTasks) > 0 {
		responseText += fmt.Sprintf("Failed: %d tasks\n", len(failedTasks))
		successRate := float64(len(createdTasks)) / float64(len(params.Arguments.InitialTasks)) * 100
		responseText += fmt.Sprintf("Success Rate: %.1f%%\n", successRate)
	}
	
	if len(createdTasks) > 0 {
		responseText += fmt.Sprintf("\n✅ Created Tasks:\n")
		for _, task := range createdTasks {
			status := task.Status
			priority := "None"
			if task.Priority != nil {
				priority = *task.Priority
			}
			assignee := "Unassigned"
			if task.AssignedTo != nil && *task.AssignedTo != "" {
				assignee = *task.AssignedTo
			}
			responseText += fmt.Sprintf("- %s (%s, %s) - %s\n", task.TaskName, status, priority, assignee)
		}
	}
	
	if len(failedTasks) > 0 {
		responseText += fmt.Sprintf("\n❌ Failed Tasks:\n")
		for _, taskSpec := range failedTasks {
			responseText += fmt.Sprintf("- %s\n", taskSpec.TaskName)
		}
	}
	
	if len(insights) > 0 {
		responseText += "\n💡 Insights:\n"
		for _, insight := range insights {
			responseText += fmt.Sprintf("- %s\n", insight)
		}
	}
	
	if len(nextSteps) > 0 {
		responseText += "\n📋 Next Steps:\n"
		for _, step := range nextSteps {
			responseText += fmt.Sprintf("- %s\n", step)
		}
	}
	
	slog.Info("Project created with initial tasks", "project_id", createdProject.ProjectID, "tasks_created", len(createdTasks), "tasks_failed", len(failedTasks))
	
	return &mcp.CallToolResultFor[map[string]any]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: responseText,
			},
		},
		Meta: result,
	}, nil
}