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

// UserTools handles user-focused MCP tools
type UserTools struct {
	apiClient *client.APIClient
}

// NewUserTools creates a new user tools handler
func NewUserTools(apiClient *client.APIClient) *UserTools {
	return &UserTools{
		apiClient: apiClient,
	}
}

// GetMyWorkParams defines input for get_my_work tool
type GetMyWorkParams struct {
	UserID         string `json:"user_id"`
	IncludeReview  bool   `json:"include_review,omitempty"`
	IncludeBlocked bool   `json:"include_blocked,omitempty"`
	ProjectID      string `json:"project_id,omitempty"`
	SortBy         string `json:"sort_by,omitempty"`
	Limit          int    `json:"limit,omitempty"`
}

// HandleGetMyWork implements the get_my_work tool
func (u *UserTools) HandleGetMyWork(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.CallToolParamsFor[GetMyWorkParams],
) (*mcp.CallToolResultFor[map[string]any], error) {
	slog.Info("Executing get_my_work tool", "params", params.Arguments)

	// Validate required fields
	if params.Arguments.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	// Build queries for different task categories
	var allUserTasks []Task
	var inProgressTasks []Task
	var reviewTasks []Task
	var blockedTasks []Task

	// Get In Progress tasks
	inProgressQuery := fmt.Sprintf("?assigned_to=%s&status=%s",
		url.QueryEscape(params.Arguments.UserID),
		url.QueryEscape("In Progress"))

	if params.Arguments.ProjectID != "" {
		inProgressQuery += fmt.Sprintf("&project_id=%s", url.QueryEscape(params.Arguments.ProjectID))
	}

	inProgressResp, err := u.apiClient.Get(ctx, "/api/v1/tasks"+inProgressQuery)
	if err != nil {
		slog.Error("Failed to get in-progress tasks", "error", err, "user_id", params.Arguments.UserID)
		return nil, fmt.Errorf("failed to get in-progress tasks: %w", err)
	}

	if err := json.Unmarshal(inProgressResp, &inProgressTasks); err != nil {
		slog.Error("Failed to parse in-progress tasks", "error", err)
		return nil, fmt.Errorf("failed to parse in-progress tasks: %w", err)
	}

	allUserTasks = append(allUserTasks, inProgressTasks...)

	// Get Review tasks if requested
	if params.Arguments.IncludeReview {
		reviewQuery := fmt.Sprintf("?assigned_to=%s&status=%s",
			url.QueryEscape(params.Arguments.UserID),
			url.QueryEscape("Review"))

		if params.Arguments.ProjectID != "" {
			reviewQuery += fmt.Sprintf("&project_id=%s", url.QueryEscape(params.Arguments.ProjectID))
		}

		reviewResp, err := u.apiClient.Get(ctx, "/api/v1/tasks"+reviewQuery)
		if err != nil {
			slog.Warn("Failed to get review tasks", "error", err, "user_id", params.Arguments.UserID)
		} else {
			if err := json.Unmarshal(reviewResp, &reviewTasks); err != nil {
				slog.Warn("Failed to parse review tasks", "error", err)
			} else {
				allUserTasks = append(allUserTasks, reviewTasks...)
			}
		}
	}

	// Get Blocked tasks if requested
	if params.Arguments.IncludeBlocked {
		blockedQuery := fmt.Sprintf("?assigned_to=%s&status=%s",
			url.QueryEscape(params.Arguments.UserID),
			url.QueryEscape("Blocked"))

		if params.Arguments.ProjectID != "" {
			blockedQuery += fmt.Sprintf("&project_id=%s", url.QueryEscape(params.Arguments.ProjectID))
		}

		blockedResp, err := u.apiClient.Get(ctx, "/api/v1/tasks"+blockedQuery)
		if err != nil {
			slog.Warn("Failed to get blocked tasks", "error", err, "user_id", params.Arguments.UserID)
		} else {
			if err := json.Unmarshal(blockedResp, &blockedTasks); err != nil {
				slog.Warn("Failed to parse blocked tasks", "error", err)
			} else {
				allUserTasks = append(allUserTasks, blockedTasks...)
			}
		}
	}

	// Analyze task workload
	priorityCounts := make(map[string]int)
	projectCounts := make(map[string]int)
	overdueTasks := []Task{}
	dueSoonTasks := []Task{}

	now := time.Now()
	dueSoonThreshold := now.Add(3 * 24 * time.Hour) // 3 days

	for _, task := range allUserTasks {
		// Count by priority
		if task.Priority != nil {
			priorityCounts[*task.Priority]++
		} else {
			priorityCounts["None"]++
		}

		// Count by project
		if task.ProjectID != nil {
			projectCounts[*task.ProjectID]++
		} else {
			projectCounts["No Project"]++
		}

		// Check due dates
		if isTaskOverdue(task) {
			overdueTasks = append(overdueTasks, task)
		} else if task.DueDate != nil {
			if dueDate, err := time.Parse(time.RFC3339, *task.DueDate); err == nil {
				if dueDate.Before(dueSoonThreshold) && dueDate.After(now) {
					dueSoonTasks = append(dueSoonTasks, task)
				}
			}
		}
	}

	// Apply sorting and limiting
	sortedTasks := allUserTasks

	// Simple priority-based sorting
	if params.Arguments.SortBy == "priority" || params.Arguments.SortBy == "" {
		// Sort by priority: High -> Medium -> Low -> None
		// This is a simplified sort - in real implementation, use sort.Slice
		var highPriority, mediumPriority, lowPriority, noPriority []Task

		for _, task := range sortedTasks {
			if task.Priority == nil {
				noPriority = append(noPriority, task)
			} else {
				switch *task.Priority {
				case "High":
					highPriority = append(highPriority, task)
				case "Medium":
					mediumPriority = append(mediumPriority, task)
				case "Low":
					lowPriority = append(lowPriority, task)
				default:
					noPriority = append(noPriority, task)
				}
			}
		}

		sortedTasks = []Task{}
		sortedTasks = append(sortedTasks, highPriority...)
		sortedTasks = append(sortedTasks, mediumPriority...)
		sortedTasks = append(sortedTasks, lowPriority...)
		sortedTasks = append(sortedTasks, noPriority...)
	}

	// Apply limit
	if params.Arguments.Limit > 0 && len(sortedTasks) > params.Arguments.Limit {
		sortedTasks = sortedTasks[:params.Arguments.Limit]
	}

	// Generate workload insights
	var insights []string

	totalTasks := len(allUserTasks)
	if totalTasks == 0 {
		insights = append(insights, "🎉 No active tasks assigned - you're all caught up!")
	} else if totalTasks == 1 {
		insights = append(insights, "✅ Light workload with one active task")
	} else if totalTasks > 10 {
		insights = append(insights, "🔥 Heavy workload - consider prioritizing or delegating")
	} else if totalTasks > 5 {
		insights = append(insights, "📊 Moderate workload - good task balance")
	}

	if len(overdueTasks) > 0 {
		insights = append(insights, fmt.Sprintf("⚠️ %d tasks are overdue and need immediate attention", len(overdueTasks)))
	}

	if len(dueSoonTasks) > 0 {
		insights = append(insights, fmt.Sprintf("📅 %d tasks due in the next 3 days", len(dueSoonTasks)))
	}

	highPriorityCount := priorityCounts["High"]
	if highPriorityCount > totalTasks/2 && totalTasks > 2 {
		insights = append(insights, "🔥 Most tasks are high priority - focus on completion")
	}

	if len(blockedTasks) > 0 {
		insights = append(insights, fmt.Sprintf("🚫 %d tasks are blocked - work on unblocking", len(blockedTasks)))
	}

	projectCount := len(projectCounts)
	if projectCount > 5 {
		insights = append(insights, "📁 Working across many projects - consider context switching overhead")
	}

	// Generate actionable recommendations
	var recommendations []string

	if len(overdueTasks) > 0 {
		recommendations = append(recommendations, "🚨 Address overdue tasks first to get back on track")
	}

	if highPriorityCount > 0 && len(overdueTasks) == 0 {
		recommendations = append(recommendations, fmt.Sprintf("🎯 Focus on %d high-priority tasks", highPriorityCount))
	}

	if len(dueSoonTasks) > 0 {
		recommendations = append(recommendations, "📅 Plan work for upcoming due dates")
	}

	if len(blockedTasks) > 0 {
		recommendations = append(recommendations, "🔓 Follow up on blocked tasks and work to resolve blockers")
	}

	if totalTasks > 8 {
		recommendations = append(recommendations, "📊 Consider breaking down large tasks or delegating")
	}

	if len(inProgressTasks) > 3 {
		recommendations = append(recommendations, "🔄 Too many concurrent tasks - consider completing some before starting new ones")
	}

	if totalTasks > 0 && len(dueSoonTasks) == 0 && len(overdueTasks) == 0 {
		recommendations = append(recommendations, "✅ Good task timing - maintain current pace")
	}

	// Build comprehensive response
	result := map[string]any{
		"all_tasks":          allUserTasks,
		"prioritized_tasks":  sortedTasks,
		"in_progress_tasks":  inProgressTasks,
		"review_tasks":       reviewTasks,
		"blocked_tasks":      blockedTasks,
		"overdue_tasks":      overdueTasks,
		"due_soon_tasks":     dueSoonTasks,
		"total_tasks":        totalTasks,
		"priority_breakdown": priorityCounts,
		"project_breakdown":  projectCounts,
		"overdue_count":      len(overdueTasks),
		"due_soon_count":     len(dueSoonTasks),
		"insights":           insights,
		"recommendations":    recommendations,
		"user_id":            params.Arguments.UserID,
	}

	// Build detailed response text
	responseText := fmt.Sprintf(`My Work Queue\n=============\n\nUser: %s\nActive Tasks: %d\n`,
		params.Arguments.UserID, totalTasks)

	if params.Arguments.ProjectID != "" {
		responseText += fmt.Sprintf("Project Filter: %s\n", params.Arguments.ProjectID)
	}

	responseText += "\n📊 Task Breakdown:\n"
	responseText += fmt.Sprintf("- In Progress: %d\n", len(inProgressTasks))
	if params.Arguments.IncludeReview {
		responseText += fmt.Sprintf("- Review: %d\n", len(reviewTasks))
	}
	if params.Arguments.IncludeBlocked {
		responseText += fmt.Sprintf("- Blocked: %d\n", len(blockedTasks))
	}

	if len(priorityCounts) > 0 {
		responseText += "\n🎯 Priority Breakdown:\n"
		for priority, count := range priorityCounts {
			responseText += fmt.Sprintf("- %s: %d\n", priority, count)
		}
	}

	if len(overdueTasks) > 0 {
		responseText += fmt.Sprintf("\n⚠️ Overdue Tasks (%d):\n", len(overdueTasks))
		for i, task := range overdueTasks {
			if i < 5 { // Show only first 5
				priority := "None"
				if task.Priority != nil {
					priority = *task.Priority
				}
				responseText += fmt.Sprintf("- %s (%s) - Due: %s\n", task.TaskName, priority, *task.DueDate)
			}
		}
		if len(overdueTasks) > 5 {
			responseText += fmt.Sprintf("... and %d more overdue tasks\n", len(overdueTasks)-5)
		}
	}

	if len(dueSoonTasks) > 0 {
		responseText += fmt.Sprintf("\n📅 Due Soon (%d):\n", len(dueSoonTasks))
		for i, task := range dueSoonTasks {
			if i < 5 { // Show only first 5
				priority := "None"
				if task.Priority != nil {
					priority = *task.Priority
				}
				responseText += fmt.Sprintf("- %s (%s) - Due: %s\n", task.TaskName, priority, *task.DueDate)
			}
		}
		if len(dueSoonTasks) > 5 {
			responseText += fmt.Sprintf("... and %d more tasks due soon\n", len(dueSoonTasks)-5)
		}
	}

	if len(sortedTasks) > 0 {
		responseText += fmt.Sprintf("\n📋 Prioritized Task List (showing %d):\n", len(sortedTasks))
		for i, task := range sortedTasks {
			if i < 8 { // Show only first 8
				priority := "None"
				if task.Priority != nil {
					priority = *task.Priority
				}

				dueInfo := ""
				if task.DueDate != nil {
					if isTaskOverdue(task) {
						dueInfo = " - OVERDUE"
					} else {
						dueInfo = fmt.Sprintf(" - Due: %s", *task.DueDate)
					}
				}

				responseText += fmt.Sprintf("%d. %s (%s, %s)%s\n", i+1, task.TaskName, task.Status, priority, dueInfo)
			}
		}
		if len(sortedTasks) > 8 {
			responseText += fmt.Sprintf("... and %d more tasks\n", len(sortedTasks)-8)
		}
	}

	if len(blockedTasks) > 0 {
		responseText += fmt.Sprintf("\n🚫 Blocked Tasks (%d):\n", len(blockedTasks))
		for _, task := range blockedTasks {
			responseText += fmt.Sprintf("- %s\n", task.TaskName)
		}
	}

	if len(insights) > 0 {
		responseText += "\n💡 Workload Insights:\n"
		for _, insight := range insights {
			responseText += fmt.Sprintf("- %s\n", insight)
		}
	}

	if len(recommendations) > 0 {
		responseText += "\n📋 Recommendations:\n"
		for _, recommendation := range recommendations {
			responseText += fmt.Sprintf("- %s\n", recommendation)
		}
	}

	slog.Info("User work queue generated", "user_id", params.Arguments.UserID, "total_tasks", totalTasks, "overdue_count", len(overdueTasks))

	return &mcp.CallToolResultFor[map[string]any]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: responseText,
			},
		},
		Meta: result,
	}, nil
}
