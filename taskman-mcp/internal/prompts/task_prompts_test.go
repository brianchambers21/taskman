package prompts

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestCreateTaskPrompts(t *testing.T) {
	prompts := CreateTaskPrompts()
	
	expectedPrompts := []string{
		"plan_task",
		"update_task_status", 
		"task_review",
		"task_breakdown",
	}
	
	if len(prompts) != len(expectedPrompts) {
		t.Fatalf("Expected %d prompts, got %d", len(expectedPrompts), len(prompts))
	}
	
	for i, prompt := range prompts {
		if prompt.Prompt.Name != expectedPrompts[i] {
			t.Errorf("Expected prompt %d to be %s, got %s", i, expectedPrompts[i], prompt.Prompt.Name)
		}
		
		if prompt.Prompt.Description == "" {
			t.Errorf("Prompt %s missing description", prompt.Prompt.Name)
		}
		
		if prompt.Handler == nil {
			t.Errorf("Prompt %s missing handler", prompt.Prompt.Name)
		}
	}
}

func TestHandlePlanTaskPrompt(t *testing.T) {
	ctx := context.Background()
	session := &mcp.ServerSession{}
	
	// Test with all arguments provided
	t.Run("WithAllArguments", func(t *testing.T) {
		params := &mcp.GetPromptParams{
			Name: "plan_task",
			Arguments: map[string]string{
				"task_name":       "Test Task",
				"project_context": "Test Project",
				"complexity":      "Complex",
			},
		}
		
		result, err := handlePlanTaskPrompt(ctx, session, params)
		if err != nil {
			t.Fatalf("handlePlanTaskPrompt failed: %v", err)
		}
		
		if result == nil {
			t.Fatal("handlePlanTaskPrompt returned nil result")
		}
		
		if len(result.Messages) != 1 {
			t.Fatalf("Expected 1 message, got %d", len(result.Messages))
		}
		
		message := result.Messages[0]
		if message.Role != "user" {
			t.Errorf("Expected role 'user', got '%s'", message.Role)
		}
		
		content, ok := message.Content.(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}
		
		if content.Text == "" {
			t.Error("Prompt text is empty")
		}
		
		// Check for key elements in the prompt
		if !contains(content.Text, "Test Task") {
			t.Error("Prompt missing task name")
		}
		if !contains(content.Text, "Test Project") {
			t.Error("Prompt missing project context")
		}
		if !contains(content.Text, "Complex") {
			t.Error("Prompt missing complexity")
		}
		if !contains(content.Text, "Planning Checklist") {
			t.Error("Prompt missing planning checklist")
		}
	})
	
	// Test with minimal arguments
	t.Run("WithMinimalArguments", func(t *testing.T) {
		params := &mcp.GetPromptParams{
			Name: "plan_task",
			Arguments: map[string]string{
				"task_name": "Minimal Task",
			},
		}
		
		result, err := handlePlanTaskPrompt(ctx, session, params)
		if err != nil {
			t.Fatalf("handlePlanTaskPrompt failed: %v", err)
		}
		
		if result == nil {
			t.Fatal("handlePlanTaskPrompt returned nil result")
		}
		
		content, ok := result.Messages[0].Content.(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}
		
		if !contains(content.Text, "Minimal Task") {
			t.Error("Prompt missing task name")
		}
		if !contains(content.Text, "Medium") {
			t.Error("Prompt missing default complexity")
		}
	})
	
	// Test with no arguments
	t.Run("WithNoArguments", func(t *testing.T) {
		params := &mcp.GetPromptParams{
			Name:      "plan_task",
			Arguments: nil,
		}
		
		result, err := handlePlanTaskPrompt(ctx, session, params)
		if err != nil {
			t.Fatalf("handlePlanTaskPrompt failed: %v", err)
		}
		
		if result == nil {
			t.Fatal("handlePlanTaskPrompt returned nil result")
		}
		
		// Should handle gracefully with empty values
		content, ok := result.Messages[0].Content.(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}
		
		if content.Text == "" {
			t.Error("Prompt text is empty")
		}
	})
}

func TestHandleUpdateTaskStatusPrompt(t *testing.T) {
	ctx := context.Background()
	session := &mcp.ServerSession{}
	
	// Test with valid status transition
	t.Run("ValidStatusTransition", func(t *testing.T) {
		params := &mcp.GetPromptParams{
			Name: "update_task_status",
			Arguments: map[string]string{
				"task_id":        "task-123",
				"current_status": "In Progress",
				"new_status":     "Review",
			},
		}
		
		result, err := handleUpdateTaskStatusPrompt(ctx, session, params)
		if err != nil {
			t.Fatalf("handleUpdateTaskStatusPrompt failed: %v", err)
		}
		
		if result == nil {
			t.Fatal("handleUpdateTaskStatusPrompt returned nil result")
		}
		
		content, ok := result.Messages[0].Content.(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}
		
		if !contains(content.Text, "task-123") {
			t.Error("Prompt missing task ID")
		}
		if !contains(content.Text, "In Progress") {
			t.Error("Prompt missing current status")
		}
		if !contains(content.Text, "Review") {
			t.Error("Prompt missing new status")
		}
		if !contains(content.Text, "Ready for review") {
			t.Error("Prompt missing status-specific guidance")
		}
	})
	
	// Test different status transitions
	statusTests := []struct {
		name      string
		newStatus string
		expected  string
	}{
		{"ToInProgress", "In Progress", "Starting work"},
		{"ToBlocked", "Blocked", "Identifying blockers"},
		{"ToComplete", "Complete", "Task completion"},
		{"ToOther", "Not Started", "General status update"},
	}
	
	for _, tt := range statusTests {
		t.Run(tt.name, func(t *testing.T) {
			params := &mcp.GetPromptParams{
				Name: "update_task_status",
				Arguments: map[string]string{
					"task_id":        "test-task",
					"current_status": "Not Started",
					"new_status":     tt.newStatus,
				},
			}
			
			result, err := handleUpdateTaskStatusPrompt(ctx, session, params)
			if err != nil {
				t.Fatalf("handleUpdateTaskStatusPrompt failed: %v", err)
			}
			
			content, ok := result.Messages[0].Content.(*mcp.TextContent)
			if !ok {
				t.Fatal("Expected TextContent")
			}
			
			if !contains(content.Text, tt.expected) {
				t.Errorf("Prompt missing expected guidance for %s: %s", tt.newStatus, tt.expected)
			}
		})
	}
}

func TestHandleTaskReviewPrompt(t *testing.T) {
	ctx := context.Background()
	session := &mcp.ServerSession{}
	
	params := &mcp.GetPromptParams{
		Name: "task_review",
		Arguments: map[string]string{
			"task_id":         "task-456",
			"completion_date": "2024-01-15",
		},
	}
	
	result, err := handleTaskReviewPrompt(ctx, session, params)
	if err != nil {
		t.Fatalf("handleTaskReviewPrompt failed: %v", err)
	}
	
	if result == nil {
		t.Fatal("handleTaskReviewPrompt returned nil result")
	}
	
	content, ok := result.Messages[0].Content.(*mcp.TextContent)
	if !ok {
		t.Fatal("Expected TextContent")
	}
	
	expectedSections := []string{
		"task-456",
		"2024-01-15",
		"Completion Assessment",
		"Lessons Learned",
		"Knowledge Transfer",
		"Impact Assessment",
	}
	
	for _, section := range expectedSections {
		if !contains(content.Text, section) {
			t.Errorf("Prompt missing section: %s", section)
		}
	}
}

func TestHandleTaskBreakdownPrompt(t *testing.T) {
	ctx := context.Background()
	session := &mcp.ServerSession{}
	
	// Test with team size for collaboration guidance
	t.Run("WithTeamSize", func(t *testing.T) {
		params := &mcp.GetPromptParams{
			Name: "task_breakdown",
			Arguments: map[string]string{
				"parent_task": "Build User Management System",
				"timeline":    "6 weeks",
				"team_size":   "3",
			},
		}
		
		result, err := handleTaskBreakdownPrompt(ctx, session, params)
		if err != nil {
			t.Fatalf("handleTaskBreakdownPrompt failed: %v", err)
		}
		
		content, ok := result.Messages[0].Content.(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}
		
		if !contains(content.Text, "Build User Management System") {
			t.Error("Prompt missing parent task")
		}
		if !contains(content.Text, "6 weeks") {
			t.Error("Prompt missing timeline")
		}
		if !contains(content.Text, "3 people") {
			t.Error("Prompt missing team size")
		}
		if !contains(content.Text, "Team Collaboration Strategy") {
			t.Error("Prompt missing team collaboration guidance")
		}
	})
	
	// Test single person approach
	t.Run("SinglePerson", func(t *testing.T) {
		params := &mcp.GetPromptParams{
			Name: "task_breakdown",
			Arguments: map[string]string{
				"parent_task": "Simple Task",
				"team_size":   "1",
			},
		}
		
		result, err := handleTaskBreakdownPrompt(ctx, session, params)
		if err != nil {
			t.Fatalf("handleTaskBreakdownPrompt failed: %v", err)
		}
		
		content, ok := result.Messages[0].Content.(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}
		
		if !contains(content.Text, "Single Person Approach") {
			t.Error("Prompt missing single person guidance")
		}
		if !contains(content.Text, "sequential phases") {
			t.Error("Prompt missing sequential guidance")
		}
	})
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s != substr && 
		   (len(s) == len(substr) || s[len(s)-len(substr):] == substr || 
		    s[:len(substr)] == substr || 
		    containsInMiddle(s, substr))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}