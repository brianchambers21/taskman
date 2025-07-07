package resources

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bchamber/taskman-mcp/internal/client"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Mock API server for task resources testing
func createTaskResourcesMockAPIServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/api/v1/tasks":
			// Return tasks based on query parameters
			assignedTo := r.URL.Query().Get("assigned_to")

			var tasks []Task
			if assignedTo == "user1" {
				tasks = []Task{
					{
						TaskID:          "task-1",
						TaskName:        "User1 Task 1",
						TaskDescription: stringPtr("Description for task 1"),
						Status:          "In Progress",
						Priority:        stringPtr("High"),
						AssignedTo:      stringPtr("user1"),
						ProjectID:       stringPtr("proj-1"),
						DueDate:         stringPtr("2024-01-15T12:00:00Z"),
						CreatedBy:       "admin",
						CreationDate:    "2024-01-01T10:00:00Z",
					},
					{
						TaskID:       "task-2",
						TaskName:     "User1 Task 2",
						Status:       "Complete",
						Priority:     stringPtr("Medium"),
						AssignedTo:   stringPtr("user1"),
						CreatedBy:    "admin",
						CreationDate: "2024-01-02T10:00:00Z",
					},
				}
			} else {
				// Return all tasks
				tasks = []Task{
					{
						TaskID:          "task-1",
						TaskName:        "Task 1",
						TaskDescription: stringPtr("Description for task 1"),
						Status:          "In Progress",
						Priority:        stringPtr("High"),
						AssignedTo:      stringPtr("user1"),
						ProjectID:       stringPtr("proj-1"),
						DueDate:         stringPtr("2024-01-15T12:00:00Z"),
						CreatedBy:       "admin",
						CreationDate:    "2024-01-01T10:00:00Z",
					},
					{
						TaskID:       "task-2",
						TaskName:     "Task 2",
						Status:       "Complete",
						Priority:     stringPtr("Medium"),
						AssignedTo:   stringPtr("user1"),
						CreatedBy:    "admin",
						CreationDate: "2024-01-02T10:00:00Z",
					},
					{
						TaskID:       "task-3",
						TaskName:     "Task 3",
						Status:       "Not Started",
						Priority:     stringPtr("Low"),
						CreatedBy:    "admin",
						CreationDate: "2024-01-03T10:00:00Z",
					},
				}
			}
			json.NewEncoder(w).Encode(tasks)

		case r.Method == "GET" && r.URL.Path == "/api/v1/tasks/task-1":
			task := Task{
				TaskID:          "task-1",
				TaskName:        "Test Task 1",
				TaskDescription: stringPtr("Test description"),
				Status:          "In Progress",
				Priority:        stringPtr("High"),
				AssignedTo:      stringPtr("user1"),
				ProjectID:       stringPtr("proj-1"),
				DueDate:         stringPtr("2024-01-15T12:00:00Z"),
				CreatedBy:       "admin",
				CreationDate:    "2024-01-01T10:00:00Z",
				LastUpdatedBy:   stringPtr("user1"),
				LastUpdateDate:  stringPtr("2024-01-10T15:30:00Z"),
			}
			json.NewEncoder(w).Encode(task)

		case r.Method == "GET" && r.URL.Path == "/api/v1/tasks/task-1/notes":
			notes := []TaskNote{
				{
					NoteID:       "note-1",
					TaskID:       "task-1",
					Note:         "Starting work on this task",
					CreatedBy:    "user1",
					CreationDate: "2024-01-10T15:30:00Z",
				},
				{
					NoteID:       "note-2",
					TaskID:       "task-1",
					Note:         "Made good progress today",
					CreatedBy:    "user1",
					CreationDate: "2024-01-11T16:00:00Z",
				},
			}
			json.NewEncoder(w).Encode(notes)

		case r.Method == "GET" && r.URL.Path == "/api/v1/projects/proj-1":
			project := Project{
				ProjectID:          "proj-1",
				ProjectName:        "Test Project",
				ProjectDescription: stringPtr("Test project description"),
				CreatedBy:          "admin",
				CreationDate:       "2024-01-01T09:00:00Z",
			}
			json.NewEncoder(w).Encode(project)

		default:
			http.NotFound(w, r)
		}
	}))
}

func stringPtr(s string) *string {
	return &s
}

func TestTaskResources_HandleTaskResource(t *testing.T) {
	server := createTaskResourcesMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	taskResources := NewTaskResources(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.ReadResourceParams{
		URI: "taskman://task/task-1",
	}

	result, err := taskResources.HandleTaskResource(ctx, session, params)
	if err != nil {
		t.Fatalf("HandleTaskResource failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if len(result.Contents) == 0 {
		t.Fatal("No contents in result")
	}

	content := result.Contents[0]
	if content.URI != params.URI {
		t.Errorf("Expected URI %s, got %s", params.URI, content.URI)
	}

	if content.MIMEType != "text/plain" {
		t.Errorf("Expected MIMEType text/plain, got %s", content.MIMEType)
	}

	if content.Text == "" {
		t.Fatal("Text content is empty")
	}

	// Check that content contains expected task information
	text := content.Text
	if !contains(text, "Test Task 1") {
		t.Error("Content missing task name")
	}
	if !contains(text, "task-1") {
		t.Error("Content missing task ID")
	}
	if !contains(text, "In Progress") {
		t.Error("Content missing status")
	}
	if !contains(text, "High") {
		t.Error("Content missing priority")
	}
	if !contains(text, "user1") {
		t.Error("Content missing assignee")
	}
	if !contains(text, "Test Project") {
		t.Error("Content missing project name")
	}
	if !contains(text, "Starting work on this task") {
		t.Error("Content missing note")
	}
}

func TestTaskResources_HandleTasksOverviewResource(t *testing.T) {
	server := createTaskResourcesMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	taskResources := NewTaskResources(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.ReadResourceParams{
		URI: "taskman://tasks/overview",
	}

	result, err := taskResources.HandleTasksOverviewResource(ctx, session, params)
	if err != nil {
		t.Fatalf("HandleTasksOverviewResource failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if len(result.Contents) == 0 {
		t.Fatal("No contents in result")
	}

	content := result.Contents[0]
	if content.URI != params.URI {
		t.Errorf("Expected URI %s, got %s", params.URI, content.URI)
	}

	if content.MIMEType != "text/plain" {
		t.Errorf("Expected MIMEType text/plain, got %s", content.MIMEType)
	}

	if content.Text == "" {
		t.Fatal("Text content is empty")
	}

	// Check that content contains expected overview information
	text := content.Text
	if !contains(text, "Tasks Overview") {
		t.Error("Content missing overview title")
	}
	if !contains(text, "**Total Tasks:** 3") {
		t.Error("Content missing total task count")
	}
	if !contains(text, "Status Breakdown") {
		t.Error("Content missing status breakdown")
	}
	if !contains(text, "Priority Breakdown") {
		t.Error("Content missing priority breakdown")
	}
	if !contains(text, "In Progress: 1") {
		t.Error("Content missing status count")
	}
}

func TestTaskResources_HandleUserTasksResource(t *testing.T) {
	server := createTaskResourcesMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	taskResources := NewTaskResources(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.ReadResourceParams{
		URI: "taskman://tasks/user/user1",
	}

	result, err := taskResources.HandleUserTasksResource(ctx, session, params)
	if err != nil {
		t.Fatalf("HandleUserTasksResource failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if len(result.Contents) == 0 {
		t.Fatal("No contents in result")
	}

	content := result.Contents[0]
	if content.URI != params.URI {
		t.Errorf("Expected URI %s, got %s", params.URI, content.URI)
	}

	if content.Text == "" {
		t.Fatal("Text content is empty")
	}

	// Check that content contains expected user tasks information
	text := content.Text
	if !contains(text, "Tasks for user1") {
		t.Error("Content missing user title")
	}
	if !contains(text, "**Total Tasks:** 2") {
		t.Error("Content missing total task count")
	}
	if !contains(text, "In Progress") {
		t.Error("Content missing status section")
	}
	if !contains(text, "Complete") {
		t.Error("Content missing completed section")
	}
}

func TestTaskResources_HandleTaskResource_InvalidURI(t *testing.T) {
	server := createTaskResourcesMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	taskResources := NewTaskResources(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.ReadResourceParams{
		URI: "taskman://invalid",
	}

	_, err := taskResources.HandleTaskResource(ctx, session, params)
	if err == nil {
		t.Fatal("Expected error for invalid URI")
	}
}

func TestTaskResources_HandleTaskResource_EmptyTaskID(t *testing.T) {
	server := createTaskResourcesMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	taskResources := NewTaskResources(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.ReadResourceParams{
		URI: "taskman://task/",
	}

	_, err := taskResources.HandleTaskResource(ctx, session, params)
	if err == nil {
		t.Fatal("Expected error for empty task ID")
	}
}

func TestTaskResources_HandleUserTasksResource_InvalidURI(t *testing.T) {
	server := createTaskResourcesMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	taskResources := NewTaskResources(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.ReadResourceParams{
		URI: "taskman://tasks/user",
	}

	_, err := taskResources.HandleUserTasksResource(ctx, session, params)
	if err == nil {
		t.Fatal("Expected error for invalid user tasks URI")
	}
}

func TestTaskResources_HandleUserTasksResource_EmptyUserID(t *testing.T) {
	server := createTaskResourcesMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	taskResources := NewTaskResources(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.ReadResourceParams{
		URI: "taskman://tasks/user/",
	}

	_, err := taskResources.HandleUserTasksResource(ctx, session, params)
	if err == nil {
		t.Fatal("Expected error for empty user ID")
	}
}

// Helper function to check if string contains substring
func contains(text, substr string) bool {
	return len(text) >= len(substr) && indexOf(text, substr) >= 0
}

// Helper function to find index of substring
func indexOf(text, substr string) int {
	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
