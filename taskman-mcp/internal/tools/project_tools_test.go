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

// Mock API server for project tools testing
func createProjectMockAPIServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/api/v1/projects/proj-1":
			project := Project{
				ProjectID:          "proj-1",
				ProjectName:        "Test Project",
				ProjectDescription: stringPtr("Test project description"),
				CreatedBy:          "admin",
				CreationDate:       "2024-01-01T10:00:00Z",
			}
			json.NewEncoder(w).Encode(project)

		case r.Method == "GET" && r.URL.Path == "/api/v1/projects/proj-1/tasks":
			tasks := []Task{
				{
					TaskID:       "task-1",
					TaskName:     "Task 1",
					Status:       "In Progress",
					Priority:     stringPtr("High"),
					AssignedTo:   stringPtr("user1"),
					ProjectID:    stringPtr("proj-1"),
					DueDate:      stringPtr("2024-01-15T12:00:00Z"),
					CreatedBy:    "admin",
					CreationDate: "2024-01-01T10:00:00Z",
				},
				{
					TaskID:       "task-2",
					TaskName:     "Task 2",
					Status:       "Complete",
					Priority:     stringPtr("Medium"),
					ProjectID:    stringPtr("proj-1"),
					CreatedBy:    "admin",
					CreationDate: "2024-01-02T10:00:00Z",
				},
				{
					TaskID:         "task-3",
					TaskName:       "Task 3",
					Status:         "Not Started",
					Priority:       stringPtr("Low"),
					ProjectID:      stringPtr("proj-1"),
					DueDate:        stringPtr("2023-12-01T12:00:00Z"), // Overdue
					CreatedBy:      "admin",
					CreationDate:   "2024-01-03T10:00:00Z",
				},
			}
			json.NewEncoder(w).Encode(tasks)

		case r.Method == "POST" && r.URL.Path == "/api/v1/projects":
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)

			project := Project{
				ProjectID:   "proj-new",
				ProjectName: req["project_name"].(string),
				CreatedBy:   req["created_by"].(string),
				CreationDate: time.Now().Format(time.RFC3339),
			}

			if desc, ok := req["project_description"]; ok {
				project.ProjectDescription = stringPtr(desc.(string))
			}

			json.NewEncoder(w).Encode(project)

		case r.Method == "POST" && r.URL.Path == "/api/v1/tasks":
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)

			taskID := "task-new-" + time.Now().Format("20060102150405")
			task := Task{
				TaskID:       taskID,
				TaskName:     req["task_name"].(string),
				Status:       "Not Started",
				CreatedBy:    req["created_by"].(string),
				CreationDate: time.Now().Format(time.RFC3339),
			}

			if desc, ok := req["task_description"]; ok {
				task.TaskDescription = stringPtr(desc.(string))
			}
			if status, ok := req["status"]; ok {
				task.Status = status.(string)
			}
			if priority, ok := req["priority"]; ok {
				task.Priority = stringPtr(priority.(string))
			}
			if assignedTo, ok := req["assigned_to"]; ok {
				task.AssignedTo = stringPtr(assignedTo.(string))
			}
			if projectID, ok := req["project_id"]; ok {
				task.ProjectID = stringPtr(projectID.(string))
			}

			json.NewEncoder(w).Encode(task)

		default:
			http.NotFound(w, r)
		}
	}))
}

func TestProjectTools_HandleGetProjectStatus(t *testing.T) {
	server := createProjectMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	projectTools := NewProjectTools(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[GetProjectStatusParams]{
		Arguments: GetProjectStatusParams{
			ProjectID: "proj-1",
		},
	}

	result, err := projectTools.HandleGetProjectStatus(ctx, session, params)
	if err != nil {
		t.Fatalf("HandleGetProjectStatus failed: %v", err)
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
	if _, ok := meta["project"]; !ok {
		t.Error("Meta missing project")
	}
	if _, ok := meta["total_tasks"]; !ok {
		t.Error("Meta missing total_tasks")
	}
	if _, ok := meta["completion_percentage"]; !ok {
		t.Error("Meta missing completion_percentage")
	}
	if _, ok := meta["status_breakdown"]; !ok {
		t.Error("Meta missing status_breakdown")
	}
	if _, ok := meta["overdue_count"]; !ok {
		t.Error("Meta missing overdue_count")
	}
}

func TestProjectTools_HandleCreateProjectWithInitialTasks(t *testing.T) {
	server := createProjectMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	projectTools := NewProjectTools(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[CreateProjectWithInitialTasksParams]{
		Arguments: CreateProjectWithInitialTasksParams{
			ProjectName:        "New Test Project",
			ProjectDescription: "Description for new project",
			CreatedBy:          "test.user",
			InitialTasks: []InitialTaskSpec{
				{
					TaskName:        "Initial Task 1",
					TaskDescription: "Description for task 1",
					Status:          "Not Started",
					Priority:        "High",
					AssignedTo:      "user1",
				},
				{
					TaskName:        "Initial Task 2",
					TaskDescription: "Description for task 2",
					Status:          "Not Started",
					Priority:        "Medium",
					AssignedTo:      "user2",
				},
			},
		},
	}

	result, err := projectTools.HandleCreateProjectWithInitialTasks(ctx, session, params)
	if err != nil {
		t.Fatalf("HandleCreateProjectWithInitialTasks failed: %v", err)
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
	if _, ok := meta["project"]; !ok {
		t.Error("Meta missing project")
	}
	if _, ok := meta["created_tasks"]; !ok {
		t.Error("Meta missing created_tasks")
	}
	if _, ok := meta["total_planned"]; !ok {
		t.Error("Meta missing total_planned")
	}
	if _, ok := meta["total_created"]; !ok {
		t.Error("Meta missing total_created")
	}
}

func TestProjectTools_HandleGetProjectStatus_MissingProjectID(t *testing.T) {
	server := createProjectMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	projectTools := NewProjectTools(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[GetProjectStatusParams]{
		Arguments: GetProjectStatusParams{
			ProjectID: "",
		},
	}

	_, err := projectTools.HandleGetProjectStatus(ctx, session, params)
	if err == nil {
		t.Fatal("Expected error for missing project_id")
	}
}

func TestProjectTools_HandleCreateProjectWithInitialTasks_MissingRequiredFields(t *testing.T) {
	server := createProjectMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	projectTools := NewProjectTools(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}

	// Test missing project_name
	params := &mcp.CallToolParamsFor[CreateProjectWithInitialTasksParams]{
		Arguments: CreateProjectWithInitialTasksParams{
			ProjectName: "",
			CreatedBy:   "test.user",
			InitialTasks: []InitialTaskSpec{
				{TaskName: "Task 1"},
			},
		},
	}

	_, err := projectTools.HandleCreateProjectWithInitialTasks(ctx, session, params)
	if err == nil {
		t.Fatal("Expected error for missing project_name")
	}

	// Test missing created_by
	params.Arguments.ProjectName = "Test Project"
	params.Arguments.CreatedBy = ""

	_, err = projectTools.HandleCreateProjectWithInitialTasks(ctx, session, params)
	if err == nil {
		t.Fatal("Expected error for missing created_by")
	}

	// Test missing initial_tasks
	params.Arguments.CreatedBy = "test.user"
	params.Arguments.InitialTasks = []InitialTaskSpec{}

	_, err = projectTools.HandleCreateProjectWithInitialTasks(ctx, session, params)
	if err == nil {
		t.Fatal("Expected error for missing initial_tasks")
	}
}