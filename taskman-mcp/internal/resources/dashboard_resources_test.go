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

// Mock API server for dashboard resources testing
func createDashboardResourcesMockAPIServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/api/v1/tasks":
			assignedTo := r.URL.Query().Get("assigned_to")
			createdBy := r.URL.Query().Get("created_by")

			var tasks []Task
			if assignedTo == "user1" {
				tasks = []Task{
					{
						TaskID:       "task-1",
						TaskName:     "User1 Task 1",
						Status:       "In Progress",
						Priority:     stringPtr("High"),
						AssignedTo:   stringPtr("user1"),
						DueDate:      stringPtr("2024-01-20T12:00:00Z"),
						CreatedBy:    "admin",
						CreationDate: "2024-01-01T10:00:00Z",
					},
					{
						TaskID:       "task-2",
						TaskName:     "User1 Task 2",
						Status:       "Review",
						Priority:     stringPtr("Medium"),
						AssignedTo:   stringPtr("user1"),
						DueDate:      stringPtr("2023-12-15T12:00:00Z"), // Overdue
						CreatedBy:    "admin",
						CreationDate: "2024-01-02T10:00:00Z",
					},
				}
			} else if createdBy == "user1" {
				tasks = []Task{
					{
						TaskID:       "task-3",
						TaskName:     "Created Task 1",
						Status:       "Complete",
						Priority:     stringPtr("Low"),
						AssignedTo:   stringPtr("user2"),
						CreatedBy:    "user1",
						CreationDate: "2024-01-03T10:00:00Z",
					},
				}
			} else {
				// Return all tasks for system dashboard
				tasks = []Task{
					{
						TaskID:       "task-1",
						TaskName:     "System Task 1",
						Status:       "In Progress",
						Priority:     stringPtr("High"),
						AssignedTo:   stringPtr("user1"),
						DueDate:      stringPtr("2024-01-20T12:00:00Z"),
						CreatedBy:    "admin",
						CreationDate: "2024-01-01T10:00:00Z",
					},
					{
						TaskID:       "task-2",
						TaskName:     "System Task 2",
						Status:       "Complete",
						Priority:     stringPtr("Medium"),
						AssignedTo:   stringPtr("user2"),
						CreatedBy:    "admin",
						CreationDate: "2024-01-02T10:00:00Z",
					},
					{
						TaskID:       "task-3",
						TaskName:     "System Task 3",
						Status:       "Not Started",
						Priority:     stringPtr("Low"),
						CreatedBy:    "admin",
						CreationDate: "2024-01-03T10:00:00Z",
					},
					{
						TaskID:       "task-4",
						TaskName:     "System Task 4",
						Status:       "Blocked",
						Priority:     stringPtr("High"),
						AssignedTo:   stringPtr("user1"),
						DueDate:      stringPtr("2023-12-01T12:00:00Z"), // Overdue
						CreatedBy:    "admin",
						CreationDate: "2024-01-04T10:00:00Z",
					},
				}
			}
			json.NewEncoder(w).Encode(tasks)

		case r.Method == "GET" && r.URL.Path == "/api/v1/projects":
			projects := []Project{
				{
					ProjectID:          "proj-1",
					ProjectName:        "Dashboard Project 1",
					ProjectDescription: stringPtr("Description 1"),
					CreatedBy:          "admin",
					CreationDate:       "2024-01-01T09:00:00Z",
				},
				{
					ProjectID:    "proj-2",
					ProjectName:  "Dashboard Project 2",
					CreatedBy:    "user1",
					CreationDate: "2024-01-02T09:00:00Z",
				},
			}
			json.NewEncoder(w).Encode(projects)

		case r.Method == "GET" && r.URL.Path == "/api/v1/projects/proj-1":
			project := Project{
				ProjectID:          "proj-1",
				ProjectName:        "Test Dashboard Project",
				ProjectDescription: stringPtr("Test project for dashboard"),
				CreatedBy:          "admin",
				CreationDate:       "2024-01-01T09:00:00Z",
			}
			json.NewEncoder(w).Encode(project)

		case r.Method == "GET" && r.URL.Path == "/api/v1/projects/proj-1/tasks":
			tasks := []Task{
				{
					TaskID:       "task-1",
					TaskName:     "Project Dashboard Task 1",
					Status:       "In Progress",
					Priority:     stringPtr("High"),
					AssignedTo:   stringPtr("user1"),
					ProjectID:    stringPtr("proj-1"),
					DueDate:      stringPtr("2023-12-01T12:00:00Z"), // Overdue
					CreatedBy:    "admin",
					CreationDate: "2024-01-01T10:00:00Z",
				},
				{
					TaskID:       "task-2",
					TaskName:     "Project Dashboard Task 2",
					Status:       "Complete",
					Priority:     stringPtr("Medium"),
					AssignedTo:   stringPtr("user2"),
					ProjectID:    stringPtr("proj-1"),
					CreatedBy:    "admin",
					CreationDate: "2024-01-02T10:00:00Z",
				},
			}
			json.NewEncoder(w).Encode(tasks)

		default:
			http.NotFound(w, r)
		}
	}))
}

func TestDashboardResources_HandleSystemDashboardResource(t *testing.T) {
	server := createDashboardResourcesMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	dashboardResources := NewDashboardResources(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.ReadResourceParams{
		URI: "taskman://dashboard/system",
	}

	result, err := dashboardResources.HandleSystemDashboardResource(ctx, session, params)
	if err != nil {
		t.Fatalf("HandleSystemDashboardResource failed: %v", err)
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

	// Check that content contains expected system dashboard information
	text := content.Text
	if !contains(text, "System Dashboard") {
		t.Error("Content missing system dashboard title")
	}
	if !contains(text, "**Total Projects:** 2") {
		t.Error("Content missing total project count")
	}
	if !contains(text, "**Total Tasks:** 4") {
		t.Error("Content missing total task count")
	}
	if !contains(text, "Completion Rate:") {
		t.Error("Content missing completion rate")
	}
	if !contains(text, "Overdue Tasks:") {
		t.Error("Content missing overdue tasks")
	}
	if !contains(text, "Task Status Distribution") {
		t.Error("Content missing status distribution")
	}
	if !contains(text, "Priority Distribution") {
		t.Error("Content missing priority distribution")
	}
	if !contains(text, "Top Assignees") {
		t.Error("Content missing top assignees")
	}
}

func TestDashboardResources_HandleUserDashboardResource(t *testing.T) {
	server := createDashboardResourcesMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	dashboardResources := NewDashboardResources(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.ReadResourceParams{
		URI: "taskman://dashboard/user/user1",
	}

	result, err := dashboardResources.HandleUserDashboardResource(ctx, session, params)
	if err != nil {
		t.Fatalf("HandleUserDashboardResource failed: %v", err)
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

	// Check that content contains expected user dashboard information
	text := content.Text
	if !contains(text, "Dashboard for user1") {
		t.Error("Content missing user dashboard title")
	}
	if !contains(text, "**Assigned Tasks:** 2") {
		t.Error("Content missing assigned tasks count")
	}
	if !contains(text, "**Created Tasks:** 1") {
		t.Error("Content missing created tasks count")
	}
	if !contains(text, "My Completion Rate:") {
		t.Error("Content missing completion rate")
	}
	if !contains(text, "My Overdue Tasks:") {
		t.Error("Content missing overdue tasks")
	}
	if !contains(text, "My Task Status") {
		t.Error("Content missing task status")
	}
	if !contains(text, "Current Workload") {
		t.Error("Content missing current workload")
	}
}

func TestDashboardResources_HandleProjectDashboardResource(t *testing.T) {
	server := createDashboardResourcesMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	dashboardResources := NewDashboardResources(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.ReadResourceParams{
		URI: "taskman://dashboard/project/proj-1",
	}

	result, err := dashboardResources.HandleProjectDashboardResource(ctx, session, params)
	if err != nil {
		t.Fatalf("HandleProjectDashboardResource failed: %v", err)
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

	// Check that content contains expected project dashboard information
	text := content.Text
	if !contains(text, "Project Dashboard: Test Dashboard Project") {
		t.Error("Content missing project dashboard title")
	}
	if !contains(text, "proj-1") {
		t.Error("Content missing project ID")
	}
	if !contains(text, "**Total Tasks:** 2") {
		t.Error("Content missing total task count")
	}
	if !contains(text, "Completion Rate:") {
		t.Error("Content missing completion rate")
	}
	if !contains(text, "Overdue Tasks:") {
		t.Error("Content missing overdue tasks")
	}
	if !contains(text, "Task Status Distribution") {
		t.Error("Content missing status distribution")
	}
	if !contains(text, "Team Workload") {
		t.Error("Content missing team workload")
	}
	if !contains(text, "Critical Tasks") {
		t.Error("Content missing critical tasks")
	}
}

func TestDashboardResources_HandleUserDashboardResource_InvalidURI(t *testing.T) {
	server := createDashboardResourcesMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	dashboardResources := NewDashboardResources(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.ReadResourceParams{
		URI: "taskman://dashboard/user",
	}

	_, err := dashboardResources.HandleUserDashboardResource(ctx, session, params)
	if err == nil {
		t.Fatal("Expected error for invalid user dashboard URI")
	}
}

func TestDashboardResources_HandleUserDashboardResource_EmptyUserID(t *testing.T) {
	server := createDashboardResourcesMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	dashboardResources := NewDashboardResources(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.ReadResourceParams{
		URI: "taskman://dashboard/user/",
	}

	_, err := dashboardResources.HandleUserDashboardResource(ctx, session, params)
	if err == nil {
		t.Fatal("Expected error for empty user ID")
	}
}

func TestDashboardResources_HandleProjectDashboardResource_InvalidURI(t *testing.T) {
	server := createDashboardResourcesMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	dashboardResources := NewDashboardResources(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.ReadResourceParams{
		URI: "taskman://dashboard/project",
	}

	_, err := dashboardResources.HandleProjectDashboardResource(ctx, session, params)
	if err == nil {
		t.Fatal("Expected error for invalid project dashboard URI")
	}
}

func TestDashboardResources_HandleProjectDashboardResource_EmptyProjectID(t *testing.T) {
	server := createDashboardResourcesMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	dashboardResources := NewDashboardResources(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.ReadResourceParams{
		URI: "taskman://dashboard/project/",
	}

	_, err := dashboardResources.HandleProjectDashboardResource(ctx, session, params)
	if err == nil {
		t.Fatal("Expected error for empty project ID")
	}
}
