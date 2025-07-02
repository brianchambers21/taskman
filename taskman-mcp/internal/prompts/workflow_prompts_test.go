package prompts

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestCreateWorkflowPrompts(t *testing.T) {
	prompts := CreateWorkflowPrompts()
	
	expectedPrompts := []string{
		"daily_standup",
		"weekly_planning",
		"task_handoff",
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

func TestHandleDailyStandupPrompt(t *testing.T) {
	ctx := context.Background()
	session := &mcp.ServerSession{}
	
	// Test individual standup
	t.Run("IndividualStandup", func(t *testing.T) {
		params := &mcp.GetPromptParams{
			Name: "daily_standup",
			Arguments: map[string]string{
				"user_id":      "john.doe",
				"standup_type": "individual",
			},
		}
		
		result, err := handleDailyStandupPrompt(ctx, session, params)
		if err != nil {
			t.Fatalf("handleDailyStandupPrompt failed: %v", err)
		}
		
		if result == nil {
			t.Fatal("handleDailyStandupPrompt returned nil result")
		}
		
		content, ok := result.Messages[0].Content.(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}
		
		if !contains(content.Text, "john.doe") {
			t.Error("Prompt missing user ID")
		}
		if !contains(content.Text, "individual") {
			t.Error("Prompt missing standup type")
		}
		if !contains(content.Text, "Personal Planning") {
			t.Error("Prompt missing individual-specific section")
		}
		if !contains(content.Text, "Time Management") {
			t.Error("Prompt missing time management section")
		}
	})
	
	// Test team standup
	t.Run("TeamStandup", func(t *testing.T) {
		params := &mcp.GetPromptParams{
			Name: "daily_standup",
			Arguments: map[string]string{
				"user_id":      "jane.smith",
				"standup_type": "team",
			},
		}
		
		result, err := handleDailyStandupPrompt(ctx, session, params)
		if err != nil {
			t.Fatalf("handleDailyStandupPrompt failed: %v", err)
		}
		
		content, ok := result.Messages[0].Content.(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}
		
		if !contains(content.Text, "Team Coordination") {
			t.Error("Prompt missing team coordination section")
		}
		if !contains(content.Text, "Collaboration Opportunities") {
			t.Error("Prompt missing collaboration opportunities")
		}
		if !contains(content.Text, "Team Dependencies") {
			t.Error("Prompt missing team dependencies")
		}
	})
	
	// Test cross-team standup
	t.Run("CrossTeamStandup", func(t *testing.T) {
		params := &mcp.GetPromptParams{
			Name: "daily_standup",
			Arguments: map[string]string{
				"user_id":      "alex.wilson",
				"standup_type": "cross-team",
			},
		}
		
		result, err := handleDailyStandupPrompt(ctx, session, params)
		if err != nil {
			t.Fatalf("handleDailyStandupPrompt failed: %v", err)
		}
		
		content, ok := result.Messages[0].Content.(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}
		
		if !contains(content.Text, "Cross-Team Updates") {
			t.Error("Prompt missing cross-team updates section")
		}
		if !contains(content.Text, "External Dependencies") {
			t.Error("Prompt missing external dependencies")
		}
		if !contains(content.Text, "Communication Priorities") {
			t.Error("Prompt missing communication priorities")
		}
	})
	
	// Test core sections present in all standup types
	t.Run("CoreSections", func(t *testing.T) {
		params := &mcp.GetPromptParams{
			Name: "daily_standup",
			Arguments: map[string]string{
				"user_id": "test.user",
			},
		}
		
		result, err := handleDailyStandupPrompt(ctx, session, params)
		if err != nil {
			t.Fatalf("handleDailyStandupPrompt failed: %v", err)
		}
		
		content, ok := result.Messages[0].Content.(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}
		
		coreSections := []string{
			"Yesterday's Accomplishments",
			"Today's Plan",
			"Blockers and Challenges",
			"Looking Ahead",
			"Standup Summary",
		}
		
		for _, section := range coreSections {
			if !contains(content.Text, section) {
				t.Errorf("Prompt missing core section: %s", section)
			}
		}
	})
}

func TestHandleWeeklyPlanningPrompt(t *testing.T) {
	ctx := context.Background()
	session := &mcp.ServerSession{}
	
	// Test this week planning
	t.Run("ThisWeekPlanning", func(t *testing.T) {
		params := &mcp.GetPromptParams{
			Name: "weekly_planning",
			Arguments: map[string]string{
				"user_id":          "planner.user",
				"planning_horizon": "this_week",
			},
		}
		
		result, err := handleWeeklyPlanningPrompt(ctx, session, params)
		if err != nil {
			t.Fatalf("handleWeeklyPlanningPrompt failed: %v", err)
		}
		
		if result == nil {
			t.Fatal("handleWeeklyPlanningPrompt returned nil result")
		}
		
		content, ok := result.Messages[0].Content.(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}
		
		if !contains(content.Text, "planner.user") {
			t.Error("Prompt missing user ID")
		}
		if !contains(content.Text, "this_week") {
			t.Error("Prompt missing planning horizon")
		}
		if !contains(content.Text, "This Week's Strategic Focus") {
			t.Error("Prompt missing this week specific guidance")
		}
	})
	
	// Test next week planning
	t.Run("NextWeekPlanning", func(t *testing.T) {
		params := &mcp.GetPromptParams{
			Name: "weekly_planning",
			Arguments: map[string]string{
				"user_id":          "future.planner",
				"planning_horizon": "next_week",
			},
		}
		
		result, err := handleWeeklyPlanningPrompt(ctx, session, params)
		if err != nil {
			t.Fatalf("handleWeeklyPlanningPrompt failed: %v", err)
		}
		
		content, ok := result.Messages[0].Content.(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}
		
		if !contains(content.Text, "Next Week Preparation Focus") {
			t.Error("Prompt missing next week specific guidance")
		}
		if !contains(content.Text, "Prerequisites") {
			t.Error("Prompt missing prerequisites section")
		}
	})
	
	// Test upcoming planning
	t.Run("UpcomingPlanning", func(t *testing.T) {
		params := &mcp.GetPromptParams{
			Name: "weekly_planning",
			Arguments: map[string]string{
				"user_id":          "strategic.planner",
				"planning_horizon": "upcoming",
			},
		}
		
		result, err := handleWeeklyPlanningPrompt(ctx, session, params)
		if err != nil {
			t.Fatalf("handleWeeklyPlanningPrompt failed: %v", err)
		}
		
		content, ok := result.Messages[0].Content.(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}
		
		if !contains(content.Text, "Medium-term Strategic Planning") {
			t.Error("Prompt missing medium-term strategic planning")
		}
		if !contains(content.Text, "2-4 weeks") {
			t.Error("Prompt missing medium-term timeframe")
		}
	})
	
	// Test core planning sections
	t.Run("CorePlanningSections", func(t *testing.T) {
		params := &mcp.GetPromptParams{
			Name: "weekly_planning",
			Arguments: map[string]string{
				"user_id": "complete.planner",
			},
		}
		
		result, err := handleWeeklyPlanningPrompt(ctx, session, params)
		if err != nil {
			t.Fatalf("handleWeeklyPlanningPrompt failed: %v", err)
		}
		
		content, ok := result.Messages[0].Content.(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}
		
		coreSections := []string{
			"Previous Week Review",
			"Current Workload Analysis",
			"Capacity Planning",
			"Strategic Priority Setting",
			"Daily Planning and Distribution",
			"Risk Management and Contingency",
			"Success Metrics and Accountability",
			"Communication and Coordination",
			"Personal Development Integration",
			"Weekly Plan Summary",
		}
		
		for _, section := range coreSections {
			if !contains(content.Text, section) {
				t.Errorf("Prompt missing core section: %s", section)
			}
		}
	})
}

func TestHandleTaskHandoffPrompt(t *testing.T) {
	ctx := context.Background()
	session := &mcp.ServerSession{}
	
	params := &mcp.GetPromptParams{
		Name: "task_handoff",
		Arguments: map[string]string{
			"task_id":   "task-handoff-123",
			"from_user": "alice.developer",
			"to_user":   "bob.engineer",
		},
	}
	
	result, err := handleTaskHandoffPrompt(ctx, session, params)
	if err != nil {
		t.Fatalf("handleTaskHandoffPrompt failed: %v", err)
	}
	
	if result == nil {
		t.Fatal("handleTaskHandoffPrompt returned nil result")
	}
	
	content, ok := result.Messages[0].Content.(*mcp.TextContent)
	if !ok {
		t.Fatal("Expected TextContent")
	}
	
	// Check for handoff-specific content
	if !contains(content.Text, "task-handoff-123") {
		t.Error("Prompt missing task ID")
	}
	if !contains(content.Text, "alice.developer") {
		t.Error("Prompt missing from user")
	}
	if !contains(content.Text, "bob.engineer") {
		t.Error("Prompt missing to user")
	}
	
	// Check for all major handoff sections
	handoffSections := []string{
		"Task Context and Background",
		"Current State Assessment",
		"Knowledge Transfer",
		"Stakeholder and Communication Context",
		"Next Steps and Immediate Actions",
		"Resources and Documentation",
		"Support and Escalation",
		"Handoff Verification",
		"Historical Reference",
	}
	
	for _, section := range handoffSections {
		if !contains(content.Text, section) {
			t.Errorf("Prompt missing handoff section: %s", section)
		}
	}
	
	// Check for specific knowledge transfer subsections
	knowledgeTransferSections := []string{
		"Technical Knowledge",
		"Domain Knowledge",
		"Architecture and Approach",
		"Implementation Details",
		"Business Rules and Logic",
		"Data and Integration",
	}
	
	for _, section := range knowledgeTransferSections {
		if !contains(content.Text, section) {
			t.Errorf("Prompt missing knowledge transfer section: %s", section)
		}
	}
	
	// Check for handoff verification elements
	verificationElements := []string{
		"Knowledge Check",
		"Transition Plan",
		"Success Metrics",
		"Verification Questions",
	}
	
	for _, element := range verificationElements {
		if !contains(content.Text, element) {
			t.Errorf("Prompt missing verification element: %s", element)
		}
	}
}

func TestWorkflowPromptsArgumentValidation(t *testing.T) {
	ctx := context.Background()
	session := &mcp.ServerSession{}
	
	// Test that workflow prompts handle missing arguments gracefully
	prompts := CreateWorkflowPrompts()
	
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
		
		// Test with partial arguments
		t.Run(prompt.Prompt.Name+"_PartialArguments", func(t *testing.T) {
			// Create partial arguments based on the first required argument
			var partialArgs map[string]string
			if len(prompt.Prompt.Arguments) > 0 {
				firstArg := prompt.Prompt.Arguments[0]
				partialArgs = map[string]string{
					firstArg.Name: "test_value",
				}
			}
			
			params := &mcp.GetPromptParams{
				Name:      prompt.Prompt.Name,
				Arguments: partialArgs,
			}
			
			result, err := prompt.Handler(ctx, session, params)
			if err != nil {
				t.Fatalf("Prompt %s failed with partial arguments: %v", prompt.Prompt.Name, err)
			}
			
			if result == nil {
				t.Fatalf("Prompt %s returned nil result with partial arguments", prompt.Prompt.Name)
			}
			
			if len(result.Messages) == 0 {
				t.Fatalf("Prompt %s returned no messages with partial arguments", prompt.Prompt.Name)
			}
		})
	}
}

func TestWorkflowPromptDescriptionsAndArguments(t *testing.T) {
	prompts := CreateWorkflowPrompts()
	
	// Test that each prompt has meaningful description and proper arguments
	for _, prompt := range prompts {
		t.Run(prompt.Prompt.Name+"_DescriptionAndArgs", func(t *testing.T) {
			// Check description length and content
			if len(prompt.Prompt.Description) < 20 {
				t.Errorf("Prompt %s has too short description: %s", prompt.Prompt.Name, prompt.Prompt.Description)
			}
			
			if !containsIgnoreCase(prompt.Prompt.Description, "template") && !containsIgnoreCase(prompt.Prompt.Description, "guide") {
				t.Errorf("Prompt %s description should mention template or guide: %s", prompt.Prompt.Name, prompt.Prompt.Description)
			}
			
			// Check that required arguments exist
			hasRequiredArg := false
			for _, arg := range prompt.Prompt.Arguments {
				if arg.Required {
					hasRequiredArg = true
				}
				
				// Check argument descriptions
				if arg.Description == "" {
					t.Errorf("Prompt %s argument %s missing description", prompt.Prompt.Name, arg.Name)
				}
				
				if len(arg.Description) < 10 {
					t.Errorf("Prompt %s argument %s has too short description: %s", prompt.Prompt.Name, arg.Name, arg.Description)
				}
			}
			
			if !hasRequiredArg {
				t.Errorf("Prompt %s should have at least one required argument", prompt.Prompt.Name)
			}
		})
	}
}

// containsIgnoreCase checks if string contains substring (case insensitive)
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}