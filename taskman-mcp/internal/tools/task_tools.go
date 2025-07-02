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

// TaskTools handles task management MCP tools
type TaskTools struct {
	apiClient *client.APIClient
}

// NewTaskTools creates a new task tools handler
func NewTaskTools(apiClient *client.APIClient) *TaskTools {
	return &TaskTools{
		apiClient: apiClient,
	}
}

// GetTaskOverviewParams defines input for get_task_overview tool
type GetTaskOverviewParams struct {
	Status     string `json:"status,omitempty"`
	AssignedTo string `json:"assigned_to,omitempty"`
	ProjectID  string `json:"project_id,omitempty"`
}

// CreateTaskWithContextParams defines input for create_task_with_context tool
type CreateTaskWithContextParams struct {
	TaskName        string  `json:"task_name"`
	TaskDescription string  `json:"task_description,omitempty"`
	Status          string  `json:"status,omitempty"`
	Priority        string  `json:"priority,omitempty"`
	AssignedTo      string  `json:"assigned_to,omitempty"`
	ProjectID       string  `json:"project_id,omitempty"`
	DueDate         string  `json:"due_date,omitempty"`
	InitialNote     string  `json:"initial_note"`
	CreatedBy       string  `json:"created_by"`
}

// GetTaskDetailsParams defines input for get_task_details tool
type GetTaskDetailsParams struct {
	TaskID string `json:"task_id"`
}

// UpdateTaskProgressParams defines input for update_task_progress tool
type UpdateTaskProgressParams struct {
	TaskID      string  `json:"task_id"`
	Status      string  `json:"status,omitempty"`
	Priority    string  `json:"priority,omitempty"`
	AssignedTo  string  `json:"assigned_to,omitempty"`
	ProgressNote string  `json:"progress_note"`
	UpdatedBy   string  `json:"updated_by"`
}

// Task represents a task from the API
type Task struct {
	TaskID          string    `json:"task_id"`
	TaskName        string    `json:"task_name"`
	TaskDescription *string   `json:"task_description"`
	Status          string    `json:"status"`
	Priority        *string   `json:"priority"`
	AssignedTo      *string   `json:"assigned_to"`
	ProjectID       *string   `json:"project_id"`
	DueDate         *string   `json:"due_date"`
	StartDate       *string   `json:"start_date"`
	CompletionDate  *string   `json:"completion_date"`
	Tags            []string  `json:"tags"`
	Archived        bool      `json:"archived"`
	CreatedBy       string    `json:"created_by"`
	CreationDate    string    `json:"creation_date"`
	LastUpdatedBy   *string   `json:"last_updated_by"`
	LastUpdateDate  *string   `json:"last_update_date"`
}

// Project represents a project from the API
type Project struct {
	ProjectID          string  `json:"project_id"`
	ProjectName        string  `json:"project_name"`
	ProjectDescription *string `json:"project_description"`
	CreatedBy          string  `json:"created_by"`
	CreationDate       string  `json:"creation_date"`
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

// Helper function to parse due dates
func parseDueDate(dueDateStr string) (*time.Time, error) {
	if dueDateStr == "" {
		return nil, nil
	}
	
	// Try common date formats
	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		time.RFC3339,
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, dueDateStr); err == nil {
			return &t, nil
		}
	}
	
	return nil, fmt.Errorf("unable to parse date: %s", dueDateStr)
}

// Helper function to check if a task is overdue
func isTaskOverdue(task Task) bool {
	if task.Status == "Complete" || task.DueDate == nil {
		return false
	}
	
	dueTime, err := time.Parse(time.RFC3339, *task.DueDate)
	if err != nil {
		return false
	}
	
	return dueTime.Before(time.Now())
}

// HandleGetTaskOverview implements the get_task_overview tool
func (t *TaskTools) HandleGetTaskOverview(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.CallToolParamsFor[GetTaskOverviewParams],
) (*mcp.CallToolResultFor[map[string]any], error) {
	slog.Info("Executing get_task_overview tool", "params", params.Arguments)
	
	// Build query parameters
	queryParams := ""
	if params.Arguments.Status != "" {
		if queryParams == "" {
			queryParams += "?"
		} else {
			queryParams += "&"
		}
		queryParams += fmt.Sprintf("status=%s", url.QueryEscape(params.Arguments.Status))
	}
	if params.Arguments.AssignedTo != "" {
		if queryParams == "" {
			queryParams += "?"
		} else {
			queryParams += "&"
		}
		queryParams += fmt.Sprintf("assigned_to=%s", url.QueryEscape(params.Arguments.AssignedTo))
	}
	if params.Arguments.ProjectID != "" {
		if queryParams == "" {
			queryParams += "?"
		} else {
			queryParams += "&"
		}
		queryParams += fmt.Sprintf("project_id=%s", url.QueryEscape(params.Arguments.ProjectID))
	}
	
	// Get tasks
	tasksResp, err := t.apiClient.Get(ctx, "/api/v1/tasks"+queryParams)
	if err != nil {
		slog.Error("Failed to get tasks", "error", err)
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}
	
	var tasks []Task
	if err := json.Unmarshal(tasksResp, &tasks); err != nil {
		slog.Error("Failed to parse tasks", "error", err)
		return nil, fmt.Errorf("failed to parse tasks: %w", err)
	}
	
	// Get projects for context
	projectsResp, err := t.apiClient.Get(ctx, "/api/v1/projects")
	if err != nil {
		slog.Error("Failed to get projects", "error", err)
		// Continue without projects - not critical
	}
	
	var projects []Project
	if err == nil {
		if err := json.Unmarshal(projectsResp, &projects); err != nil {
			slog.Error("Failed to parse projects", "error", err)
		}
	}
	
	// Analyze tasks
	statusCounts := make(map[string]int)
	overdueTasks := []Task{}
	recentTasks := []Task{}
	projectTaskCounts := make(map[string]int)
	
	now := time.Now()
	dayAgo := now.Add(-24 * time.Hour)
	
	for _, task := range tasks {
		// Count by status
		statusCounts[task.Status]++
		
		// Check if overdue
		if isTaskOverdue(task) {
			overdueTasks = append(overdueTasks, task)
		}
		
		// Check if recent
		if created, err := time.Parse(time.RFC3339, task.CreationDate); err == nil {
			if created.After(dayAgo) {
				recentTasks = append(recentTasks, task)
			}
		}
		
		// Count by project
		if task.ProjectID != nil {
			projectTaskCounts[*task.ProjectID]++
		}
	}
	
	// Build overview
	overview := map[string]any{
		"total_tasks": len(tasks),
		"status_breakdown": statusCounts,
		"overdue_count": len(overdueTasks),
		"overdue_tasks": overdueTasks,
		"recent_activity": map[string]any{
			"tasks_created_24h": len(recentTasks),
			"recent_tasks": recentTasks,
		},
		"project_summary": projectTaskCounts,
		"projects": projects,
	}
	
	// Generate insights
	var insights []string
	
	if len(overdueTasks) > 0 {
		insights = append(insights, fmt.Sprintf("⚠️ %d tasks are overdue and need immediate attention", len(overdueTasks)))
	}
	
	if notStarted, ok := statusCounts["Not Started"]; ok && notStarted > len(tasks)/2 {
		insights = append(insights, "📋 More than half of tasks haven't been started yet")
	}
	
	if inProgress, ok := statusCounts["In Progress"]; ok && inProgress > 5 {
		insights = append(insights, fmt.Sprintf("🔄 %d tasks are currently in progress - consider if any are blocked", inProgress))
	}
	
	if len(recentTasks) > 10 {
		insights = append(insights, "📈 High activity: many new tasks created in the last 24 hours")
	}
	
	overview["insights"] = insights
	
	// Build response text
	responseText := fmt.Sprintf(`Task Overview Dashboard
=====================

Total Tasks: %d

Status Breakdown:
`, len(tasks))
	
	for status, count := range statusCounts {
		responseText += fmt.Sprintf("- %s: %d\n", status, count)
	}
	
	if len(overdueTasks) > 0 {
		responseText += fmt.Sprintf("\n⚠️ Overdue Tasks (%d):\n", len(overdueTasks))
		for _, task := range overdueTasks {
			responseText += fmt.Sprintf("- %s (Due: %s)\n", task.TaskName, *task.DueDate)
		}
	}
	
	responseText += fmt.Sprintf("\n📊 Recent Activity:\n- Tasks created in last 24h: %d\n", len(recentTasks))
	
	if len(insights) > 0 {
		responseText += "\n💡 Insights:\n"
		for _, insight := range insights {
			responseText += fmt.Sprintf("- %s\n", insight)
		}
	}
	
	slog.Info("Task overview generated", "total_tasks", len(tasks), "overdue", len(overdueTasks))
	
	return &mcp.CallToolResultFor[map[string]any]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: responseText,
			},
		},
		Meta: overview,
	}, nil
}

// HandleCreateTaskWithContext implements the create_task_with_context tool
func (t *TaskTools) HandleCreateTaskWithContext(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.CallToolParamsFor[CreateTaskWithContextParams],
) (*mcp.CallToolResultFor[map[string]any], error) {
	slog.Info("Executing create_task_with_context tool", "params", params.Arguments)
	
	// Validate required fields
	if params.Arguments.TaskName == "" {
		return nil, fmt.Errorf("task_name is required")
	}
	if params.Arguments.InitialNote == "" {
		return nil, fmt.Errorf("initial_note is required")
	}
	if params.Arguments.CreatedBy == "" {
		return nil, fmt.Errorf("created_by is required")
	}
	
	// Parse due date if provided
	var dueDate *time.Time
	if params.Arguments.DueDate != "" {
		parsed, err := parseDueDate(params.Arguments.DueDate)
		if err != nil {
			slog.Warn("Failed to parse due date", "due_date", params.Arguments.DueDate, "error", err)
		} else {
			dueDate = parsed
		}
	}
	
	// Build task creation request
	taskRequest := map[string]interface{}{
		"task_name": params.Arguments.TaskName,
		"created_by": params.Arguments.CreatedBy,
	}
	
	if params.Arguments.TaskDescription != "" {
		taskRequest["task_description"] = params.Arguments.TaskDescription
	}
	if params.Arguments.Status != "" {
		taskRequest["status"] = params.Arguments.Status
	} else {
		taskRequest["status"] = "Not Started"
	}
	if params.Arguments.Priority != "" {
		taskRequest["priority"] = params.Arguments.Priority
	}
	if params.Arguments.AssignedTo != "" {
		taskRequest["assigned_to"] = params.Arguments.AssignedTo
	}
	if params.Arguments.ProjectID != "" {
		taskRequest["project_id"] = params.Arguments.ProjectID
	}
	if dueDate != nil {
		taskRequest["due_date"] = dueDate.Format(time.RFC3339)
	}
	
	// Create the task
	taskResp, err := t.apiClient.Post(ctx, "/api/v1/tasks", taskRequest)
	if err != nil {
		slog.Error("Failed to create task", "error", err)
		return nil, fmt.Errorf("failed to create task: %w", err)
	}
	
	var createdTask Task
	if err := json.Unmarshal(taskResp, &createdTask); err != nil {
		slog.Error("Failed to parse created task", "error", err)
		return nil, fmt.Errorf("failed to parse created task: %w", err)
	}
	
	// Add initial planning note
	noteRequest := map[string]interface{}{
		"note": params.Arguments.InitialNote,
		"created_by": params.Arguments.CreatedBy,
	}
	
	noteResp, err := t.apiClient.Post(ctx, fmt.Sprintf("/api/v1/tasks/%s/notes", createdTask.TaskID), noteRequest)
	if err != nil {
		slog.Error("Failed to create initial note", "error", err, "task_id", createdTask.TaskID)
		// Note creation failed, but task was created - return partial success
	}
	
	var createdNote TaskNote
	if err == nil {
		if err := json.Unmarshal(noteResp, &createdNote); err != nil {
			slog.Error("Failed to parse created note", "error", err)
		}
	}
	
	// Generate next steps based on task details
	nextSteps := []string{
		"✅ Task created successfully",
	}
	
	if createdTask.AssignedTo == nil || *createdTask.AssignedTo == "" {
		nextSteps = append(nextSteps, "👤 Assign the task to a team member")
	}
	
	if createdTask.Priority == nil || *createdTask.Priority == "" {
		nextSteps = append(nextSteps, "🎯 Set task priority (Low/Medium/High)")
	}
	
	if createdTask.DueDate == nil {
		nextSteps = append(nextSteps, "📅 Set a due date for the task")
	}
	
	if createdTask.ProjectID == nil {
		nextSteps = append(nextSteps, "📁 Consider associating with a project")
	}
	
	if createdTask.Tags == nil || len(createdTask.Tags) == 0 {
		nextSteps = append(nextSteps, "🏷️ Add relevant tags for better organization")
	}
	
	// Build response
	result := map[string]any{
		"task": createdTask,
		"initial_note": createdNote,
		"next_steps": nextSteps,
		"success": true,
	}
	
	// Build response text
	responseText := fmt.Sprintf(`Task Created Successfully
========================

Task: %s
ID: %s
Status: %s
`, createdTask.TaskName, createdTask.TaskID, createdTask.Status)
	
	if createdTask.Priority != nil {
		responseText += fmt.Sprintf("Priority: %s\n", *createdTask.Priority)
	}
	
	if createdTask.AssignedTo != nil {
		responseText += fmt.Sprintf("Assigned to: %s\n", *createdTask.AssignedTo)
	}
	
	if createdTask.DueDate != nil {
		responseText += fmt.Sprintf("Due Date: %s\n", *createdTask.DueDate)
	}
	
	if createdTask.ProjectID != nil {
		responseText += fmt.Sprintf("Project ID: %s\n", *createdTask.ProjectID)
	}
	
	responseText += fmt.Sprintf("\nInitial Note Added:\n%s\n", params.Arguments.InitialNote)
	
	responseText += "\n📋 Suggested Next Steps:\n"
	for _, step := range nextSteps {
		responseText += fmt.Sprintf("- %s\n", step)
	}
	
	slog.Info("Task created with context", "task_id", createdTask.TaskID, "has_note", err == nil)
	
	return &mcp.CallToolResultFor[map[string]any]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: responseText,
			},
		},
		Meta: result,
	}, nil
}

// HandleGetTaskDetails implements the get_task_details tool
func (t *TaskTools) HandleGetTaskDetails(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.CallToolParamsFor[GetTaskDetailsParams],
) (*mcp.CallToolResultFor[map[string]any], error) {
	slog.Info("Executing get_task_details tool", "params", params.Arguments)
	
	// Validate required fields
	if params.Arguments.TaskID == "" {
		return nil, fmt.Errorf("task_id is required")
	}
	
	// Get task details
	taskResp, err := t.apiClient.Get(ctx, fmt.Sprintf("/api/v1/tasks/%s", params.Arguments.TaskID))
	if err != nil {
		slog.Error("Failed to get task", "error", err, "task_id", params.Arguments.TaskID)
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	
	var task Task
	if err := json.Unmarshal(taskResp, &task); err != nil {
		slog.Error("Failed to parse task", "error", err)
		return nil, fmt.Errorf("failed to parse task: %w", err)
	}
	
	// Get task notes
	notesResp, err := t.apiClient.Get(ctx, fmt.Sprintf("/api/v1/tasks/%s/notes", params.Arguments.TaskID))
	if err != nil {
		slog.Error("Failed to get task notes", "error", err, "task_id", params.Arguments.TaskID)
		// Continue without notes - not critical for task details
	}
	
	var notes []TaskNote
	if err == nil {
		if err := json.Unmarshal(notesResp, &notes); err != nil {
			slog.Error("Failed to parse task notes", "error", err)
		}
	}
	
	// Get project details if task has a project
	var project *Project
	if task.ProjectID != nil && *task.ProjectID != "" {
		projectResp, err := t.apiClient.Get(ctx, fmt.Sprintf("/api/v1/projects/%s", *task.ProjectID))
		if err != nil {
			slog.Error("Failed to get project", "error", err, "project_id", *task.ProjectID)
			// Continue without project - not critical
		} else {
			var proj Project
			if err := json.Unmarshal(projectResp, &proj); err != nil {
				slog.Error("Failed to parse project", "error", err)
			} else {
				project = &proj
			}
		}
	}
	
	// Analyze task for insights
	var insights []string
	
	// Check if task is overdue
	if isTaskOverdue(task) {
		insights = append(insights, "⚠️ This task is overdue and needs immediate attention")
	}
	
	// Check if task has been idle
	if task.LastUpdateDate != nil {
		lastUpdate, err := time.Parse(time.RFC3339, *task.LastUpdateDate)
		if err == nil && time.Since(lastUpdate) > 7*24*time.Hour {
			insights = append(insights, "📅 Task hasn't been updated in over a week")
		}
	}
	
	// Check completion criteria
	if task.Status == "In Progress" && len(notes) == 0 {
		insights = append(insights, "📝 Consider adding progress notes to track work")
	}
	
	if task.Priority == nil || *task.Priority == "" {
		insights = append(insights, "🎯 Task priority is not set")
	}
	
	if task.AssignedTo == nil || *task.AssignedTo == "" {
		insights = append(insights, "👤 Task is not assigned to anyone")
	}
	
	if task.DueDate == nil {
		insights = append(insights, "📅 No due date set for this task")
	}
	
	// Check if task is blocked
	if task.Status == "Blocked" && len(notes) > 0 {
		insights = append(insights, "🚫 Task is blocked - check latest notes for blocker details")
	}
	
	// Generate suggested next actions
	var nextActions []string
	
	switch task.Status {
	case "Not Started":
		nextActions = append(nextActions, "▶️ Move task to 'In Progress' when work begins")
		if task.AssignedTo == nil {
			nextActions = append(nextActions, "👤 Assign task to team member")
		}
	case "In Progress":
		nextActions = append(nextActions, "📝 Add progress notes to document current work")
		nextActions = append(nextActions, "🔄 Update status if task is complete or blocked")
	case "Blocked":
		nextActions = append(nextActions, "🔓 Resolve blocker and update status")
		nextActions = append(nextActions, "📋 Document blocker resolution in notes")
	case "Review":
		nextActions = append(nextActions, "✅ Complete review and mark as complete or provide feedback")
	case "Complete":
		nextActions = append(nextActions, "🎉 Task is complete - consider archiving if no longer needed")
	}
	
	// Build comprehensive response
	result := map[string]any{
		"task":        task,
		"notes":       notes,
		"project":     project,
		"insights":    insights,
		"next_actions": nextActions,
		"note_count":  len(notes),
		"has_project": project != nil,
	}
	
	// Build detailed response text
	responseText := fmt.Sprintf(`Task Details\n============\n\nTask: %s\nID: %s\nStatus: %s\n`, 
		task.TaskName, task.TaskID, task.Status)
	
	if task.TaskDescription != nil && *task.TaskDescription != "" {
		responseText += fmt.Sprintf("Description: %s\n", *task.TaskDescription)
	}
	
	if task.Priority != nil {
		responseText += fmt.Sprintf("Priority: %s\n", *task.Priority)
	}
	
	if task.AssignedTo != nil {
		responseText += fmt.Sprintf("Assigned to: %s\n", *task.AssignedTo)
	}
	
	if task.DueDate != nil {
		responseText += fmt.Sprintf("Due Date: %s\n", *task.DueDate)
	}
	
	if task.StartDate != nil {
		responseText += fmt.Sprintf("Start Date: %s\n", *task.StartDate)
	}
	
	if task.CompletionDate != nil {
		responseText += fmt.Sprintf("Completion Date: %s\n", *task.CompletionDate)
	}
	
	if len(task.Tags) > 0 {
		responseText += fmt.Sprintf("Tags: %v\n", task.Tags)
	}
	
	if project != nil {
		responseText += fmt.Sprintf("\n📁 Project: %s\nProject ID: %s\n", 
			project.ProjectName, project.ProjectID)
		if project.ProjectDescription != nil {
			responseText += fmt.Sprintf("Project Description: %s\n", *project.ProjectDescription)
		}
	}
	
	responseText += fmt.Sprintf("\nCreated by: %s\nCreated: %s\n", 
		task.CreatedBy, task.CreationDate)
	
	if task.LastUpdatedBy != nil {
		responseText += fmt.Sprintf("Last updated by: %s\nLast updated: %s\n", 
			*task.LastUpdatedBy, *task.LastUpdateDate)
	}
	
	if len(notes) > 0 {
		responseText += fmt.Sprintf("\n📝 Notes (%d):\n", len(notes))
		for i, note := range notes {
			if i < 5 { // Show only latest 5 notes
				responseText += fmt.Sprintf("- [%s] %s (by %s)\n", 
					note.CreationDate, note.Note, note.CreatedBy)
			}
		}
		if len(notes) > 5 {
			responseText += fmt.Sprintf("... and %d more notes\n", len(notes)-5)
		}
	} else {
		responseText += "\n📝 No notes available\n"
	}
	
	if len(insights) > 0 {
		responseText += "\n💡 Insights:\n"
		for _, insight := range insights {
			responseText += fmt.Sprintf("- %s\n", insight)
		}
	}
	
	if len(nextActions) > 0 {
		responseText += "\n📋 Suggested Next Actions:\n"
		for _, action := range nextActions {
			responseText += fmt.Sprintf("- %s\n", action)
		}
	}
	
	slog.Info("Task details retrieved", "task_id", task.TaskID, "note_count", len(notes), "has_project", project != nil)
	
	return &mcp.CallToolResultFor[map[string]any]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: responseText,
			},
		},
		Meta: result,
	}, nil
}

// HandleUpdateTaskProgress implements the update_task_progress tool
func (t *TaskTools) HandleUpdateTaskProgress(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.CallToolParamsFor[UpdateTaskProgressParams],
) (*mcp.CallToolResultFor[map[string]any], error) {
	slog.Info("Executing update_task_progress tool", "params", params.Arguments)
	
	// Validate required fields
	if params.Arguments.TaskID == "" {
		return nil, fmt.Errorf("task_id is required")
	}
	if params.Arguments.ProgressNote == "" {
		return nil, fmt.Errorf("progress_note is required")
	}
	if params.Arguments.UpdatedBy == "" {
		return nil, fmt.Errorf("updated_by is required")
	}
	
	// Get current task state first
	taskResp, err := t.apiClient.Get(ctx, fmt.Sprintf("/api/v1/tasks/%s", params.Arguments.TaskID))
	if err != nil {
		slog.Error("Failed to get current task", "error", err, "task_id", params.Arguments.TaskID)
		return nil, fmt.Errorf("failed to get current task: %w", err)
	}
	
	var currentTask Task
	if err := json.Unmarshal(taskResp, &currentTask); err != nil {
		slog.Error("Failed to parse current task", "error", err)
		return nil, fmt.Errorf("failed to parse current task: %w", err)
	}
	
	// Build task update request
	updateRequest := map[string]interface{}{
		"last_updated_by": params.Arguments.UpdatedBy,
	}
	
	// Track changes for the response
	changes := []string{}
	
	if params.Arguments.Status != "" && params.Arguments.Status != currentTask.Status {
		updateRequest["status"] = params.Arguments.Status
		changes = append(changes, fmt.Sprintf("Status: %s → %s", currentTask.Status, params.Arguments.Status))
	}
	
	if params.Arguments.Priority != "" {
		var currentPriority string
		if currentTask.Priority != nil {
			currentPriority = *currentTask.Priority
		}
		if params.Arguments.Priority != currentPriority {
			updateRequest["priority"] = params.Arguments.Priority
			changes = append(changes, fmt.Sprintf("Priority: %s → %s", currentPriority, params.Arguments.Priority))
		}
	}
	
	if params.Arguments.AssignedTo != "" {
		var currentAssignee string
		if currentTask.AssignedTo != nil {
			currentAssignee = *currentTask.AssignedTo
		}
		if params.Arguments.AssignedTo != currentAssignee {
			updateRequest["assigned_to"] = params.Arguments.AssignedTo
			changes = append(changes, fmt.Sprintf("Assigned to: %s → %s", currentAssignee, params.Arguments.AssignedTo))
		}
	}
	
	// Set completion date if status is Complete
	if params.Arguments.Status == "Complete" {
		updateRequest["completion_date"] = time.Now().Format(time.RFC3339)
		changes = append(changes, "Completion date set")
	}
	
	// Set start date if status changed to In Progress and no start date exists
	if params.Arguments.Status == "In Progress" && currentTask.StartDate == nil {
		updateRequest["start_date"] = time.Now().Format(time.RFC3339)
		changes = append(changes, "Start date set")
	}
	
	// Update the task if there are changes
	var updatedTask Task
	if len(changes) > 0 {
		updateResp, err := t.apiClient.Put(ctx, fmt.Sprintf("/api/v1/tasks/%s", params.Arguments.TaskID), updateRequest)
		if err != nil {
			slog.Error("Failed to update task", "error", err, "task_id", params.Arguments.TaskID)
			return nil, fmt.Errorf("failed to update task: %w", err)
		}
		
		if err := json.Unmarshal(updateResp, &updatedTask); err != nil {
			slog.Error("Failed to parse updated task", "error", err)
			return nil, fmt.Errorf("failed to parse updated task: %w", err)
		}
	} else {
		updatedTask = currentTask
	}
	
	// Add progress note
	noteRequest := map[string]interface{}{
		"note":       params.Arguments.ProgressNote,
		"created_by": params.Arguments.UpdatedBy,
	}
	
	noteResp, err := t.apiClient.Post(ctx, fmt.Sprintf("/api/v1/tasks/%s/notes", params.Arguments.TaskID), noteRequest)
	if err != nil {
		slog.Error("Failed to create progress note", "error", err, "task_id", params.Arguments.TaskID)
		// Continue - task update succeeded even if note failed
	}
	
	var createdNote TaskNote
	if err == nil {
		if err := json.Unmarshal(noteResp, &createdNote); err != nil {
			slog.Error("Failed to parse created note", "error", err)
		}
	}
	
	// Generate insights based on the update
	var insights []string
	
	if params.Arguments.Status == "Complete" {
		insights = append(insights, "🎉 Task marked as complete!")
		
		// Check completion time
		if currentTask.DueDate != nil {
			dueDate, err := time.Parse(time.RFC3339, *currentTask.DueDate)
			if err == nil {
				if time.Now().Before(dueDate) {
					insights = append(insights, "✅ Task completed before due date")
				} else {
					insights = append(insights, "⏰ Task completed after due date")
				}
			}
		}
	}
	
	if params.Arguments.Status == "Blocked" {
		insights = append(insights, "🚫 Task is now blocked - ensure blocker is documented in the note")
	}
	
	if params.Arguments.Status == "In Progress" && currentTask.Status == "Not Started" {
		insights = append(insights, "▶️ Work has begun on this task")
	}
	
	if params.Arguments.Priority == "High" && (currentTask.Priority == nil || *currentTask.Priority != "High") {
		insights = append(insights, "🔥 Task priority elevated to High")
	}
	
	// Generate next steps based on new status
	var nextSteps []string
	
	switch params.Arguments.Status {
	case "In Progress":
		nextSteps = append(nextSteps, "📝 Continue adding progress notes as work proceeds")
		nextSteps = append(nextSteps, "🔄 Update status to 'Review' or 'Complete' when ready")
	case "Blocked":
		nextSteps = append(nextSteps, "🔓 Work on resolving the blocker")
		nextSteps = append(nextSteps, "👥 Consider escalating if blocker persists")
	case "Review":
		nextSteps = append(nextSteps, "👀 Assign reviewer or notify stakeholders")
		nextSteps = append(nextSteps, "📋 Prepare review criteria/checklist")
	case "Complete":
		nextSteps = append(nextSteps, "🏷️ Consider archiving if no longer needed")
		nextSteps = append(nextSteps, "📊 Update project metrics if applicable")
	}
	
	// Build comprehensive response
	result := map[string]any{
		"task":           updatedTask,
		"progress_note":  createdNote,
		"changes_made":   changes,
		"insights":       insights,
		"next_steps":     nextSteps,
		"update_success": true,
		"note_added":     err == nil,
	}
	
	// Build response text
	responseText := fmt.Sprintf(`Task Progress Updated\n====================\n\nTask: %s\nID: %s\n`, 
		updatedTask.TaskName, updatedTask.TaskID)
	
	if len(changes) > 0 {
		responseText += "\n📊 Changes Made:\n"
		for _, change := range changes {
			responseText += fmt.Sprintf("- %s\n", change)
		}
	} else {
		responseText += "\n📝 No field changes made (progress note added)\n"
	}
	
	responseText += fmt.Sprintf("\n📝 Progress Note Added:\n%s\n", params.Arguments.ProgressNote)
	responseText += fmt.Sprintf("Added by: %s\n", params.Arguments.UpdatedBy)
	
	if len(insights) > 0 {
		responseText += "\n💡 Insights:\n"
		for _, insight := range insights {
			responseText += fmt.Sprintf("- %s\n", insight)
		}
	}
	
	if len(nextSteps) > 0 {
		responseText += "\n📋 Suggested Next Steps:\n"
		for _, step := range nextSteps {
			responseText += fmt.Sprintf("- %s\n", step)
		}
	}
	
	responseText += fmt.Sprintf("\nCurrent Status: %s\n", updatedTask.Status)
	if updatedTask.Priority != nil {
		responseText += fmt.Sprintf("Priority: %s\n", *updatedTask.Priority)
	}
	if updatedTask.AssignedTo != nil {
		responseText += fmt.Sprintf("Assigned to: %s\n", *updatedTask.AssignedTo)
	}
	
	slog.Info("Task progress updated", "task_id", updatedTask.TaskID, "changes", len(changes), "note_added", err == nil)
	
	return &mcp.CallToolResultFor[map[string]any]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: responseText,
			},
		},
		Meta: result,
	}, nil
}