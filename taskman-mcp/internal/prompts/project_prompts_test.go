package prompts

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestCreateProjectPrompts(t *testing.T) {
	prompts := CreateProjectPrompts()

	expectedPrompts := []string{
		"create_project_plan",
		"project_status_review",
		"project_retrospective",
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

func TestHandleCreateProjectPlanPrompt(t *testing.T) {
	ctx := context.Background()
	session := &mcp.ServerSession{}

	// Test with development project type
	t.Run("DevelopmentProject", func(t *testing.T) {
		params := &mcp.GetPromptParams{
			Name: "create_project_plan",
			Arguments: map[string]string{
				"project_name": "User Authentication System",
				"project_type": "Development",
				"duration":     "8 weeks",
			},
		}

		result, err := handleCreateProjectPlanPrompt(ctx, session, params)
		if err != nil {
			t.Fatalf("handleCreateProjectPlanPrompt failed: %v", err)
		}

		if result == nil {
			t.Fatal("handleCreateProjectPlanPrompt returned nil result")
		}

		content, ok := result.Messages[0].Content.(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}

		// Check for project-specific content
		if !contains(content.Text, "User Authentication System") {
			t.Error("Prompt missing project name")
		}
		if !contains(content.Text, "Development") {
			t.Error("Prompt missing project type")
		}
		if !contains(content.Text, "8 weeks") {
			t.Error("Prompt missing duration")
		}

		// Check for development-specific phases
		developmentPhases := []string{
			"Planning & Design",
			"Core Development",
			"Testing & Refinement",
			"Deployment & Closure",
		}

		for _, phase := range developmentPhases {
			if !contains(content.Text, phase) {
				t.Errorf("Prompt missing development phase: %s", phase)
			}
		}
	})

	// Test with research project type
	t.Run("ResearchProject", func(t *testing.T) {
		params := &mcp.GetPromptParams{
			Name: "create_project_plan",
			Arguments: map[string]string{
				"project_name": "Market Analysis Study",
				"project_type": "Research",
			},
		}

		result, err := handleCreateProjectPlanPrompt(ctx, session, params)
		if err != nil {
			t.Fatalf("handleCreateProjectPlanPrompt failed: %v", err)
		}

		content, ok := result.Messages[0].Content.(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}

		// Check for research-specific phases
		researchPhases := []string{
			"Research Design",
			"Data Collection",
			"Analysis & Synthesis",
			"Reporting & Presentation",
		}

		for _, phase := range researchPhases {
			if !contains(content.Text, phase) {
				t.Errorf("Prompt missing research phase: %s", phase)
			}
		}
	})

	// Test with generic project type
	t.Run("GenericProject", func(t *testing.T) {
		params := &mcp.GetPromptParams{
			Name: "create_project_plan",
			Arguments: map[string]string{
				"project_name": "Process Improvement Initiative",
			},
		}

		result, err := handleCreateProjectPlanPrompt(ctx, session, params)
		if err != nil {
			t.Fatalf("handleCreateProjectPlanPrompt failed: %v", err)
		}

		content, ok := result.Messages[0].Content.(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}

		// Check for generic phases
		genericPhases := []string{
			"Planning & Preparation",
			"Core Execution",
			"Completion & Closure",
		}

		for _, phase := range genericPhases {
			if !contains(content.Text, phase) {
				t.Errorf("Prompt missing generic phase: %s", phase)
			}
		}
	})

	// Test essential sections present in all project types
	t.Run("EssentialSections", func(t *testing.T) {
		params := &mcp.GetPromptParams{
			Name: "create_project_plan",
			Arguments: map[string]string{
				"project_name": "Test Project",
			},
		}

		result, err := handleCreateProjectPlanPrompt(ctx, session, params)
		if err != nil {
			t.Fatalf("handleCreateProjectPlanPrompt failed: %v", err)
		}

		content, ok := result.Messages[0].Content.(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}

		essentialSections := []string{
			"Project Definition & Scope",
			"Stakeholder Analysis",
			"Risk Assessment & Mitigation",
			"Resource Planning",
			"High-Level Timeline & Phases",
			"Initial Task Breakdown",
			"Success Metrics & Monitoring",
			"Next Steps",
		}

		for _, section := range essentialSections {
			if !contains(content.Text, section) {
				t.Errorf("Prompt missing essential section: %s", section)
			}
		}
	})
}

func TestHandleProjectStatusReviewPrompt(t *testing.T) {
	ctx := context.Background()
	session := &mcp.ServerSession{}

	// Test with weekly review
	t.Run("WeeklyReview", func(t *testing.T) {
		params := &mcp.GetPromptParams{
			Name: "project_status_review",
			Arguments: map[string]string{
				"project_id":    "proj-123",
				"review_period": "weekly",
			},
		}

		result, err := handleProjectStatusReviewPrompt(ctx, session, params)
		if err != nil {
			t.Fatalf("handleProjectStatusReviewPrompt failed: %v", err)
		}

		if result == nil {
			t.Fatal("handleProjectStatusReviewPrompt returned nil result")
		}

		content, ok := result.Messages[0].Content.(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}

		if !contains(content.Text, "proj-123") {
			t.Error("Prompt missing project ID")
		}
		if !contains(content.Text, "weekly") {
			t.Error("Prompt missing review period")
		}
		if !contains(content.Text, "Weekly Adjustments") {
			t.Error("Prompt missing weekly-specific guidance")
		}
	})

	// Test with monthly review
	t.Run("MonthlyReview", func(t *testing.T) {
		params := &mcp.GetPromptParams{
			Name: "project_status_review",
			Arguments: map[string]string{
				"project_id":    "proj-456",
				"review_period": "monthly",
			},
		}

		result, err := handleProjectStatusReviewPrompt(ctx, session, params)
		if err != nil {
			t.Fatalf("handleProjectStatusReviewPrompt failed: %v", err)
		}

		content, ok := result.Messages[0].Content.(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}

		if !contains(content.Text, "Strategic Assessment") {
			t.Error("Prompt missing strategic assessment for monthly review")
		}
		if !contains(content.Text, "Deep Dive Analysis") {
			t.Error("Prompt missing deep dive analysis for monthly review")
		}
	})

	// Test with daily review
	t.Run("DailyReview", func(t *testing.T) {
		params := &mcp.GetPromptParams{
			Name: "project_status_review",
			Arguments: map[string]string{
				"project_id":    "proj-789",
				"review_period": "daily",
			},
		}

		result, err := handleProjectStatusReviewPrompt(ctx, session, params)
		if err != nil {
			t.Fatalf("handleProjectStatusReviewPrompt failed: %v", err)
		}

		content, ok := result.Messages[0].Content.(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}

		if !contains(content.Text, "Daily Focus Areas") {
			t.Error("Prompt missing daily focus areas")
		}
		if !contains(content.Text, "Immediate Blockers") {
			t.Error("Prompt missing immediate blockers section")
		}
	})

	// Test core sections present in all reviews
	t.Run("CoreSections", func(t *testing.T) {
		params := &mcp.GetPromptParams{
			Name: "project_status_review",
			Arguments: map[string]string{
				"project_id": "proj-core",
			},
		}

		result, err := handleProjectStatusReviewPrompt(ctx, session, params)
		if err != nil {
			t.Fatalf("handleProjectStatusReviewPrompt failed: %v", err)
		}

		content, ok := result.Messages[0].Content.(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}

		coreSections := []string{
			"Progress Assessment",
			"Health Check Indicators",
			"Issue and Risk Management",
			"Stakeholder Communication",
			"Action Items and Next Steps",
			"Review Summary",
		}

		for _, section := range coreSections {
			if !contains(content.Text, section) {
				t.Errorf("Prompt missing core section: %s", section)
			}
		}
	})
}

func TestHandleProjectRetrospectivePrompt(t *testing.T) {
	ctx := context.Background()
	session := &mcp.ServerSession{}

	params := &mcp.GetPromptParams{
		Name: "project_retrospective",
		Arguments: map[string]string{
			"project_id":      "proj-retro",
			"project_outcome": "Success",
		},
	}

	result, err := handleProjectRetrospectivePrompt(ctx, session, params)
	if err != nil {
		t.Fatalf("handleProjectRetrospectivePrompt failed: %v", err)
	}

	if result == nil {
		t.Fatal("handleProjectRetrospectivePrompt returned nil result")
	}

	content, ok := result.Messages[0].Content.(*mcp.TextContent)
	if !ok {
		t.Fatal("Expected TextContent")
	}

	// Check for project-specific content
	if !contains(content.Text, "proj-retro") {
		t.Error("Prompt missing project ID")
	}
	if !contains(content.Text, "Success") {
		t.Error("Prompt missing project outcome")
	}

	// Check for all major retrospective sections
	retrospectiveSections := []string{
		"Project Summary and Outcomes",
		"What Went Well (Continue)",
		"What Could Be Improved (Change)",
		"Lessons Learned and Insights",
		"Impact Assessment",
		"Actionable Recommendations",
		"Knowledge Transfer and Documentation",
		"Follow-up Actions and Next Steps",
		"Retrospective Summary",
	}

	for _, section := range retrospectiveSections {
		if !contains(content.Text, section) {
			t.Errorf("Prompt missing retrospective section: %s", section)
		}
	}

	// Check for specific subsections
	subsections := []string{
		"Quantitative Metrics",
		"Process Successes",
		"Standout Achievements",
		"Process Challenges",
		"Technical and Quality Issues",
		"Key Learning Themes",
		"Unexpected Discoveries",
		"Business and Organizational Impact",
		"For Future Similar Projects",
		"For Organizational Process Improvement",
	}

	for _, subsection := range subsections {
		if !contains(content.Text, subsection) {
			t.Errorf("Prompt missing subsection: %s", subsection)
		}
	}
}

func TestProjectPromptsArgumentValidation(t *testing.T) {
	ctx := context.Background()
	session := &mcp.ServerSession{}

	// Test that prompts handle missing arguments gracefully
	prompts := CreateProjectPrompts()

	for _, prompt := range prompts {
		t.Run(prompt.Prompt.Name+"_NoArguments", func(t *testing.T) {
			params := &mcp.GetPromptParams{
				Name:      prompt.Prompt.Name,
				Arguments: nil,
			}

			result, err := prompt.Handler(ctx, session, params)
			if err != nil {
				t.Fatalf("Prompt %s failed with no arguments: %v", prompt.Prompt.Name, err)
			}

			if result == nil {
				t.Fatalf("Prompt %s returned nil result", prompt.Prompt.Name)
			}

			if len(result.Messages) == 0 {
				t.Fatalf("Prompt %s returned no messages", prompt.Prompt.Name)
			}

			content, ok := result.Messages[0].Content.(*mcp.TextContent)
			if !ok {
				t.Fatalf("Prompt %s returned non-text content", prompt.Prompt.Name)
			}

			if content.Text == "" {
				t.Errorf("Prompt %s returned empty text", prompt.Prompt.Name)
			}
		})
	}
}
