package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/bchamber/taskman-mcp/internal/client"
)

// Mock API server for user tools testing
func createUserMockAPIServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/api/v1/tasks":
			status := r.URL.Query().Get("status")
			assignedTo := r.URL.Query().Get("assigned_to")

			var tasks []Task

			// Return different tasks based on status and assigned user
			if assignedTo == "user1" {
				switch status {
				case "In Progress":
					tasks = []Task{
						{
							TaskID:       "task-1",
							TaskName:     "User1 In Progress Task 1",
							Status:       "In Progress",
							Priority:     stringPtr("High"),
							AssignedTo:   stringPtr("user1"),
							DueDate:      stringPtr("2024-01-15T12:00:00Z"),
							CreatedBy:    "admin",
							CreationDate: "2024-01-01T10:00:00Z",
						},
						{
							TaskID:       "task-2",
							TaskName:     "User1 In Progress Task 2",
							Status:       "In Progress",
							Priority:     stringPtr("Medium"),
							AssignedTo:   stringPtr("user1"),
							DueDate:      stringPtr("2023-12-01T12:00:00Z"), // Overdue
							CreatedBy:    "admin",
							CreationDate: "2024-01-02T10:00:00Z",
						},
					}
				case "Review":
					tasks = []Task{
						{
							TaskID:       "task-3",
							TaskName:     "User1 Review Task",
							Status:       "Review",
							Priority:     stringPtr("High"),
							AssignedTo:   stringPtr("user1"),
							CreatedBy:    "admin",
							CreationDate: "2024-01-03T10:00:00Z",
						},
					}
				case "Blocked":
					tasks = []Task{
						{
							TaskID:       "task-4",
							TaskName:     "User1 Blocked Task",
							Status:       "Blocked",
							Priority:     stringPtr("Medium"),
							AssignedTo:   stringPtr("user1"),
							CreatedBy:    "admin",
							CreationDate: "2024-01-04T10:00:00Z",
						},
					}
				}
			}

			json.NewEncoder(w).Encode(tasks)

		default:
			http.NotFound(w, r)
		}
	}))
}

func TestUserTools_HandleGetMyWork(t *testing.T) {
	server := createUserMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	userTools := NewUserTools(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[GetMyWorkParams]{
		Arguments: GetMyWorkParams{
			UserID:         "user1",
			IncludeReview:  true,
			IncludeBlocked: true,
			SortBy:         "priority",
			Limit:          10,
		},
	}

	result, err := userTools.HandleGetMyWork(ctx, session, params)
	if err != nil {
		t.Fatalf("HandleGetMyWork failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if len(result.Content) == 0 {
		t.Fatal("No content in result")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("First content item is not TextContent")
	}

	if textContent.Text == "" {
		t.Fatal("Text content is empty")
	}

	// Check that meta contains expected fields
	if result.Meta == nil {
		t.Fatal("Meta is nil")
	}

	meta := result.Meta
	if _, ok := meta["all_tasks"]; !ok {
		t.Error("Meta missing all_tasks")
	}
	if _, ok := meta["prioritized_tasks"]; !ok {
		t.Error("Meta missing prioritized_tasks")
	}
	if _, ok := meta["in_progress_tasks"]; !ok {
		t.Error("Meta missing in_progress_tasks")
	}
	if _, ok := meta["total_tasks"]; !ok {
		t.Error("Meta missing total_tasks")
	}
	if _, ok := meta["priority_breakdown"]; !ok {
		t.Error("Meta missing priority_breakdown")
	}
	if _, ok := meta["overdue_count"]; !ok {
		t.Error("Meta missing overdue_count")
	}
	if _, ok := meta["user_id"]; !ok {
		t.Error("Meta missing user_id")
	}

	// Verify user_id is correct
	if userID := meta["user_id"].(string); userID != "user1" {
		t.Errorf("Expected user_id 'user1', got '%s'", userID)
	}
}

func TestUserTools_HandleGetMyWork_MinimalParams(t *testing.T) {
	server := createUserMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	userTools := NewUserTools(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[GetMyWorkParams]{
		Arguments: GetMyWorkParams{
			UserID: "user1",
		},
	}

	result, err := userTools.HandleGetMyWork(ctx, session, params)
	if err != nil {
		t.Fatalf("HandleGetMyWork with minimal params failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	// Check that result contains at least in-progress tasks
	meta := result.Meta
	inProgressTasks := meta["in_progress_tasks"].([]Task)
	if len(inProgressTasks) == 0 {
		t.Error("Expected at least some in-progress tasks")
	}

	// Review tasks should be empty since IncludeReview is false
	reviewTasks := meta["review_tasks"].([]Task)
	if len(reviewTasks) != 0 {
		t.Error("Expected no review tasks when IncludeReview is false")
	}

	// Blocked tasks should be empty since IncludeBlocked is false
	blockedTasks := meta["blocked_tasks"].([]Task)
	if len(blockedTasks) != 0 {
		t.Error("Expected no blocked tasks when IncludeBlocked is false")
	}
}

func TestUserTools_HandleGetMyWork_WithProjectFilter(t *testing.T) {
	server := createUserMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	userTools := NewUserTools(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[GetMyWorkParams]{
		Arguments: GetMyWorkParams{
			UserID:    "user1",
			ProjectID: "proj-1",
		},
	}

	result, err := userTools.HandleGetMyWork(ctx, session, params)
	if err != nil {
		t.Fatalf("HandleGetMyWork with project filter failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	// Check that response text mentions project filter
	textContent := result.Content[0].(*mcp.TextContent)
	if textContent.Text == "" {
		t.Fatal("Text content is empty")
	}

	// The response should mention the project filter
	// This is a simple check - in a real test you might parse the content more carefully
}

func TestUserTools_HandleGetMyWork_MissingUserID(t *testing.T) {
	server := createUserMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	userTools := NewUserTools(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[GetMyWorkParams]{
		Arguments: GetMyWorkParams{
			UserID: "",
		},
	}

	_, err := userTools.HandleGetMyWork(ctx, session, params)
	if err == nil {
		t.Fatal("Expected error for missing user_id")
	}
}

func TestUserTools_HandleGetMyWork_WithLimit(t *testing.T) {
	server := createUserMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	userTools := NewUserTools(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[GetMyWorkParams]{
		Arguments: GetMyWorkParams{
			UserID:         "user1",
			IncludeReview:  true,
			IncludeBlocked: true,
			Limit:          1, // Limit to 1 task
		},
	}

	result, err := userTools.HandleGetMyWork(ctx, session, params)
	if err != nil {
		t.Fatalf("HandleGetMyWork with limit failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	meta := result.Meta
	prioritizedTasks := meta["prioritized_tasks"].([]Task)
	
	// Should respect the limit
	if len(prioritizedTasks) > 1 {
		t.Errorf("Expected at most 1 task due to limit, got %d", len(prioritizedTasks))
	}
}