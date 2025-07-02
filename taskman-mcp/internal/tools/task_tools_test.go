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

// Mock API server for testing
func createMockAPIServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/api/v1/tasks":
			// Filter tasks based on query parameters
			status := r.URL.Query().Get("status")
			
			// Handle space in status parameter for URL encoding
			if status == "In Progress" || status == "In+Progress" {
				status = "In Progress"
			}
			
			tasks := []Task{
				{
					TaskID:         "task-1",
					TaskName:       "Test Task 1",
					Status:         "In Progress",
					Priority:       stringPtr("High"),
					AssignedTo:     stringPtr("john.doe"),
					DueDate:        stringPtr("2024-01-15T12:00:00Z"),
					CreatedBy:      "admin",
					CreationDate:   "2024-01-01T10:00:00Z",
				},
				{
					TaskID:       "task-2",
					TaskName:     "Test Task 2",
					Status:       "Complete",
					CreatedBy:    "admin",
					CreationDate: "2024-01-02T10:00:00Z",
				},
			}
			
			// Apply status filter if provided
			if status != "" {
				var filteredTasks []Task
				for _, task := range tasks {
					if task.Status == status {
						filteredTasks = append(filteredTasks, task)
					}
				}
				tasks = filteredTasks
			}
			
			json.NewEncoder(w).Encode(tasks)

		case r.Method == "GET" && r.URL.Path == "/api/v1/projects":
			projects := []Project{
				{
					ProjectID:   "proj-1",
					ProjectName: "Test Project",
					CreatedBy:   "admin",
					CreationDate: "2024-01-01T09:00:00Z",
				},
			}
			json.NewEncoder(w).Encode(projects)

		case r.Method == "POST" && r.URL.Path == "/api/v1/tasks":
			task := Task{
				TaskID:       "task-new",
				TaskName:     "New Task",
				Status:       "Not Started",
				CreatedBy:    "test.user",
				CreationDate: time.Now().Format(time.RFC3339),
			}
			json.NewEncoder(w).Encode(task)

		case r.Method == "POST" && r.URL.Path == "/api/v1/tasks/task-new/notes":
			note := TaskNote{
				NoteID:       "note-1",
				TaskID:       "task-new",
				Note:         "Initial planning note",
				CreatedBy:    "test.user",
				CreationDate: time.Now().Format(time.RFC3339),
			}
			json.NewEncoder(w).Encode(note)

		case r.Method == "GET" && r.URL.Path == "/api/v1/tasks/task-1":
			task := Task{
				TaskID:         "task-1",
				TaskName:       "Test Task 1",
				TaskDescription: stringPtr("Test description"),
				Status:         "In Progress",
				Priority:       stringPtr("High"),
				AssignedTo:     stringPtr("john.doe"),
				ProjectID:      stringPtr("proj-1"),
				DueDate:        stringPtr("2024-01-15T12:00:00Z"),
				CreatedBy:      "admin",
				CreationDate:   "2024-01-01T10:00:00Z",
				LastUpdatedBy:  stringPtr("john.doe"),
				LastUpdateDate: stringPtr("2024-01-10T15:30:00Z"),
			}
			json.NewEncoder(w).Encode(task)

		case r.Method == "GET" && r.URL.Path == "/api/v1/tasks/task-1/notes":
			notes := []TaskNote{
				{
					NoteID:       "note-1",
					TaskID:       "task-1",
					Note:         "Starting work on this task",
					CreatedBy:    "john.doe",
					CreationDate: "2024-01-10T15:30:00Z",
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

		case r.Method == "PUT" && r.URL.Path == "/api/v1/tasks/task-1":
			task := Task{
				TaskID:         "task-1",
				TaskName:       "Test Task 1",
				Status:         "Complete",
				Priority:       stringPtr("High"),
				AssignedTo:     stringPtr("john.doe"),
				ProjectID:      stringPtr("proj-1"),
				DueDate:        stringPtr("2024-01-15T12:00:00Z"),
				CompletionDate: stringPtr(time.Now().Format(time.RFC3339)),
				CreatedBy:      "admin",
				CreationDate:   "2024-01-01T10:00:00Z",
				LastUpdatedBy:  stringPtr("test.user"),
				LastUpdateDate: stringPtr(time.Now().Format(time.RFC3339)),
			}
			json.NewEncoder(w).Encode(task)

		case r.Method == "POST" && r.URL.Path == "/api/v1/tasks/task-1/notes":
			note := TaskNote{
				NoteID:       "note-2",
				TaskID:       "task-1",
				Note:         "Task completed successfully",
				CreatedBy:    "test.user",
				CreationDate: time.Now().Format(time.RFC3339),
			}
			json.NewEncoder(w).Encode(note)

		default:
			http.NotFound(w, r)
		}
	}))
}

func stringPtr(s string) *string {
	return &s
}

func TestTaskTools_HandleGetTaskOverview(t *testing.T) {
	server := createMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	taskTools := NewTaskTools(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[GetTaskOverviewParams]{
		Arguments: GetTaskOverviewParams{
			Status: "In Progress",
		},
	}

	result, err := taskTools.HandleGetTaskOverview(ctx, session, params)
	if err != nil {
		t.Fatalf("HandleGetTaskOverview failed: %v", err)
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
	if _, ok := meta["total_tasks"]; !ok {
		t.Error("Meta missing total_tasks")
	}
	if _, ok := meta["status_breakdown"]; !ok {
		t.Error("Meta missing status_breakdown")
	}
	if _, ok := meta["overdue_count"]; !ok {
		t.Error("Meta missing overdue_count")
	}
}

func TestTaskTools_HandleCreateTaskWithContext(t *testing.T) {
	server := createMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	taskTools := NewTaskTools(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[CreateTaskWithContextParams]{
		Arguments: CreateTaskWithContextParams{
			TaskName:        "New Test Task",
			TaskDescription: "Description for new task",
			Status:          "Not Started",
			Priority:        "Medium",
			InitialNote:     "Initial planning note",
			CreatedBy:       "test.user",
		},
	}

	result, err := taskTools.HandleCreateTaskWithContext(ctx, session, params)
	if err != nil {
		t.Fatalf("HandleCreateTaskWithContext failed: %v", err)
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
	if _, ok := meta["task"]; !ok {
		t.Error("Meta missing task")
	}
	if _, ok := meta["success"]; !ok {
		t.Error("Meta missing success")
	}
}

func TestTaskTools_HandleGetTaskDetails(t *testing.T) {
	server := createMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	taskTools := NewTaskTools(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[GetTaskDetailsParams]{
		Arguments: GetTaskDetailsParams{
			TaskID: "task-1",
		},
	}

	result, err := taskTools.HandleGetTaskDetails(ctx, session, params)
	if err != nil {
		t.Fatalf("HandleGetTaskDetails failed: %v", err)
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
	if _, ok := meta["task"]; !ok {
		t.Error("Meta missing task")
	}
	if _, ok := meta["notes"]; !ok {
		t.Error("Meta missing notes")
	}
	if _, ok := meta["project"]; !ok {
		t.Error("Meta missing project")
	}
	if _, ok := meta["has_project"]; !ok {
		t.Error("Meta missing has_project")
	}
}

func TestTaskTools_HandleUpdateTaskProgress(t *testing.T) {
	server := createMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	taskTools := NewTaskTools(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[UpdateTaskProgressParams]{
		Arguments: UpdateTaskProgressParams{
			TaskID:       "task-1",
			Status:       "Complete",
			Priority:     "High",
			ProgressNote: "Task completed successfully",
			UpdatedBy:    "test.user",
		},
	}

	result, err := taskTools.HandleUpdateTaskProgress(ctx, session, params)
	if err != nil {
		t.Fatalf("HandleUpdateTaskProgress failed: %v", err)
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
	if _, ok := meta["task"]; !ok {
		t.Error("Meta missing task")
	}
	if _, ok := meta["changes_made"]; !ok {
		t.Error("Meta missing changes_made")
	}
	if _, ok := meta["update_success"]; !ok {
		t.Error("Meta missing update_success")
	}
}

func TestTaskTools_HandleGetTaskOverview_EmptyParams(t *testing.T) {
	server := createMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	taskTools := NewTaskTools(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[GetTaskOverviewParams]{
		Arguments: GetTaskOverviewParams{},
	}

	result, err := taskTools.HandleGetTaskOverview(ctx, session, params)
	if err != nil {
		t.Fatalf("HandleGetTaskOverview failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}
}

func TestTaskTools_HandleCreateTaskWithContext_MissingRequiredFields(t *testing.T) {
	server := createMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	taskTools := NewTaskTools(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}

	// Test missing task_name
	params := &mcp.CallToolParamsFor[CreateTaskWithContextParams]{
		Arguments: CreateTaskWithContextParams{
			InitialNote: "Initial note",
			CreatedBy:   "test.user",
		},
	}

	_, err := taskTools.HandleCreateTaskWithContext(ctx, session, params)
	if err == nil {
		t.Fatal("Expected error for missing task_name")
	}

	// Test missing initial_note
	params.Arguments.TaskName = "Test Task"
	params.Arguments.InitialNote = ""

	_, err = taskTools.HandleCreateTaskWithContext(ctx, session, params)
	if err == nil {
		t.Fatal("Expected error for missing initial_note")
	}

	// Test missing created_by
	params.Arguments.InitialNote = "Initial note"
	params.Arguments.CreatedBy = ""

	_, err = taskTools.HandleCreateTaskWithContext(ctx, session, params)
	if err == nil {
		t.Fatal("Expected error for missing created_by")
	}
}

func TestTaskTools_HandleGetTaskDetails_MissingTaskID(t *testing.T) {
	server := createMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	taskTools := NewTaskTools(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.CallToolParamsFor[GetTaskDetailsParams]{
		Arguments: GetTaskDetailsParams{
			TaskID: "",
		},
	}

	_, err := taskTools.HandleGetTaskDetails(ctx, session, params)
	if err == nil {
		t.Fatal("Expected error for missing task_id")
	}
}

func TestTaskTools_HandleUpdateTaskProgress_MissingRequiredFields(t *testing.T) {
	server := createMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	taskTools := NewTaskTools(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}

	// Test missing task_id
	params := &mcp.CallToolParamsFor[UpdateTaskProgressParams]{
		Arguments: UpdateTaskProgressParams{
			ProgressNote: "Progress note",
			UpdatedBy:    "test.user",
		},
	}

	_, err := taskTools.HandleUpdateTaskProgress(ctx, session, params)
	if err == nil {
		t.Fatal("Expected error for missing task_id")
	}

	// Test missing progress_note
	params.Arguments.TaskID = "task-1"
	params.Arguments.ProgressNote = ""

	_, err = taskTools.HandleUpdateTaskProgress(ctx, session, params)
	if err == nil {
		t.Fatal("Expected error for missing progress_note")
	}

	// Test missing updated_by
	params.Arguments.ProgressNote = "Progress note"
	params.Arguments.UpdatedBy = ""

	_, err = taskTools.HandleUpdateTaskProgress(ctx, session, params)
	if err == nil {
		t.Fatal("Expected error for missing updated_by")
	}
}