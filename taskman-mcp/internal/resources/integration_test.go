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

// Integration test with full MCP server
func TestResourcesIntegration(t *testing.T) {
	// Create a comprehensive mock API server
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/health":
			json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})

		case r.Method == "GET" && r.URL.Path == "/api/v1/tasks":
			tasks := []Task{
				{
					TaskID:          "int-task-1",
					TaskName:        "Integration Task 1",
					TaskDescription: stringPtr("Description for integration task 1"),
					Status:          "In Progress",
					Priority:        stringPtr("High"),
					AssignedTo:      stringPtr("integration-user"),
					ProjectID:       stringPtr("int-proj-1"),
					DueDate:         stringPtr("2024-02-15T12:00:00Z"),
					CreatedBy:       "admin",
					CreationDate:    "2024-01-01T10:00:00Z",
				},
				{
					TaskID:       "int-task-2",
					TaskName:     "Integration Task 2",
					Status:       "Complete",
					Priority:     stringPtr("Medium"),
					AssignedTo:   stringPtr("integration-user"),
					CreatedBy:    "admin",
					CreationDate: "2024-01-02T10:00:00Z",
				},
			}
			json.NewEncoder(w).Encode(tasks)

		case r.Method == "GET" && r.URL.Path == "/api/v1/tasks/int-task-1":
			task := Task{
				TaskID:          "int-task-1",
				TaskName:        "Integration Task 1",
				TaskDescription: stringPtr("Description for integration task 1"),
				Status:          "In Progress",
				Priority:        stringPtr("High"),
				AssignedTo:      stringPtr("integration-user"),
				ProjectID:       stringPtr("int-proj-1"),
				DueDate:         stringPtr("2024-02-15T12:00:00Z"),
				CreatedBy:       "admin",
				CreationDate:    "2024-01-01T10:00:00Z",
			}
			json.NewEncoder(w).Encode(task)

		case r.Method == "GET" && r.URL.Path == "/api/v1/tasks/int-task-1/notes":
			notes := []TaskNote{
				{
					NoteID:       "int-note-1",
					TaskID:       "int-task-1",
					Note:         "Integration test note",
					CreatedBy:    "integration-user",
					CreationDate: "2024-01-10T15:30:00Z",
				},
			}
			json.NewEncoder(w).Encode(notes)

		case r.Method == "GET" && r.URL.Path == "/api/v1/projects":
			projects := []Project{
				{
					ProjectID:          "int-proj-1",
					ProjectName:        "Integration Project",
					ProjectDescription: stringPtr("Project for integration testing"),
					CreatedBy:          "admin",
					CreationDate:       "2024-01-01T09:00:00Z",
				},
			}
			json.NewEncoder(w).Encode(projects)

		case r.Method == "GET" && r.URL.Path == "/api/v1/projects/int-proj-1":
			project := Project{
				ProjectID:          "int-proj-1",
				ProjectName:        "Integration Project",
				ProjectDescription: stringPtr("Project for integration testing"),
				CreatedBy:          "admin",
				CreationDate:       "2024-01-01T09:00:00Z",
			}
			json.NewEncoder(w).Encode(project)

		case r.Method == "GET" && r.URL.Path == "/api/v1/projects/int-proj-1/tasks":
			tasks := []Task{
				{
					TaskID:       "int-task-1",
					TaskName:     "Integration Task 1",
					Status:       "In Progress",
					Priority:     stringPtr("High"),
					AssignedTo:   stringPtr("integration-user"),
					ProjectID:    stringPtr("int-proj-1"),
					CreatedBy:    "admin",
					CreationDate: "2024-01-01T10:00:00Z",
				},
			}
			json.NewEncoder(w).Encode(tasks)

		default:
			http.NotFound(w, r)
		}
	}))
	defer apiServer.Close()

	// Test resources through direct handler calls
	ctx := context.Background()
	session := &mcp.ServerSession{}

	// Test 1: Task resource
	t.Run("TaskResource", func(t *testing.T) {
		params := &mcp.ReadResourceParams{
			URI: "taskman://task/int-task-1",
		}

		// We can't directly test through the MCP server without setting up transport
		// So we'll test the resource handlers directly but through the server's registered resources
		taskResources := NewTaskResources(client.NewAPIClient(apiServer.URL, 30*time.Second))

		result, err := taskResources.HandleTaskResource(ctx, session, params)
		if err != nil {
			t.Fatalf("TaskResource integration test failed: %v", err)
		}

		if result == nil || len(result.Contents) == 0 {
			t.Fatal("TaskResource returned no content")
		}

		content := result.Contents[0].Text
		if !contains(content, "Integration Task 1") {
			t.Error("TaskResource missing task name")
		}
		if !contains(content, "Integration test note") {
			t.Error("TaskResource missing note content")
		}
		if !contains(content, "Integration Project") {
			t.Error("TaskResource missing project information")
		}
	})

	// Test 2: Tasks overview resource
	t.Run("TasksOverviewResource", func(t *testing.T) {
		params := &mcp.ReadResourceParams{
			URI: "taskman://tasks/overview",
		}

		taskResources := NewTaskResources(client.NewAPIClient(apiServer.URL, 30*time.Second))

		result, err := taskResources.HandleTasksOverviewResource(ctx, session, params)
		if err != nil {
			t.Fatalf("TasksOverviewResource integration test failed: %v", err)
		}

		if result == nil || len(result.Contents) == 0 {
			t.Fatal("TasksOverviewResource returned no content")
		}

		content := result.Contents[0].Text
		if !contains(content, "Tasks Overview") {
			t.Error("TasksOverviewResource missing title")
		}
		if !contains(content, "**Total Tasks:** 2") {
			t.Error("TasksOverviewResource missing task count")
		}
		if !contains(content, "Status Breakdown") {
			t.Error("TasksOverviewResource missing status breakdown")
		}
	})

	// Test 3: Project resource
	t.Run("ProjectResource", func(t *testing.T) {
		params := &mcp.ReadResourceParams{
			URI: "taskman://project/int-proj-1",
		}

		projectResources := NewProjectResources(client.NewAPIClient(apiServer.URL, 30*time.Second))

		result, err := projectResources.HandleProjectResource(ctx, session, params)
		if err != nil {
			t.Fatalf("ProjectResource integration test failed: %v", err)
		}

		if result == nil || len(result.Contents) == 0 {
			t.Fatal("ProjectResource returned no content")
		}

		content := result.Contents[0].Text
		if !contains(content, "Integration Project") {
			t.Error("ProjectResource missing project name")
		}
		if !contains(content, "**Total Tasks:** 1") {
			t.Error("ProjectResource missing task count")
		}
		if !contains(content, "Project for integration testing") {
			t.Error("ProjectResource missing description")
		}
	})

	// Test 4: System dashboard resource
	t.Run("SystemDashboardResource", func(t *testing.T) {
		params := &mcp.ReadResourceParams{
			URI: "taskman://dashboard/system",
		}

		dashboardResources := NewDashboardResources(client.NewAPIClient(apiServer.URL, 30*time.Second))

		result, err := dashboardResources.HandleSystemDashboardResource(ctx, session, params)
		if err != nil {
			t.Fatalf("SystemDashboardResource integration test failed: %v", err)
		}

		if result == nil || len(result.Contents) == 0 {
			t.Fatal("SystemDashboardResource returned no content")
		}

		content := result.Contents[0].Text
		if !contains(content, "System Dashboard") {
			t.Error("SystemDashboardResource missing title")
		}
		if !contains(content, "**Total Projects:** 1") {
			t.Error("SystemDashboardResource missing project count")
		}
		if !contains(content, "**Total Tasks:** 2") {
			t.Error("SystemDashboardResource missing task count")
		}
		if !contains(content, "Completion Rate:") {
			t.Error("SystemDashboardResource missing completion rate")
		}
	})

	// Test 5: User dashboard resource
	t.Run("UserDashboardResource", func(t *testing.T) {
		// Add assigned_to filter support to mock server
		apiServerWithFilter := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == "GET" && r.URL.Path == "/api/v1/tasks":
				assignedTo := r.URL.Query().Get("assigned_to")
				createdBy := r.URL.Query().Get("created_by")

				var tasks []Task
				if assignedTo == "integration-user" {
					tasks = []Task{
						{
							TaskID:       "int-task-1",
							TaskName:     "User Assigned Task",
							Status:       "In Progress",
							Priority:     stringPtr("High"),
							AssignedTo:   stringPtr("integration-user"),
							CreatedBy:    "admin",
							CreationDate: "2024-01-01T10:00:00Z",
						},
					}
				} else if createdBy == "integration-user" {
					tasks = []Task{
						{
							TaskID:       "int-task-3",
							TaskName:     "User Created Task",
							Status:       "Complete",
							Priority:     stringPtr("Medium"),
							CreatedBy:    "integration-user",
							CreationDate: "2024-01-03T10:00:00Z",
						},
					}
				}
				json.NewEncoder(w).Encode(tasks)
			default:
				http.NotFound(w, r)
			}
		}))
		defer apiServerWithFilter.Close()

		params := &mcp.ReadResourceParams{
			URI: "taskman://dashboard/user/integration-user",
		}

		dashboardResources := NewDashboardResources(client.NewAPIClient(apiServerWithFilter.URL, 30*time.Second))

		result, err := dashboardResources.HandleUserDashboardResource(ctx, session, params)
		if err != nil {
			t.Fatalf("UserDashboardResource integration test failed: %v", err)
		}

		if result == nil || len(result.Contents) == 0 {
			t.Fatal("UserDashboardResource returned no content")
		}

		content := result.Contents[0].Text
		if !contains(content, "Dashboard for integration-user") {
			t.Error("UserDashboardResource missing user title")
		}
		if !contains(content, "**Assigned Tasks:** 1") {
			t.Error("UserDashboardResource missing assigned tasks count")
		}
		if !contains(content, "**Created Tasks:** 1") {
			t.Error("UserDashboardResource missing created tasks count")
		}
	})
}

// Test error handling in integration scenarios
func TestResourcesIntegrationErrorHandling(t *testing.T) {
	// Create API server that returns errors
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/api/v1/tasks/nonexistent":
			http.NotFound(w, r)
		case r.Method == "GET" && r.URL.Path == "/api/v1/projects/nonexistent":
			http.NotFound(w, r)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}))
	defer apiServer.Close()

	ctx := context.Background()
	session := &mcp.ServerSession{}

	// Test task resource with nonexistent task
	t.Run("TaskResourceNotFound", func(t *testing.T) {
		params := &mcp.ReadResourceParams{
			URI: "taskman://task/nonexistent",
		}

		taskResources := NewTaskResources(client.NewAPIClient(apiServer.URL, 30*time.Second))

		_, err := taskResources.HandleTaskResource(ctx, session, params)
		if err == nil {
			t.Fatal("Expected error for nonexistent task")
		}

		if !contains(err.Error(), "failed to get task") {
			t.Error("Error message should mention failed to get task")
		}
	})

	// Test project resource with nonexistent project
	t.Run("ProjectResourceNotFound", func(t *testing.T) {
		params := &mcp.ReadResourceParams{
			URI: "taskman://project/nonexistent",
		}

		projectResources := NewProjectResources(client.NewAPIClient(apiServer.URL, 30*time.Second))

		_, err := projectResources.HandleProjectResource(ctx, session, params)
		if err == nil {
			t.Fatal("Expected error for nonexistent project")
		}

		if !contains(err.Error(), "failed to get project") {
			t.Error("Error message should mention failed to get project")
		}
	})

	// Test system dashboard with API errors
	t.Run("SystemDashboardAPIError", func(t *testing.T) {
		params := &mcp.ReadResourceParams{
			URI: "taskman://dashboard/system",
		}

		dashboardResources := NewDashboardResources(client.NewAPIClient(apiServer.URL, 30*time.Second))

		_, err := dashboardResources.HandleSystemDashboardResource(ctx, session, params)
		if err == nil {
			t.Fatal("Expected error for API server error")
		}

		if !contains(err.Error(), "failed to get tasks") {
			t.Error("Error message should mention failed to get tasks")
		}
	})
}
