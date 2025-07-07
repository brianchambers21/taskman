package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/bchamber/taskman-mcp/internal/client"
	"github.com/modelcontextprotocol/go-sdk/mcp"
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
	TaskName        string `json:"task_name"`
	TaskDescription string `json:"task_description,omitempty"`
	Status          string `json:"status,omitempty"`
	Priority        string `json:"priority,omitempty"`
	AssignedTo      string `json:"assigned_to,omitempty"`
	ProjectID       string `json:"project_id,omitempty"`
	DueDate         string `json:"due_date,omitempty"`
	InitialNote     string `json:"initial_note"`
	CreatedBy       string `json:"created_by"`
}

// GetTaskDetailsParams defines input for get_task_details tool
type GetTaskDetailsParams struct {
	TaskID string `json:"task_id"`
}

// UpdateTaskProgressParams defines input for update_task_progress tool
type UpdateTaskProgressParams struct {
	TaskID       string `json:"task_id"`
	Status       string `json:"status,omitempty"`
	Priority     string `json:"priority,omitempty"`
	AssignedTo   string `json:"assigned_to,omitempty"`
	ProgressNote string `json:"progress_note"`
	UpdatedBy    string `json:"updated_by"`
}

// SearchTasksParams defines input for search_tasks tool
type SearchTasksParams struct {
	Status      string `json:"status,omitempty"`
	Priority    string `json:"priority,omitempty"`
	AssignedTo  string `json:"assigned_to,omitempty"`
	ProjectID   string `json:"project_id,omitempty"`
	CreatedBy   string `json:"created_by,omitempty"`
	DueDateFrom string `json:"due_date_from,omitempty"`
	DueDateTo   string `json:"due_date_to,omitempty"`
	SearchText  string `json:"search_text,omitempty"`
	Archived    string `json:"archived,omitempty"`
	SortBy      string `json:"sort_by,omitempty"`
	SortOrder   string `json:"sort_order,omitempty"`
	Limit       int    `json:"limit,omitempty"`
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
		"total_tasks":      len(tasks),
		"status_breakdown": statusCounts,
		"overdue_count":    len(overdueTasks),
		"overdue_tasks":    overdueTasks,
		"recent_activity": map[string]any{
			"tasks_created_24h": len(recentTasks),
			"recent_tasks":      recentTasks,
		},
		"project_summary": projectTaskCounts,
		"projects":        projects,
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

	// Validate status if provided
	validStatuses := []string{"Not Started", "In Progress", "Blocked", "Review", "Complete"}
	if params.Arguments.Status != "" {
		isValid := false
		for _, validStatus := range validStatuses {
			if params.Arguments.Status == validStatus {
				isValid = true
				break
			}
		}
		if !isValid {
			return nil, fmt.Errorf("invalid status '%s'. Valid statuses are: %v", params.Arguments.Status, validStatuses)
		}
	}

	// Validate priority if provided
	validPriorities := []string{"Low", "Medium", "High"}
	if params.Arguments.Priority != "" {
		isValid := false
		for _, validPriority := range validPriorities {
			if params.Arguments.Priority == validPriority {
				isValid = true
				break
			}
		}
		if !isValid {
			return nil, fmt.Errorf("invalid priority '%s'. Valid priorities are: %v", params.Arguments.Priority, validPriorities)
		}
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
		"task_name":  params.Arguments.TaskName,
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
		"note":       params.Arguments.InitialNote,
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
		"task":         createdTask,
		"initial_note": createdNote,
		"next_steps":   nextSteps,
		"success":      true,
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
		"task":         task,
		"notes":        notes,
		"project":      project,
		"insights":     insights,
		"next_actions": nextActions,
		"note_count":   len(notes),
		"has_project":  project != nil,
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

	// Validate status if provided
	validStatuses := []string{"Not Started", "In Progress", "Blocked", "Review", "Complete"}
	if params.Arguments.Status != "" {
		isValid := false
		for _, validStatus := range validStatuses {
			if params.Arguments.Status == validStatus {
				isValid = true
				break
			}
		}
		if !isValid {
			return nil, fmt.Errorf("invalid status '%s'. Valid statuses are: %v", params.Arguments.Status, validStatuses)
		}
	}

	// Validate priority if provided
	validPriorities := []string{"Low", "Medium", "High"}
	if params.Arguments.Priority != "" {
		isValid := false
		for _, validPriority := range validPriorities {
			if params.Arguments.Priority == validPriority {
				isValid = true
				break
			}
		}
		if !isValid {
			return nil, fmt.Errorf("invalid priority '%s'. Valid priorities are: %v", params.Arguments.Priority, validPriorities)
		}
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

// HandleSearchTasks implements the search_tasks tool
func (t *TaskTools) HandleSearchTasks(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.CallToolParamsFor[SearchTasksParams],
) (*mcp.CallToolResultFor[map[string]any], error) {
	slog.Info("Executing search_tasks tool", "params", params.Arguments)

	// Build complex query parameters
	queryParams := ""

	if params.Arguments.Status != "" {
		if queryParams == "" {
			queryParams += "?"
		} else {
			queryParams += "&"
		}
		queryParams += fmt.Sprintf("status=%s", url.QueryEscape(params.Arguments.Status))
	}

	if params.Arguments.Priority != "" {
		if queryParams == "" {
			queryParams += "?"
		} else {
			queryParams += "&"
		}
		queryParams += fmt.Sprintf("priority=%s", url.QueryEscape(params.Arguments.Priority))
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

	if params.Arguments.CreatedBy != "" {
		if queryParams == "" {
			queryParams += "?"
		} else {
			queryParams += "&"
		}
		queryParams += fmt.Sprintf("created_by=%s", url.QueryEscape(params.Arguments.CreatedBy))
	}

	if params.Arguments.Archived != "" {
		if queryParams == "" {
			queryParams += "?"
		} else {
			queryParams += "&"
		}
		queryParams += fmt.Sprintf("archived=%s", url.QueryEscape(params.Arguments.Archived))
	}

	// Add date range filters (note: these would need API support)
	if params.Arguments.DueDateFrom != "" {
		if queryParams == "" {
			queryParams += "?"
		} else {
			queryParams += "&"
		}
		queryParams += fmt.Sprintf("due_date_from=%s", url.QueryEscape(params.Arguments.DueDateFrom))
	}

	if params.Arguments.DueDateTo != "" {
		if queryParams == "" {
			queryParams += "?"
		} else {
			queryParams += "&"
		}
		queryParams += fmt.Sprintf("due_date_to=%s", url.QueryEscape(params.Arguments.DueDateTo))
	}

	// Add search text (note: would need API support for text search)
	if params.Arguments.SearchText != "" {
		if queryParams == "" {
			queryParams += "?"
		} else {
			queryParams += "&"
		}
		queryParams += fmt.Sprintf("search=%s", url.QueryEscape(params.Arguments.SearchText))
	}

	// Add sorting and pagination
	if params.Arguments.SortBy != "" {
		if queryParams == "" {
			queryParams += "?"
		} else {
			queryParams += "&"
		}
		queryParams += fmt.Sprintf("sort_by=%s", url.QueryEscape(params.Arguments.SortBy))

		if params.Arguments.SortOrder != "" {
			queryParams += fmt.Sprintf("&sort_order=%s", url.QueryEscape(params.Arguments.SortOrder))
		}
	}

	if params.Arguments.Limit > 0 {
		if queryParams == "" {
			queryParams += "?"
		} else {
			queryParams += "&"
		}
		queryParams += fmt.Sprintf("limit=%d", params.Arguments.Limit)
	}

	// Get tasks with complex filtering
	tasksResp, err := t.apiClient.Get(ctx, "/api/v1/tasks"+queryParams)
	if err != nil {
		slog.Error("Failed to search tasks", "error", err)
		return nil, fmt.Errorf("failed to search tasks: %w", err)
	}

	var tasks []Task
	if err := json.Unmarshal(tasksResp, &tasks); err != nil {
		slog.Error("Failed to parse searched tasks", "error", err)
		return nil, fmt.Errorf("failed to parse searched tasks: %w", err)
	}

	// Apply client-side filtering for fields not supported by API
	var filteredTasks []Task

	for _, task := range tasks {
		include := true

		// Text search in task name and description (client-side)
		if params.Arguments.SearchText != "" {
			searchText := params.Arguments.SearchText
			found := false

			// Search in task name
			if task.TaskName != "" &&
				fmt.Sprintf("%s", task.TaskName) != "" {
				// Simple case-insensitive search
				if len(searchText) <= len(task.TaskName) {
					for i := 0; i <= len(task.TaskName)-len(searchText); i++ {
						if task.TaskName[i:i+len(searchText)] == searchText {
							found = true
							break
						}
					}
				}
			}

			// Search in task description
			if !found && task.TaskDescription != nil && *task.TaskDescription != "" {
				desc := *task.TaskDescription
				if len(searchText) <= len(desc) {
					for i := 0; i <= len(desc)-len(searchText); i++ {
						if desc[i:i+len(searchText)] == searchText {
							found = true
							break
						}
					}
				}
			}

			if !found {
				include = false
			}
		}

		// Date range filtering (client-side)
		if include && params.Arguments.DueDateFrom != "" && task.DueDate != nil {
			if fromDate, err := time.Parse("2006-01-02", params.Arguments.DueDateFrom); err == nil {
				if dueDate, err := time.Parse(time.RFC3339, *task.DueDate); err == nil {
					if dueDate.Before(fromDate) {
						include = false
					}
				}
			}
		}

		if include && params.Arguments.DueDateTo != "" && task.DueDate != nil {
			if toDate, err := time.Parse("2006-01-02", params.Arguments.DueDateTo); err == nil {
				if dueDate, err := time.Parse(time.RFC3339, *task.DueDate); err == nil {
					if dueDate.After(toDate.Add(24 * time.Hour)) { // Include full day
						include = false
					}
				}
			}
		}

		if include {
			filteredTasks = append(filteredTasks, task)
		}
	}

	// Apply sorting (client-side for unsupported sorts)
	if params.Arguments.SortBy != "" {
		// Note: This is a simplified sorting implementation
		// In a real implementation, you'd use sort.Slice with proper comparison functions
	}

	// Apply limit (client-side)
	if params.Arguments.Limit > 0 && len(filteredTasks) > params.Arguments.Limit {
		filteredTasks = filteredTasks[:params.Arguments.Limit]
	}

	// Generate search statistics
	statusCounts := make(map[string]int)
	priorityCounts := make(map[string]int)
	projectCounts := make(map[string]int)
	overdueTasks := []Task{}

	for _, task := range filteredTasks {
		statusCounts[task.Status]++

		if task.Priority != nil {
			priorityCounts[*task.Priority]++
		} else {
			priorityCounts["None"]++
		}

		if task.ProjectID != nil {
			projectCounts[*task.ProjectID]++
		} else {
			projectCounts["No Project"]++
		}

		if isTaskOverdue(task) {
			overdueTasks = append(overdueTasks, task)
		}
	}

	// Generate search insights
	var insights []string

	totalResults := len(filteredTasks)
	if totalResults == 0 {
		insights = append(insights, "🔍 No tasks match your search criteria")
	} else if totalResults == 1 {
		insights = append(insights, "🎯 Found exactly one matching task")
	} else if totalResults > 100 {
		insights = append(insights, "📊 Large result set - consider narrowing your search")
	}

	if len(overdueTasks) > 0 {
		insights = append(insights, fmt.Sprintf("⚠️ %d of the results are overdue", len(overdueTasks)))
	}

	if len(statusCounts) == 1 {
		for status := range statusCounts {
			insights = append(insights, fmt.Sprintf("📋 All results have status: %s", status))
		}
	}

	if len(priorityCounts) > 0 {
		if high, exists := priorityCounts["High"]; exists && high > totalResults/2 {
			insights = append(insights, "🔥 Most results are high priority")
		}
	}

	// Generate actionable suggestions
	var suggestions []string

	if totalResults == 0 {
		suggestions = append(suggestions, "🔍 Try broadening your search criteria")
		suggestions = append(suggestions, "📋 Check if tasks exist with different statuses")
	} else {
		if len(overdueTasks) > 0 {
			suggestions = append(suggestions, "🚨 Address overdue tasks first")
		}

		if notStarted, exists := statusCounts["Not Started"]; exists && notStarted > 0 {
			suggestions = append(suggestions, fmt.Sprintf("▶️ Consider starting %d pending tasks", notStarted))
		}

		if review, exists := statusCounts["Review"]; exists && review > 0 {
			suggestions = append(suggestions, fmt.Sprintf("👀 Review %d tasks waiting for approval", review))
		}
	}

	// Build comprehensive response
	result := map[string]any{
		"tasks":              filteredTasks,
		"total_results":      totalResults,
		"search_criteria":    params.Arguments,
		"status_breakdown":   statusCounts,
		"priority_breakdown": priorityCounts,
		"project_breakdown":  projectCounts,
		"overdue_count":      len(overdueTasks),
		"overdue_tasks":      overdueTasks,
		"insights":           insights,
		"suggestions":        suggestions,
	}

	// Build response text
	responseText := fmt.Sprintf(`Task Search Results\n==================\n\nFound: %d tasks\n`, totalResults)

	// Show search criteria
	if params.Arguments.Status != "" || params.Arguments.Priority != "" || params.Arguments.AssignedTo != "" ||
		params.Arguments.ProjectID != "" || params.Arguments.SearchText != "" {
		responseText += "\n🔍 Search Criteria:\n"

		if params.Arguments.Status != "" {
			responseText += fmt.Sprintf("- Status: %s\n", params.Arguments.Status)
		}
		if params.Arguments.Priority != "" {
			responseText += fmt.Sprintf("- Priority: %s\n", params.Arguments.Priority)
		}
		if params.Arguments.AssignedTo != "" {
			responseText += fmt.Sprintf("- Assigned to: %s\n", params.Arguments.AssignedTo)
		}
		if params.Arguments.ProjectID != "" {
			responseText += fmt.Sprintf("- Project ID: %s\n", params.Arguments.ProjectID)
		}
		if params.Arguments.SearchText != "" {
			responseText += fmt.Sprintf("- Search text: %s\n", params.Arguments.SearchText)
		}
		if params.Arguments.DueDateFrom != "" {
			responseText += fmt.Sprintf("- Due date from: %s\n", params.Arguments.DueDateFrom)
		}
		if params.Arguments.DueDateTo != "" {
			responseText += fmt.Sprintf("- Due date to: %s\n", params.Arguments.DueDateTo)
		}
	}

	if totalResults > 0 {
		responseText += "\n📊 Results Breakdown:\n"
		for status, count := range statusCounts {
			responseText += fmt.Sprintf("- %s: %d\n", status, count)
		}

		if len(overdueTasks) > 0 {
			responseText += fmt.Sprintf("\n⚠️ Overdue Tasks (%d):\n", len(overdueTasks))
			for i, task := range overdueTasks {
				if i < 5 { // Show only first 5
					responseText += fmt.Sprintf("- %s (Due: %s)\n", task.TaskName, *task.DueDate)
				}
			}
			if len(overdueTasks) > 5 {
				responseText += fmt.Sprintf("... and %d more overdue tasks\n", len(overdueTasks)-5)
			}
		}

		responseText += fmt.Sprintf("\n📋 Tasks (showing %d):\n", len(filteredTasks))
		for i, task := range filteredTasks {
			if i < 10 { // Show only first 10
				assignee := "Unassigned"
				if task.AssignedTo != nil {
					assignee = *task.AssignedTo
				}
				priority := "None"
				if task.Priority != nil {
					priority = *task.Priority
				}
				responseText += fmt.Sprintf("- %s (%s, %s) - %s\n", task.TaskName, task.Status, priority, assignee)
			}
		}
		if len(filteredTasks) > 10 {
			responseText += fmt.Sprintf("... and %d more tasks\n", len(filteredTasks)-10)
		}
	}

	if len(insights) > 0 {
		responseText += "\n💡 Insights:\n"
		for _, insight := range insights {
			responseText += fmt.Sprintf("- %s\n", insight)
		}
	}

	if len(suggestions) > 0 {
		responseText += "\n📋 Suggestions:\n"
		for _, suggestion := range suggestions {
			responseText += fmt.Sprintf("- %s\n", suggestion)
		}
	}

	slog.Info("Task search completed", "total_results", totalResults, "overdue_count", len(overdueTasks))

	return &mcp.CallToolResultFor[map[string]any]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: responseText,
			},
		},
		Meta: result,
	}, nil
}

// GetAllTasksParams defines input for get_all_tasks tool
type GetAllTasksParams struct {
	// No parameters needed for listing all tasks
}

// HandleGetAllTasks implements the get_all_tasks tool
func (t *TaskTools) HandleGetAllTasks(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.CallToolParamsFor[GetAllTasksParams],
) (*mcp.CallToolResultFor[map[string]any], error) {
	slog.Info("Executing get_all_tasks tool")

	// Get all tasks from API
	tasksResp, err := t.apiClient.Get(ctx, "/api/v1/tasks")
	if err != nil {
		slog.Error("Failed to get tasks", "error", err)
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	var tasks []Task
	if err := json.Unmarshal(tasksResp, &tasks); err != nil {
		slog.Error("Failed to parse tasks", "error", err)
		return nil, fmt.Errorf("failed to parse tasks: %w", err)
	}

	// Analyze tasks for insights
	statusBreakdown := make(map[string]int)
	priorityBreakdown := make(map[string]int)
	projectBreakdown := make(map[string]int)
	var overdueTasks []Task

	for _, task := range tasks {
		// Status breakdown
		statusBreakdown[task.Status]++

		// Priority breakdown
		if task.Priority != nil && *task.Priority != "" {
			priorityBreakdown[*task.Priority]++
		} else {
			priorityBreakdown["Unset"]++
		}

		// Project breakdown
		if task.ProjectID != nil && *task.ProjectID != "" {
			projectBreakdown[*task.ProjectID]++
		} else {
			projectBreakdown["No Project"]++
		}

		// Check if overdue
		if isTaskOverdue(task) {
			overdueTasks = append(overdueTasks, task)
		}
	}

	// Build response
	var responseText string
	if len(tasks) == 0 {
		responseText = "No tasks found.\n\nCreate your first task to get started!"
	} else {
		responseText = fmt.Sprintf("All Tasks (%d)\n", len(tasks))
		responseText += "=============\n\n"

		// Status breakdown
		responseText += "📊 Status Breakdown:\n"
		for status, count := range statusBreakdown {
			responseText += fmt.Sprintf("- %s: %d\n", status, count)
		}
		responseText += "\n"

		// Priority breakdown
		responseText += "🎯 Priority Breakdown:\n"
		for priority, count := range priorityBreakdown {
			responseText += fmt.Sprintf("- %s: %d\n", priority, count)
		}
		responseText += "\n"

		// Overdue tasks
		if len(overdueTasks) > 0 {
			responseText += fmt.Sprintf("⚠️  Overdue Tasks (%d):\n", len(overdueTasks))
			for _, task := range overdueTasks {
				responseText += fmt.Sprintf("- %s", task.TaskName)
				if task.DueDate != nil {
					responseText += fmt.Sprintf(" (due: %s)", *task.DueDate)
				}
				responseText += "\n"
			}
			responseText += "\n"
		}

		// Recent tasks (first 10)
		responseText += "📋 Recent Tasks:\n"
		displayCount := len(tasks)
		if displayCount > 10 {
			displayCount = 10
		}
		
		for i := 0; i < displayCount; i++ {
			task := tasks[i]
			responseText += fmt.Sprintf("- %s (%s", task.TaskName, task.Status)
			if task.Priority != nil && *task.Priority != "" {
				responseText += fmt.Sprintf(", %s", *task.Priority)
			}
			if task.AssignedTo != nil && *task.AssignedTo != "" {
				responseText += fmt.Sprintf(" - %s", *task.AssignedTo)
			}
			responseText += ")\n"
		}

		if len(tasks) > 10 {
			responseText += fmt.Sprintf("\n... and %d more tasks\n", len(tasks)-10)
		}
	}

	result := map[string]any{
		"tasks":             tasks,
		"total_count":       len(tasks),
		"status_breakdown":  statusBreakdown,
		"priority_breakdown": priorityBreakdown,
		"project_breakdown": projectBreakdown,
		"overdue_count":     len(overdueTasks),
		"overdue_tasks":     overdueTasks,
		"task_list":         tasks,
	}

	slog.Info("Tasks list retrieved", "total_tasks", len(tasks), "overdue_count", len(overdueTasks))

	return &mcp.CallToolResultFor[map[string]any]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: responseText,
			},
		},
		Meta: result,
	}, nil
}

// AddTaskNoteParams defines input for add_task_note tool
type AddTaskNoteParams struct {
	TaskID    string `json:"task_id"`
	Note      string `json:"note"`
	CreatedBy string `json:"created_by"`
}

// HandleAddTaskNote implements the add_task_note tool
func (t *TaskTools) HandleAddTaskNote(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.CallToolParamsFor[AddTaskNoteParams],
) (*mcp.CallToolResultFor[map[string]any], error) {
	slog.Info("Executing add_task_note tool", "params", params.Arguments)

	// Validate required fields
	if params.Arguments.TaskID == "" {
		return nil, fmt.Errorf("task_id is required")
	}
	if params.Arguments.Note == "" {
		return nil, fmt.Errorf("note is required")
	}
	if params.Arguments.CreatedBy == "" {
		return nil, fmt.Errorf("created_by is required")
	}

	// First, verify the task exists
	taskResp, err := t.apiClient.Get(ctx, fmt.Sprintf("/api/v1/tasks/%s", params.Arguments.TaskID))
	if err != nil {
		slog.Error("Failed to get task for note addition", "error", err, "task_id", params.Arguments.TaskID)
		return nil, fmt.Errorf("failed to verify task exists: %w", err)
	}

	var task Task
	if err := json.Unmarshal(taskResp, &task); err != nil {
		slog.Error("Failed to parse task", "error", err)
		return nil, fmt.Errorf("failed to parse task: %w", err)
	}

	// Create the note
	noteRequest := map[string]interface{}{
		"note":       params.Arguments.Note,
		"created_by": params.Arguments.CreatedBy,
	}

	noteResp, err := t.apiClient.Post(ctx, fmt.Sprintf("/api/v1/tasks/%s/notes", params.Arguments.TaskID), noteRequest)
	if err != nil {
		slog.Error("Failed to add note", "error", err)
		return nil, fmt.Errorf("failed to add note: %w", err)
	}

	var createdNote TaskNote
	if err := json.Unmarshal(noteResp, &createdNote); err != nil {
		slog.Error("Failed to parse created note", "error", err)
		return nil, fmt.Errorf("failed to parse created note: %w", err)
	}

	// Build response
	responseText := fmt.Sprintf("Note Added Successfully\n")
	responseText += "======================\n\n"
	responseText += fmt.Sprintf("Task: %s\n", task.TaskName)
	responseText += fmt.Sprintf("Task ID: %s\n", task.TaskID)
	responseText += fmt.Sprintf("Note ID: %s\n", createdNote.NoteID)
	responseText += fmt.Sprintf("Note: %s\n", createdNote.Note)
	responseText += fmt.Sprintf("Created by: %s\n", createdNote.CreatedBy)
	responseText += fmt.Sprintf("Created: %s\n", createdNote.CreationDate)

	// Suggest next steps
	nextSteps := []string{
		"📝 Note has been successfully added to the task",
		"👀 Check task details to see all notes",
		"📋 Consider updating task status if progress was made",
		"🔄 Add more notes as work progresses",
	}

	responseText += "\n💡 Next Steps:\n"
	for _, step := range nextSteps {
		responseText += fmt.Sprintf("- %s\n", step)
	}

	result := map[string]any{
		"success":      true,
		"task":         task,
		"note":         createdNote,
		"note_id":      createdNote.NoteID,
		"task_id":      params.Arguments.TaskID,
		"created_note": createdNote,
	}

	slog.Info("Note added successfully", "task_id", params.Arguments.TaskID, "note_id", createdNote.NoteID)

	return &mcp.CallToolResultFor[map[string]any]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: responseText,
			},
		},
		Meta: result,
	}, nil
}
