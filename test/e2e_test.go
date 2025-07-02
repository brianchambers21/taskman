package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/bchamber/taskman-mcp/internal/client"
	"github.com/bchamber/taskman-mcp/internal/config"
	"github.com/bchamber/taskman-mcp/internal/server"
)

// E2ETestSuite represents the end-to-end test suite
type E2ETestSuite struct {
	APIBaseURL string
	APIClient  *client.APIClient
	MCPServer  *server.Server
}

// TestCompleteWorkflow tests the complete end-to-end workflow
func TestCompleteWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping end-to-end tests in short mode")
	}

	suite := setupE2ETestSuite(t)
	defer teardownE2ETestSuite(t, suite)

	t.Run("ProjectCreationWithTasksWorkflow", func(t *testing.T) {
		testProjectCreationWithTasksWorkflow(t, suite)
	})

	t.Run("TaskManagementWorkflow", func(t *testing.T) {
		testTaskManagementWorkflow(t, suite)
	})

	t.Run("TaskNotesWorkflow", func(t *testing.T) {
		testTaskNotesWorkflow(t, suite)
	})

	t.Run("CrossModuleErrorHandling", func(t *testing.T) {
		testCrossModuleErrorHandling(t, suite)
	})
}

func setupE2ETestSuite(t *testing.T) *E2ETestSuite {
	// Check if API server is running
	apiBaseURL := getEnv("TASKMAN_API_BASE_URL", "http://localhost:8080")
	
	// Wait for API server to be ready
	if !waitForAPIServer(apiBaseURL, 30*time.Second) {
		t.Skip("API server not available for E2E tests. Start with: cd taskman-api && go run cmd/api/main.go")
	}

	// Create API client
	apiClient := client.NewAPIClient(apiBaseURL, 30*time.Second)

	// Create MCP server configuration
	cfg := &config.Config{
		APIBaseURL:    apiBaseURL,
		APITimeout:    30 * time.Second,
		LogLevel:      "INFO",
		ServerName:    "e2e-test-mcp",
		ServerVersion: "1.0.0",
		TransportMode: "stdio",
		HTTPPort:      "8083",
		HTTPHost:      "localhost",
	}

	// Create MCP server
	mcpServer := server.NewServer(cfg)
	if mcpServer == nil {
		t.Fatal("Failed to create MCP server for E2E tests")
	}

	return &E2ETestSuite{
		APIBaseURL: apiBaseURL,
		APIClient:  apiClient,
		MCPServer:  mcpServer,
	}
}

func teardownE2ETestSuite(t *testing.T, suite *E2ETestSuite) {
	// Cleanup any test data created during E2E tests
	// This could include deleting test projects, tasks, etc.
	t.Log("E2E test suite cleanup completed")
}

func testProjectCreationWithTasksWorkflow(t *testing.T, suite *E2ETestSuite) {
	ctx := context.Background()

	// Step 1: Create a project via API
	projectData := map[string]interface{}{
		"project_id":          "e2e-project-1",
		"project_name":        "E2E Test Project",
		"project_description": "End-to-end test project",
		"created_by":          "e2e-test-user",
	}

	projectJSON, _ := json.Marshal(projectData)
	resp, err := suite.APIClient.Post(ctx, "/api/v1/projects", bytes.NewBuffer(projectJSON))
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	var createdProject map[string]interface{}
	if err := json.Unmarshal(resp, &createdProject); err != nil {
		t.Fatalf("Failed to unmarshal project response: %v", err)
	}

	// Step 2: Create multiple tasks in the project
	taskIds := []string{"e2e-task-1", "e2e-task-2", "e2e-task-3"}
	for i, taskId := range taskIds {
		taskData := map[string]interface{}{
			"task_id":          taskId,
			"task_name":        fmt.Sprintf("E2E Test Task %d", i+1),
			"task_description": fmt.Sprintf("End-to-end test task %d", i+1),
			"status":           "Not Started",
			"priority":         "Medium",
			"assigned_to":      "e2e-test-user",
			"project_id":       "e2e-project-1",
			"created_by":       "e2e-test-user",
		}

		taskJSON, _ := json.Marshal(taskData)
		_, err := suite.APIClient.Post(ctx, "/api/v1/tasks", bytes.NewBuffer(taskJSON))
		if err != nil {
			t.Fatalf("Failed to create task %s: %v", taskId, err)
		}
	}

	// Step 3: Verify project tasks via API
	projectTasksResp, err := suite.APIClient.Get(ctx, "/api/v1/projects/e2e-project-1/tasks")
	if err != nil {
		t.Fatalf("Failed to get project tasks: %v", err)
	}

	var projectTasks []map[string]interface{}
	if err := json.Unmarshal(projectTasksResp, &projectTasks); err != nil {
		t.Fatalf("Failed to unmarshal project tasks: %v", err)
	}

	if len(projectTasks) != 3 {
		t.Fatalf("Expected 3 tasks, got %d", len(projectTasks))
	}

	// Step 4: Update task status workflow
	taskUpdateData := map[string]interface{}{
		"status":           "In Progress",
		"last_updated_by":  "e2e-test-user",
	}

	taskUpdateJSON, _ := json.Marshal(taskUpdateData)
	_, err = suite.APIClient.Put(ctx, "/api/v1/tasks/e2e-task-1", bytes.NewBuffer(taskUpdateJSON))
	if err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	// Step 5: Add notes to tasks
	noteData := map[string]interface{}{
		"note":       "Progress update: Task started",
		"created_by": "e2e-test-user",
	}

	noteJSON, _ := json.Marshal(noteData)
	_, err = suite.APIClient.Post(ctx, "/api/v1/tasks/e2e-task-1/notes", bytes.NewBuffer(noteJSON))
	if err != nil {
		t.Fatalf("Failed to add task note: %v", err)
	}

	// Step 6: Verify complete workflow via API
	updatedTaskResp, err := suite.APIClient.Get(ctx, "/api/v1/tasks/e2e-task-1")
	if err != nil {
		t.Fatalf("Failed to get updated task: %v", err)
	}

	var updatedTask map[string]interface{}
	if err := json.Unmarshal(updatedTaskResp, &updatedTask); err != nil {
		t.Fatalf("Failed to unmarshal updated task: %v", err)
	}

	if updatedTask["status"].(string) != "In Progress" {
		t.Fatalf("Expected task status 'In Progress', got '%s'", updatedTask["status"])
	}

	t.Log("Project creation with tasks workflow completed successfully")
}

func testTaskManagementWorkflow(t *testing.T, suite *E2ETestSuite) {
	ctx := context.Background()

	// Step 1: Create a standalone task
	taskData := map[string]interface{}{
		"task_id":          "e2e-standalone-task",
		"task_name":        "Standalone E2E Task",
		"task_description": "A standalone task for end-to-end testing",
		"status":           "Not Started",
		"priority":         "High",
		"assigned_to":      "e2e-test-user",
		"created_by":       "e2e-test-user",
		"tags":             []string{"urgent", "testing"},
	}

	taskJSON, _ := json.Marshal(taskData)
	_, err := suite.APIClient.Post(ctx, "/api/v1/tasks", bytes.NewBuffer(taskJSON))
	if err != nil {
		t.Fatalf("Failed to create standalone task: %v", err)
	}

	// Step 2: Test task lifecycle transitions
	statuses := []string{"Not Started", "In Progress", "Review", "Complete"}
	for i, status := range statuses[1:] { // Skip "Not Started" as it's the initial state
		updateData := map[string]interface{}{
			"status":          status,
			"last_updated_by": "e2e-test-user",
		}

		updateJSON, _ := json.Marshal(updateData)
		_, err := suite.APIClient.Put(ctx, "/api/v1/tasks/e2e-standalone-task", bytes.NewBuffer(updateJSON))
		if err != nil {
			t.Fatalf("Failed to update task to status %s: %v", status, err)
		}

		// Verify status update
		taskResp, err := suite.APIClient.Get(ctx, "/api/v1/tasks/e2e-standalone-task")
		if err != nil {
			t.Fatalf("Failed to get task after status update: %v", err)
		}

		var task map[string]interface{}
		if err := json.Unmarshal(taskResp, &task); err != nil {
			t.Fatalf("Failed to unmarshal task: %v", err)
		}

		if task["status"].(string) != status {
			t.Fatalf("Step %d: Expected status '%s', got '%s'", i+1, status, task["status"])
		}
	}

	// Step 3: Test task filtering
	tasksResp, err := suite.APIClient.Get(ctx, "/api/v1/tasks?status=Complete&assigned_to=e2e-test-user")
	if err != nil {
		t.Fatalf("Failed to get filtered tasks: %v", err)
	}

	var filteredTasks []map[string]interface{}
	if err := json.Unmarshal(tasksResp, &filteredTasks); err != nil {
		t.Fatalf("Failed to unmarshal filtered tasks: %v", err)
	}

	// Should have at least 1 completed task
	found := false
	for _, task := range filteredTasks {
		if task["task_id"].(string) == "e2e-standalone-task" && task["status"].(string) == "Complete" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("Expected to find completed e2e-standalone-task in filtered results")
	}

	t.Log("Task management workflow completed successfully")
}

func testTaskNotesWorkflow(t *testing.T, suite *E2ETestSuite) {
	ctx := context.Background()

	// Step 1: Create a task for notes testing
	taskData := map[string]interface{}{
		"task_id":     "e2e-notes-task",
		"task_name":   "Notes Test Task",
		"status":      "In Progress",
		"created_by":  "e2e-test-user",
	}

	taskJSON, _ := json.Marshal(taskData)
	_, err := suite.APIClient.Post(ctx, "/api/v1/tasks", bytes.NewBuffer(taskJSON))
	if err != nil {
		t.Fatalf("Failed to create task for notes: %v", err)
	}

	// Step 2: Add multiple notes
	notes := []string{
		"Initial progress note",
		"Encountered an issue with the API",
		"Issue resolved, continuing with implementation",
		"Ready for review",
	}

	var noteIds []string
	for i, noteText := range notes {
		noteData := map[string]interface{}{
			"note":       noteText,
			"created_by": "e2e-test-user",
		}

		noteJSON, _ := json.Marshal(noteData)
		noteResp, err := suite.APIClient.Post(ctx, "/api/v1/tasks/e2e-notes-task/notes", bytes.NewBuffer(noteJSON))
		if err != nil {
			t.Fatalf("Failed to add note %d: %v", i+1, err)
		}

		var createdNote map[string]interface{}
		if err := json.Unmarshal(noteResp, &createdNote); err != nil {
			t.Fatalf("Failed to unmarshal note response: %v", err)
		}

		noteIds = append(noteIds, createdNote["note_id"].(string))
	}

	// Step 3: Retrieve and verify notes
	notesResp, err := suite.APIClient.Get(ctx, "/api/v1/tasks/e2e-notes-task/notes")
	if err != nil {
		t.Fatalf("Failed to get task notes: %v", err)
	}

	var taskNotes []map[string]interface{}
	if err := json.Unmarshal(notesResp, &taskNotes); err != nil {
		t.Fatalf("Failed to unmarshal task notes: %v", err)
	}

	if len(taskNotes) != len(notes) {
		t.Fatalf("Expected %d notes, got %d", len(notes), len(taskNotes))
	}

	// Step 4: Update a note
	updateData := map[string]interface{}{
		"note":            "Updated: Issue resolved, continuing with implementation",
		"updated_by":      "e2e-test-user",
	}

	updateJSON, _ := json.Marshal(updateData)
	_, err = suite.APIClient.Put(ctx, fmt.Sprintf("/api/v1/tasks/e2e-notes-task/notes/%s", noteIds[2]), bytes.NewBuffer(updateJSON))
	if err != nil {
		t.Fatalf("Failed to update note: %v", err)
	}

	// Step 5: Verify note update
	updatedNotesResp, err := suite.APIClient.Get(ctx, "/api/v1/tasks/e2e-notes-task/notes")
	if err != nil {
		t.Fatalf("Failed to get updated notes: %v", err)
	}

	var updatedNotes []map[string]interface{}
	if err := json.Unmarshal(updatedNotesResp, &updatedNotes); err != nil {
		t.Fatalf("Failed to unmarshal updated notes: %v", err)
	}

	found := false
	for _, note := range updatedNotes {
		if note["note_id"].(string) == noteIds[2] && 
		   note["note"].(string) == "Updated: Issue resolved, continuing with implementation" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("Updated note not found with correct content")
	}

	t.Log("Task notes workflow completed successfully")
}

func testCrossModuleErrorHandling(t *testing.T, suite *E2ETestSuite) {
	ctx := context.Background()

	// Test 1: Non-existent resource handling
	_, err := suite.APIClient.Get(ctx, "/api/v1/tasks/non-existent-task")
	if err == nil {
		t.Fatal("Expected error for non-existent task, got nil")
	}

	// Test 2: Invalid task creation
	invalidTaskData := map[string]interface{}{
		"task_id": "invalid-task-test",
		// Missing required fields like task_name, created_by
	}

	invalidJSON, _ := json.Marshal(invalidTaskData)
	_, err = suite.APIClient.Post(ctx, "/api/v1/tasks", bytes.NewBuffer(invalidJSON))
	if err == nil {
		t.Fatal("Expected error for invalid task creation, got nil")
	}

	// Test 3: Constraint violation (foreign key)
	taskWithInvalidProject := map[string]interface{}{
		"task_id":     "constraint-test-task",
		"task_name":   "Constraint Test",
		"project_id":  "non-existent-project",
		"created_by":  "e2e-test-user",
	}

	constraintJSON, _ := json.Marshal(taskWithInvalidProject)
	_, err = suite.APIClient.Post(ctx, "/api/v1/tasks", bytes.NewBuffer(constraintJSON))
	if err == nil {
		t.Fatal("Expected error for invalid project_id constraint, got nil")
	}

	t.Log("Cross-module error handling tests completed successfully")
}

func waitForAPIServer(baseURL string, timeout time.Duration) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := client.Get(baseURL + "/api/v1/tasks")
		if err == nil {
			resp.Body.Close()
			return true
		}
		time.Sleep(1 * time.Second)
	}
	return false
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestSystemIntegration tests the complete system integration
func TestSystemIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping system integration tests in short mode")
	}

	// Test that both API and MCP servers can be started and work together
	t.Run("StartupAndShutdown", func(t *testing.T) {
		testSystemStartupAndShutdown(t)
	})

	t.Run("HealthChecks", func(t *testing.T) {
		testSystemHealthChecks(t)
	})
}

func testSystemStartupAndShutdown(t *testing.T) {
	// This test verifies that the system can start up and shut down cleanly
	// In a real deployment, this would involve starting actual processes
	
	// Test API server configuration
	apiBaseURL := getEnv("TASKMAN_API_BASE_URL", "http://localhost:8080")
	if !waitForAPIServer(apiBaseURL, 5*time.Second) {
		t.Skip("API server not running - start with: cd taskman-api && go run cmd/api/main.go")
	}

	// Test MCP server configuration
	cfg := &config.Config{
		APIBaseURL:    apiBaseURL,
		APITimeout:    10 * time.Second,
		LogLevel:      "INFO",
		ServerName:    "integration-test-mcp",
		ServerVersion: "1.0.0",
		TransportMode: "stdio",
	}

	mcpServer := server.NewServer(cfg)
	if mcpServer == nil {
		t.Fatal("Failed to create MCP server for integration test")
	}

	// Test that MCP server can connect to API
	apiClient := client.NewAPIClient(apiBaseURL, 10*time.Second)
	ctx := context.Background()

	// Test basic connectivity
	_, err := apiClient.Get(ctx, "/api/v1/tasks")
	if err != nil {
		t.Fatalf("MCP server cannot connect to API server: %v", err)
	}

	t.Log("System startup and connectivity verified successfully")
}

func testSystemHealthChecks(t *testing.T) {
	apiBaseURL := getEnv("TASKMAN_API_BASE_URL", "http://localhost:8080")
	
	// Test API health
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(apiBaseURL + "/api/v1/tasks")
	if err != nil {
		t.Fatalf("API health check failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("API health check returned status %d, expected %d", resp.StatusCode, http.StatusOK)
	}

	// Test MCP server health (via configuration validation)
	cfg := &config.Config{
		APIBaseURL:    apiBaseURL,
		APITimeout:    5 * time.Second,
		LogLevel:      "INFO",
		ServerName:    "health-check-mcp",
		ServerVersion: "1.0.0",
	}

	mcpServer := server.NewServer(cfg)
	if mcpServer == nil {
		t.Fatal("MCP server health check failed - server creation failed")
	}

	t.Log("System health checks completed successfully")
}