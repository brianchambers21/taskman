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

// Mock API server for project resources testing
func createProjectResourcesMockAPIServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/api/v1/projects":
			projects := []Project{
				{
					ProjectID:          "proj-1",
					ProjectName:        "Project 1",
					ProjectDescription: stringPtr("Description for project 1"),
					CreatedBy:          "admin",
					CreationDate:       "2024-01-01T09:00:00Z",
				},
				{
					ProjectID:    "proj-2",
					ProjectName:  "Project 2",
					CreatedBy:    "user1",
					CreationDate: "2024-01-02T09:00:00Z",
				},
			}
			json.NewEncoder(w).Encode(projects)

		case r.Method == "GET" && r.URL.Path == "/api/v1/projects/proj-1":
			project := Project{
				ProjectID:          "proj-1",
				ProjectName:        "Test Project 1",
				ProjectDescription: stringPtr("Test project description"),
				CreatedBy:          "admin",
				CreationDate:       "2024-01-01T09:00:00Z",
			}
			json.NewEncoder(w).Encode(project)

		case r.Method == "GET" && r.URL.Path == "/api/v1/projects/proj-1/tasks":
			tasks := []Task{
				{
					TaskID:       "task-1",
					TaskName:     "Project Task 1",
					Status:       "In Progress",
					Priority:     stringPtr("High"),
					AssignedTo:   stringPtr("user1"),
					ProjectID:    stringPtr("proj-1"),
					CreatedBy:    "admin",
					CreationDate: "2024-01-01T10:00:00Z",
				},
				{
					TaskID:       "task-2",
					TaskName:     "Project Task 2",
					Status:       "Complete",
					Priority:     stringPtr("Medium"),
					AssignedTo:   stringPtr("user2"),
					ProjectID:    stringPtr("proj-1"),
					CreatedBy:    "admin",
					CreationDate: "2024-01-02T10:00:00Z",
				},
				{
					TaskID:       "task-3",
					TaskName:     "Project Task 3",
					Status:       "Not Started",
					Priority:     stringPtr("Low"),
					ProjectID:    stringPtr("proj-1"),
					CreatedBy:    "admin",
					CreationDate: "2024-01-03T10:00:00Z",
				},
			}
			json.NewEncoder(w).Encode(tasks)

		default:
			http.NotFound(w, r)
		}
	}))
}

func TestProjectResources_HandleProjectResource(t *testing.T) {
	server := createProjectResourcesMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	projectResources := NewProjectResources(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.ReadResourceParams{
		URI: "taskman://project/proj-1",
	}

	result, err := projectResources.HandleProjectResource(ctx, session, params)
	if err != nil {
		t.Fatalf("HandleProjectResource failed: %v", err)
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

	// Check that content contains expected project information
	text := content.Text
	if !contains(text, "Test Project 1") {
		t.Error("Content missing project name")
	}
	if !contains(text, "proj-1") {
		t.Error("Content missing project ID")
	}
	if !contains(text, "admin") {
		t.Error("Content missing creator")
	}
	if !contains(text, "**Total Tasks:** 3") {
		t.Error("Content missing task count")
	}
	if !contains(text, "**Completion:** 33.3%") {
		t.Error("Content missing completion percentage")
	}
	if !contains(text, "Status Breakdown") {
		t.Error("Content missing status breakdown")
	}
}

func TestProjectResources_HandleProjectsOverviewResource(t *testing.T) {
	server := createProjectResourcesMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	projectResources := NewProjectResources(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.ReadResourceParams{
		URI: "taskman://projects/overview",
	}

	result, err := projectResources.HandleProjectsOverviewResource(ctx, session, params)
	if err != nil {
		t.Fatalf("HandleProjectsOverviewResource failed: %v", err)
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

	// Check that content contains expected overview information
	text := content.Text
	if !contains(text, "Projects Overview") {
		t.Error("Content missing overview title")
	}
	if !contains(text, "**Total Projects:** 2") {
		t.Error("Content missing total project count")
	}
	if !contains(text, "Projects by Creator") {
		t.Error("Content missing creator breakdown")
	}
	if !contains(text, "All Projects") {
		t.Error("Content missing projects list")
	}
	if !contains(text, "Project 1") {
		t.Error("Content missing project name")
	}
}

func TestProjectResources_HandleProjectTasksResource(t *testing.T) {
	server := createProjectResourcesMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	projectResources := NewProjectResources(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.ReadResourceParams{
		URI: "taskman://project/proj-1/tasks",
	}

	result, err := projectResources.HandleProjectTasksResource(ctx, session, params)
	if err != nil {
		t.Fatalf("HandleProjectTasksResource failed: %v", err)
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

	// Check that content contains expected project tasks information
	text := content.Text
	if !contains(text, "Tasks in Project: Test Project 1") {
		t.Error("Content missing project tasks title")
	}
	if !contains(text, "**Total Tasks:** 3") {
		t.Error("Content missing total task count")
	}
	if !contains(text, "In Progress (1)") {
		t.Error("Content missing status section")
	}
	if !contains(text, "Complete (1)") {
		t.Error("Content missing completed section")
	}
	if !contains(text, "Not Started (1)") {
		t.Error("Content missing not started section")
	}
	if !contains(text, "Project Task 1") {
		t.Error("Content missing task name")
	}
}

func TestProjectResources_HandleProjectResource_InvalidURI(t *testing.T) {
	server := createProjectResourcesMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	projectResources := NewProjectResources(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.ReadResourceParams{
		URI: "taskman://invalid",
	}

	_, err := projectResources.HandleProjectResource(ctx, session, params)
	if err == nil {
		t.Fatal("Expected error for invalid URI")
	}
}

func TestProjectResources_HandleProjectResource_EmptyProjectID(t *testing.T) {
	server := createProjectResourcesMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	projectResources := NewProjectResources(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.ReadResourceParams{
		URI: "taskman://project/",
	}

	_, err := projectResources.HandleProjectResource(ctx, session, params)
	if err == nil {
		t.Fatal("Expected error for empty project ID")
	}
}

func TestProjectResources_HandleProjectTasksResource_InvalidURI(t *testing.T) {
	server := createProjectResourcesMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	projectResources := NewProjectResources(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.ReadResourceParams{
		URI: "taskman://project/tasks",
	}

	_, err := projectResources.HandleProjectTasksResource(ctx, session, params)
	if err == nil {
		t.Fatal("Expected error for invalid project tasks URI")
	}
}

func TestProjectResources_HandleProjectTasksResource_EmptyProjectID(t *testing.T) {
	server := createProjectResourcesMockAPIServer()
	defer server.Close()

	apiClient := client.NewAPIClient(server.URL, 30*time.Second)
	projectResources := NewProjectResources(apiClient)

	ctx := context.Background()
	session := &mcp.ServerSession{}
	params := &mcp.ReadResourceParams{
		URI: "taskman://project//tasks",
	}

	_, err := projectResources.HandleProjectTasksResource(ctx, session, params)
	if err == nil {
		t.Fatal("Expected error for empty project ID")
	}
}
