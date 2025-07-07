package tools

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

// Integration test server that mimics real API behavior
func createIntegrationAPIServer() *httptest.Server {
	// In-memory storage for testing
	tasks := make(map[string]Task)
	projects := make(map[string]Project)
	notes := make(map[string][]TaskNote)

	// Initialize with some test data
	tasks["task-1"] = Task{
		TaskID:          "task-1",
		TaskName:        "Integration Test Task",
		TaskDescription: stringPtr("Test task for integration testing"),
		Status:          "Not Started",
		Priority:        stringPtr("Medium"),
		AssignedTo:      stringPtr("test.user"),
		ProjectID:       stringPtr("proj-1"),
		DueDate:         stringPtr("2024-12-31T23:59:59Z"),
		CreatedBy:       "admin",
		CreationDate:    "2024-01-01T10:00:00Z",
	}

	projects["proj-1"] = Project{
		ProjectID:          "proj-1",
		ProjectName:        "Integration Test Project",
		ProjectDescription: stringPtr("Test project for integration testing"),
		CreatedBy:          "admin",
		CreationDate:       "2024-01-01T09:00:00Z",
	}

	notes["task-1"] = []TaskNote{
		{
			NoteID:       "note-1",
			TaskID:       "task-1",
			Note:         "Initial task note",
			CreatedBy:    "admin",
			CreationDate: "2024-01-01T10:30:00Z",
		},
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/api/v1/tasks":
			// Return all tasks or filter by query params
			status := r.URL.Query().Get("status")
			assignedTo := r.URL.Query().Get("assigned_to")
			projectID := r.URL.Query().Get("project_id")

			var result []Task
			for _, task := range tasks {
				include := true
				if status != "" && task.Status != status {
					include = false
				}
				if assignedTo != "" && (task.AssignedTo == nil || *task.AssignedTo != assignedTo) {
					include = false
				}
				if projectID != "" && (task.ProjectID == nil || *task.ProjectID != projectID) {
					include = false
				}
				if include {
					result = append(result, task)
				}
			}
			json.NewEncoder(w).Encode(result)

		case r.Method == "GET" && r.URL.Path == "/api/v1/projects":
			var result []Project
			for _, project := range projects {
				result = append(result, project)
			}
			json.NewEncoder(w).Encode(result)

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
			if dueDate, ok := req["due_date"]; ok {
				task.DueDate = stringPtr(dueDate.(string))
			}

			tasks[taskID] = task
			notes[taskID] = []TaskNote{}
			json.NewEncoder(w).Encode(task)

		case r.Method == "GET" && len(r.URL.Path) > 14 && r.URL.Path[:14] == "/api/v1/tasks/" && (len(r.URL.Path) < 6 || r.URL.Path[len(r.URL.Path)-6:] != "/notes"):
			taskID := r.URL.Path[14:] // Extract task ID from path
			if task, exists := tasks[taskID]; exists {
				json.NewEncoder(w).Encode(task)
			} else {
				http.NotFound(w, r)
			}

		case r.Method == "PUT" && len(r.URL.Path) > 14 && r.URL.Path[:14] == "/api/v1/tasks/" && (len(r.URL.Path) < 6 || r.URL.Path[len(r.URL.Path)-6:] != "/notes"):
			taskID := r.URL.Path[14:] // Extract task ID from path
			if task, exists := tasks[taskID]; exists {
				var req map[string]interface{}
				json.NewDecoder(r.Body).Decode(&req)

				// Update task fields
				if status, ok := req["status"]; ok {
					task.Status = status.(string)
				}
				if priority, ok := req["priority"]; ok {
					task.Priority = stringPtr(priority.(string))
				}
				if assignedTo, ok := req["assigned_to"]; ok {
					task.AssignedTo = stringPtr(assignedTo.(string))
				}
				if updatedBy, ok := req["last_updated_by"]; ok {
					task.LastUpdatedBy = stringPtr(updatedBy.(string))
				}
				if completionDate, ok := req["completion_date"]; ok {
					task.CompletionDate = stringPtr(completionDate.(string))
				}
				if startDate, ok := req["start_date"]; ok {
					task.StartDate = stringPtr(startDate.(string))
				}

				task.LastUpdateDate = stringPtr(time.Now().Format(time.RFC3339))
				tasks[taskID] = task
				json.NewEncoder(w).Encode(task)
			} else {
				http.NotFound(w, r)
			}

		case r.Method == "GET" && len(r.URL.Path) > 25 && r.URL.Path[:14] == "/api/v1/tasks/" && r.URL.Path[len(r.URL.Path)-6:] == "/notes":
			taskID := r.URL.Path[14 : len(r.URL.Path)-6] // Extract task ID from path
			if taskNotes, exists := notes[taskID]; exists {
				json.NewEncoder(w).Encode(taskNotes)
			} else {
				json.NewEncoder(w).Encode([]TaskNote{})
			}

		case r.Method == "POST" && len(r.URL.Path) > 25 && r.URL.Path[:14] == "/api/v1/tasks/" && r.URL.Path[len(r.URL.Path)-6:] == "/notes":
			taskID := r.URL.Path[14 : len(r.URL.Path)-6] // Extract task ID from path
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)

			noteID := "note-" + time.Now().Format("20060102150405")
			note := TaskNote{
				NoteID:       noteID,
				TaskID:       taskID,
				Note:         req["note"].(string),
				CreatedBy:    req["created_by"].(string),
				CreationDate: time.Now().Format(time.RFC3339),
			}

			if taskNotes, exists := notes[taskID]; exists {
				notes[taskID] = append(taskNotes, note)
			} else {
				notes[taskID] = []TaskNote{note}
			}
			json.NewEncoder(w).Encode(note)

		case r.Method == "GET" && len(r.URL.Path) > 17 && r.URL.Path[:17] == "/api/v1/projects/":
			projectID := r.URL.Path[17:] // Extract project ID from path
			if project, exists := projects[projectID]; exists {
				json.NewEncoder(w).Encode(project)
			} else {
				http.NotFound(w, r)
			}

		default:
			http.NotFound(w, r)
		}
	}))
}

func TestTaskTools_IntegrationWorkflow(t *testing.T) {
	server := createIntegrationAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	taskTools := NewTaskTools(apiClient)
	ctx := context.Background()
	session := &mcp.ServerSession{}

	// Test 1: Get initial task overview
	t.Run("GetInitialOverview", func(t *testing.T) {
		params := &mcp.CallToolParamsFor[GetTaskOverviewParams]{
			Arguments: GetTaskOverviewParams{},
		}

		result, err := taskTools.HandleGetTaskOverview(ctx, session, params)
		if err != nil {
			t.Fatalf("Failed to get task overview: %v", err)
		}

		if result == nil || len(result.Content) == 0 {
			t.Fatal("No content in overview result")
		}

		meta := result.Meta
		if totalTasks, ok := meta["total_tasks"]; !ok || totalTasks.(int) != 1 {
			t.Errorf("Expected 1 task, got %v", totalTasks)
		}
	})

	// Test 2: Create a new task with context
	t.Run("CreateTaskWithContext", func(t *testing.T) {
		params := &mcp.CallToolParamsFor[CreateTaskWithContextParams]{
			Arguments: CreateTaskWithContextParams{
				TaskName:        "Integration Test New Task",
				TaskDescription: "A task created during integration testing",
				Status:          "Not Started",
				Priority:        "High",
				AssignedTo:      "integration.tester",
				ProjectID:       "proj-1",
				InitialNote:     "Starting work on this integration test task",
				CreatedBy:       "integration.test",
			},
		}

		result, err := taskTools.HandleCreateTaskWithContext(ctx, session, params)
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}

		if result == nil || len(result.Content) == 0 {
			t.Fatal("No content in create result")
		}

		meta := result.Meta
		if success, ok := meta["success"]; !ok || !success.(bool) {
			t.Error("Task creation was not successful")
		}

		// Extract task ID for next tests
		taskData := meta["task"].(Task)
		if taskData.TaskID == "" {
			t.Fatal("No task ID returned")
		}
	})

	// Test 3: Get details of an existing task
	t.Run("GetTaskDetails", func(t *testing.T) {
		params := &mcp.CallToolParamsFor[GetTaskDetailsParams]{
			Arguments: GetTaskDetailsParams{
				TaskID: "task-1",
			},
		}

		result, err := taskTools.HandleGetTaskDetails(ctx, session, params)
		if err != nil {
			t.Fatalf("Failed to get task details: %v", err)
		}

		if result == nil || len(result.Content) == 0 {
			t.Fatal("No content in details result")
		}

		meta := result.Meta
		if hasProject, ok := meta["has_project"]; !ok || !hasProject.(bool) {
			t.Error("Expected task to have project")
		}

		if noteCount, ok := meta["note_count"]; !ok || noteCount.(int) != 1 {
			t.Errorf("Expected 1 note, got %v", noteCount)
		}
	})

	// Test 4: Update task progress
	t.Run("UpdateTaskProgress", func(t *testing.T) {
		params := &mcp.CallToolParamsFor[UpdateTaskProgressParams]{
			Arguments: UpdateTaskProgressParams{
				TaskID:       "task-1",
				Status:       "In Progress",
				Priority:     "High",
				ProgressNote: "Started working on this task during integration test",
				UpdatedBy:    "integration.test",
			},
		}

		result, err := taskTools.HandleUpdateTaskProgress(ctx, session, params)
		if err != nil {
			t.Fatalf("Failed to update task progress: %v", err)
		}

		if result == nil || len(result.Content) == 0 {
			t.Fatal("No content in update result")
		}

		meta := result.Meta
		if updateSuccess, ok := meta["update_success"]; !ok || !updateSuccess.(bool) {
			t.Error("Task update was not successful")
		}

		if noteAdded, ok := meta["note_added"]; !ok || !noteAdded.(bool) {
			t.Error("Progress note was not added")
		}

		changes := meta["changes_made"].([]string)
		if len(changes) == 0 {
			t.Error("No changes were recorded")
		}
	})

	// Test 5: Get updated overview after changes
	t.Run("GetUpdatedOverview", func(t *testing.T) {
		params := &mcp.CallToolParamsFor[GetTaskOverviewParams]{
			Arguments: GetTaskOverviewParams{
				Status: "In Progress",
			},
		}

		result, err := taskTools.HandleGetTaskOverview(ctx, session, params)
		if err != nil {
			t.Fatalf("Failed to get updated overview: %v", err)
		}

		if result == nil || len(result.Content) == 0 {
			t.Fatal("No content in updated overview result")
		}

		meta := result.Meta
		if totalTasks, ok := meta["total_tasks"]; !ok || totalTasks.(int) != 1 {
			t.Errorf("Expected 1 'In Progress' task, got %v", totalTasks)
		}
	})

	// Test 6: Complete the task
	t.Run("CompleteTask", func(t *testing.T) {
		params := &mcp.CallToolParamsFor[UpdateTaskProgressParams]{
			Arguments: UpdateTaskProgressParams{
				TaskID:       "task-1",
				Status:       "Complete",
				ProgressNote: "Task completed successfully during integration test",
				UpdatedBy:    "integration.test",
			},
		}

		result, err := taskTools.HandleUpdateTaskProgress(ctx, session, params)
		if err != nil {
			t.Fatalf("Failed to complete task: %v", err)
		}

		if result == nil || len(result.Content) == 0 {
			t.Fatal("No content in completion result")
		}

		meta := result.Meta
		changes := meta["changes_made"].([]string)
		hasStatusChange := false
		hasCompletionDate := false

		for _, change := range changes {
			if change == "Status: In Progress → Complete" {
				hasStatusChange = true
			}
			if change == "Completion date set" {
				hasCompletionDate = true
			}
		}

		if !hasStatusChange {
			t.Error("Status change not recorded")
		}
		if !hasCompletionDate {
			t.Error("Completion date not set")
		}
	})
}

func TestTaskTools_IntegrationErrorHandling(t *testing.T) {
	server := createIntegrationAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	taskTools := NewTaskTools(apiClient)
	ctx := context.Background()
	session := &mcp.ServerSession{}

	// Test error handling for non-existent task
	t.Run("GetNonExistentTask", func(t *testing.T) {
		params := &mcp.CallToolParamsFor[GetTaskDetailsParams]{
			Arguments: GetTaskDetailsParams{
				TaskID: "task-nonexistent",
			},
		}

		_, err := taskTools.HandleGetTaskDetails(ctx, session, params)
		if err == nil {
			t.Fatal("Expected error for non-existent task")
		}
	})

	// Test error handling for updating non-existent task
	t.Run("UpdateNonExistentTask", func(t *testing.T) {
		params := &mcp.CallToolParamsFor[UpdateTaskProgressParams]{
			Arguments: UpdateTaskProgressParams{
				TaskID:       "task-nonexistent",
				Status:       "Complete",
				ProgressNote: "This should fail",
				UpdatedBy:    "test.user",
			},
		}

		_, err := taskTools.HandleUpdateTaskProgress(ctx, session, params)
		if err == nil {
			t.Fatal("Expected error for non-existent task")
		}
	})
}
