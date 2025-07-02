package prompts

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// CreateTaskPrompts returns all task-related prompt definitions
func CreateTaskPrompts() []*mcp.ServerPrompt {
	return []*mcp.ServerPrompt{
		{
			Prompt: &mcp.Prompt{
				Name:        "plan_task",
				Description: "Guide comprehensive task planning with context gathering and requirement analysis",
				Arguments: []*mcp.PromptArgument{
					{
						Name:        "task_name",
						Description: "Name or brief description of the task to plan",
						Required:    true,
					},
					{
						Name:        "project_context",
						Description: "Optional project context or ID for this task",
						Required:    false,
					},
					{
						Name:        "complexity",
						Description: "Estimated complexity level (Simple, Medium, Complex)",
						Required:    false,
					},
				},
			},
			Handler: handlePlanTaskPrompt,
		},
		{
			Prompt: &mcp.Prompt{
				Name:        "update_task_status",
				Description: "Template for updating task status with proper documentation and next steps",
				Arguments: []*mcp.PromptArgument{
					{
						Name:        "task_id",
						Description: "ID of the task being updated",
						Required:    true,
					},
					{
						Name:        "current_status",
						Description: "Current status of the task",
						Required:    true,
					},
					{
						Name:        "new_status",
						Description: "Desired new status for the task",
						Required:    true,
					},
				},
			},
			Handler: handleUpdateTaskStatusPrompt,
		},
		{
			Prompt: &mcp.Prompt{
				Name:        "task_review",
				Description: "Template for comprehensive task completion review and lessons learned",
				Arguments: []*mcp.PromptArgument{
					{
						Name:        "task_id",
						Description: "ID of the completed task to review",
						Required:    true,
					},
					{
						Name:        "completion_date",
						Description: "Date when the task was completed",
						Required:    false,
					},
				},
			},
			Handler: handleTaskReviewPrompt,
		},
		{
			Prompt: &mcp.Prompt{
				Name:        "task_breakdown",
				Description: "Break down complex tasks into manageable subtasks with dependencies",
				Arguments: []*mcp.PromptArgument{
					{
						Name:        "parent_task",
						Description: "Description of the complex task to break down",
						Required:    true,
					},
					{
						Name:        "timeline",
						Description: "Optional timeline or deadline for the parent task",
						Required:    false,
					},
					{
						Name:        "team_size",
						Description: "Number of people who will work on this task",
						Required:    false,
					},
				},
			},
			Handler: handleTaskBreakdownPrompt,
		},
	}
}

// handlePlanTaskPrompt generates a comprehensive task planning guide
func handlePlanTaskPrompt(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.GetPromptParams,
) (*mcp.GetPromptResult, error) {
	slog.Info("Generating plan_task prompt", "name", params.Name)

	// Extract arguments
	taskName := ""
	projectContext := ""
	complexity := "Medium"

	if params.Arguments != nil {
		if name, ok := params.Arguments["task_name"]; ok {
			taskName = name
		}
		if project, ok := params.Arguments["project_context"]; ok {
			projectContext = project
		}
		if comp, ok := params.Arguments["complexity"]; ok {
			complexity = comp
		}
	}

	// Generate comprehensive planning prompt
	promptText := fmt.Sprintf(`# Task Planning Guide: %s

## Task Overview
**Task Name:** %s
**Complexity Level:** %s`, taskName, taskName, complexity)

	if projectContext != "" {
		promptText += fmt.Sprintf(`
**Project Context:** %s`, projectContext)
	}

	promptText += `

## Planning Checklist

### 1. Requirements Analysis
- [ ] What is the specific objective or outcome?
- [ ] What are the acceptance criteria for completion?
- [ ] Who are the stakeholders or beneficiaries?
- [ ] What constraints or limitations exist?

### 2. Resource Assessment
- [ ] What skills or expertise are required?
- [ ] What tools, systems, or access is needed?
- [ ] What is the estimated time investment?
- [ ] Are there any budget considerations?

### 3. Dependencies & Prerequisites
- [ ] What other tasks must be completed first?
- [ ] Who else needs to be involved or consulted?
- [ ] What information or decisions are needed before starting?
- [ ] Are there any external dependencies or approvals required?

### 4. Risk Assessment
- [ ] What could go wrong or cause delays?
- [ ] What are the backup plans or alternatives?
- [ ] How will progress be measured and communicated?
- [ ] What would warrant escalation or scope changes?

### 5. Implementation Strategy`

	switch complexity {
	case "Simple":
		promptText += `
- [ ] Define the single main deliverable
- [ ] Set a realistic completion timeline
- [ ] Identify the primary assignee
- [ ] Plan one checkpoint for progress review`

	case "Complex":
		promptText += `
- [ ] Break down into 3-5 major phases or milestones
- [ ] Identify critical path dependencies
- [ ] Plan regular review points and stakeholder updates
- [ ] Consider pilot testing or proof-of-concept approach
- [ ] Define clear handoff procedures between phases`

	default: // Medium
		promptText += `
- [ ] Break down into 2-3 manageable work packages
- [ ] Set intermediate milestones for progress tracking
- [ ] Plan mid-point review for course correction
- [ ] Define clear deliverables for each phase`
	}

	promptText += `

### 6. Next Steps
Based on your planning analysis above:

1. **Priority Level:** [High/Medium/Low] - Why?
2. **Estimated Timeline:** [Duration] - What factors influenced this estimate?
3. **Assignee Recommendation:** [Person/Team] - What skills make them suitable?
4. **First Action Item:** What specific step should be taken immediately?
5. **Success Metrics:** How will you know this task is truly complete?

## Task Creation Details
Use this analysis to create your task with:
- Clear, specific task name and description
- Appropriate priority level based on urgency and impact
- Realistic due date considering dependencies
- Detailed acceptance criteria
- Initial planning note with key insights from this analysis`

	return &mcp.GetPromptResult{
		Description: "Comprehensive task planning guidance",
		Messages: []*mcp.PromptMessage{
			{
				Role: "user",
				Content: &mcp.TextContent{
					Text: promptText,
				},
			},
		},
	}, nil
}

// handleUpdateTaskStatusPrompt generates a status update template
func handleUpdateTaskStatusPrompt(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.GetPromptParams,
) (*mcp.GetPromptResult, error) {
	slog.Info("Generating update_task_status prompt", "name", params.Name)

	// Extract arguments
	taskID := ""
	currentStatus := ""
	newStatus := ""

	if params.Arguments != nil {
		if id, ok := params.Arguments["task_id"]; ok {
			taskID = id
		}
		if current, ok := params.Arguments["current_status"]; ok {
			currentStatus = current
		}
		if newStat, ok := params.Arguments["new_status"]; ok {
			newStatus = newStat
		}
	}

	promptText := fmt.Sprintf(`# Task Status Update: %s

## Status Transition
**From:** %s → **To:** %s

## Update Documentation

### 1. Progress Summary
**What was accomplished since the last update?**
- [ ] List specific deliverables or milestones completed
- [ ] Quantify progress made (e.g., "75%% of research completed")
- [ ] Note any unexpected discoveries or insights

### 2. Current State Assessment`, taskID, currentStatus, newStatus)

	// Add status-specific guidance
	switch newStatus {
	case "In Progress":
		promptText += `
**Starting work - Key considerations:**
- [ ] Are all prerequisites and dependencies satisfied?
- [ ] Do you have the necessary resources and access?
- [ ] Is the approach and methodology clear?
- [ ] Have you communicated start date to stakeholders?

**Next Steps:**
- [ ] Define immediate next action (next 1-2 days)
- [ ] Set first checkpoint or milestone date
- [ ] Identify any early blockers or questions`

	case "Blocked":
		promptText += `
**Identifying blockers - Critical information:**
- [ ] What specific obstacle is preventing progress?
- [ ] Who or what is needed to resolve the blocker?
- [ ] Is this a dependency, resource, or decision blocker?
- [ ] What is the estimated timeline for resolution?

**Escalation Plan:**
- [ ] Who needs to be notified about this blocker?
- [ ] Is there alternative work that can proceed in parallel?
- [ ] Should the timeline or scope be reconsidered?`

	case "Review":
		promptText += `
**Ready for review - Completion checklist:**
- [ ] All deliverables completed according to acceptance criteria
- [ ] Quality check performed (testing, proofreading, validation)
- [ ] Documentation updated and complete
- [ ] Stakeholders notified that review is ready

**Review Requirements:**
- [ ] Who needs to review this work?
- [ ] What format should the review take?
- [ ] What are the review criteria or checklist?
- [ ] What is the expected timeline for feedback?`

	case "Complete":
		promptText += `
**Task completion - Final verification:**
- [ ] All acceptance criteria have been met
- [ ] Stakeholders have accepted the deliverables
- [ ] Documentation is complete and stored properly
- [ ] Any required handoffs or transitions completed

**Closure Activities:**
- [ ] Update any related tasks or dependencies
- [ ] Archive or organize work materials
- [ ] Capture lessons learned for future reference
- [ ] Communicate completion to relevant parties`

	default:
		promptText += `
**General status update considerations:**
- [ ] What work has been completed?
- [ ] What challenges or obstacles were encountered?
- [ ] Are there any changes to scope, timeline, or approach?
- [ ] What support or resources are needed going forward?`
	}

	promptText += `

### 3. Communication Plan
- [ ] **Who needs to be updated?** (stakeholders, team members, manager)
- [ ] **What format?** (email, meeting, dashboard update, chat)
- [ ] **When?** (immediate, daily standup, weekly report)
- [ ] **Key message:** What's the most important information to convey?

### 4. Next Actions
- [ ] **Immediate next step:** What specific action will be taken first?
- [ ] **Timeline:** When will the next update or milestone occur?
- [ ] **Dependencies:** What needs to happen before progress can continue?
- [ ] **Success criteria:** How will you know the next phase is complete?

## Note Template
Use this structure for your task update note:

**Progress Made:**
[Summarize key accomplishments]

**Current Status:**
[Explain the status change and reasoning]

**Next Steps:**
[List immediate actions and timeline]

**Notes/Issues:**
[Any concerns, blockers, or important observations]`

	return &mcp.GetPromptResult{
		Description: "Task status update guidance and documentation template",
		Messages: []*mcp.PromptMessage{
			{
				Role: "user",
				Content: &mcp.TextContent{
					Text: promptText,
				},
			},
		},
	}, nil
}

// handleTaskReviewPrompt generates a completion review template
func handleTaskReviewPrompt(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.GetPromptParams,
) (*mcp.GetPromptResult, error) {
	slog.Info("Generating task_review prompt", "name", params.Name)

	// Extract arguments
	taskID := ""
	completionDate := ""

	if params.Arguments != nil {
		if id, ok := params.Arguments["task_id"]; ok {
			taskID = id
		}
		if date, ok := params.Arguments["completion_date"]; ok {
			completionDate = date
		}
	}

	promptText := fmt.Sprintf(`# Task Completion Review: %s

## Review Overview
**Task ID:** %s`, taskID, taskID)

	if completionDate != "" {
		promptText += fmt.Sprintf(`
**Completion Date:** %s`, completionDate)
	}

	promptText += `

## Completion Assessment

### 1. Deliverable Quality Review
- [ ] **Acceptance Criteria Met:** Were all original requirements satisfied?
- [ ] **Quality Standards:** Does the work meet or exceed expected quality?
- [ ] **Stakeholder Satisfaction:** Have stakeholders accepted the deliverables?
- [ ] **Documentation Complete:** Is all necessary documentation in place?

### 2. Process Effectiveness
- [ ] **Timeline Performance:** Was the task completed on schedule?
- [ ] **Resource Utilization:** Were resources used efficiently?
- [ ] **Communication:** Was stakeholder communication effective throughout?
- [ ] **Problem Resolution:** Were issues and blockers handled appropriately?

### 3. Learning and Development
- [ ] **New Skills Acquired:** What knowledge or capabilities were developed?
- [ ] **Challenges Overcome:** What obstacles were successfully navigated?
- [ ] **Tools and Methods:** What techniques proved most effective?
- [ ] **Areas for Growth:** What could be improved for future tasks?

## Lessons Learned

### What Went Well
**Successes to replicate in future tasks:**
- [ ] Planning approach or methodology
- [ ] Communication strategies
- [ ] Tool usage or technical approach
- [ ] Team collaboration methods
- [ ] Problem-solving techniques

### What Could Be Improved
**Areas for enhancement next time:**
- [ ] Planning accuracy and thoroughness
- [ ] Resource estimation and allocation
- [ ] Risk anticipation and mitigation
- [ ] Stakeholder engagement
- [ ] Technical execution

### Unexpected Discoveries
**Insights gained during execution:**
- [ ] New understanding about the problem space
- [ ] Improved processes or workflows discovered
- [ ] Better tools or methods identified
- [ ] Relationships or dependencies uncovered

## Knowledge Transfer

### Documentation and Artifacts
- [ ] **Code/Files:** Where are deliverables stored and how to access them?
- [ ] **Documentation:** What guides or references were created?
- [ ] **Processes:** What new procedures or workflows resulted?
- [ ] **Contacts:** Who were key stakeholders or experts consulted?

### Recommendations for Future Work
- [ ] **Follow-up Tasks:** What additional work should be considered?
- [ ] **Process Improvements:** How can similar tasks be executed better?
- [ ] **Tool Recommendations:** What resources proved most valuable?
- [ ] **Team Insights:** What team dynamics or structures worked well?

## Impact Assessment
- [ ] **Business Value:** What concrete benefits were delivered?
- [ ] **User Impact:** How did this work affect end users or customers?
- [ ] **Team Learning:** What organizational knowledge was gained?
- [ ] **Strategic Alignment:** How did this contribute to larger goals?

## Final Review Note
**Summary for future reference:**

**Outcome:** [Brief description of what was delivered]

**Key Success Factors:** [Top 2-3 things that made this task successful]

**Lessons for Next Time:** [Most important insights for future similar work]

**Recommendations:** [Specific suggestions for follow-up or improvement]

---
*This review helps capture institutional knowledge and improves future task execution.*`

	return &mcp.GetPromptResult{
		Description: "Comprehensive task completion review and lessons learned template",
		Messages: []*mcp.PromptMessage{
			{
				Role: "user",
				Content: &mcp.TextContent{
					Text: promptText,
				},
			},
		},
	}, nil
}

// handleTaskBreakdownPrompt generates a task decomposition guide
func handleTaskBreakdownPrompt(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.GetPromptParams,
) (*mcp.GetPromptResult, error) {
	slog.Info("Generating task_breakdown prompt", "name", params.Name)

	// Extract arguments
	parentTask := ""
	timeline := ""
	teamSize := "1"

	if params.Arguments != nil {
		if task, ok := params.Arguments["parent_task"]; ok {
			parentTask = task
		}
		if time, ok := params.Arguments["timeline"]; ok {
			timeline = time
		}
		if size, ok := params.Arguments["team_size"]; ok {
			teamSize = size
		}
	}

	promptText := fmt.Sprintf(`# Task Breakdown Analysis: %s

## Parent Task Overview
**Complex Task:** %s`, parentTask, parentTask)

	if timeline != "" {
		promptText += fmt.Sprintf(`
**Timeline:** %s`, timeline)
	}

	promptText += fmt.Sprintf(`
**Team Size:** %s people

## Decomposition Strategy

### 1. Task Analysis Framework
**Break down the complex task by asking:**
- [ ] **What are the major phases or stages?** (Usually 3-5 main phases)
- [ ] **What are the key deliverables?** (Concrete outputs at each stage)
- [ ] **What expertise is required?** (Different skills for different phases)
- [ ] **What are the natural handoff points?** (Where work can be transferred)

### 2. Subtask Identification Matrix

**For each potential subtask, consider:**

| Subtask | Deliverable | Skills Needed | Dependencies | Effort | Owner |
|---------|-------------|---------------|--------------|--------|-------|
| 1. | | | | | |
| 2. | | | | | |
| 3. | | | | | |
| 4. | | | | | |
| 5. | | | | | |

### 3. Dependency Mapping
**Critical Path Analysis:**
- [ ] Which subtasks must be completed sequentially?
- [ ] Which subtasks can be worked on in parallel?
- [ ] Where are the bottlenecks or single points of failure?
- [ ] What external dependencies exist outside the team?

**Dependency Relationships:**
- **Sequential:** Task B cannot start until Task A is complete
- **Parallel:** Tasks can be worked on simultaneously
- [ ] **Prerequisites:** Tasks that provide inputs to other tasks
- **Optional:** Tasks that enhance but are not critical to completion

### 4. Resource Allocation Strategy`, teamSize)

	// Adjust guidance based on team size
	switch teamSize {
	case "1":
		promptText += `
**Single Person Approach:**
- [ ] Focus on clear sequential phases to avoid context switching
- [ ] Create natural stopping points for progress evaluation
- [ ] Plan work packages that can be completed in 1-3 day chunks
- [ ] Build in buffer time for learning and problem-solving`

	default:
		promptText += `
**Team Collaboration Strategy:**
- [ ] Assign subtasks based on individual expertise and interests
- [ ] Ensure clear interfaces between different people's work
- [ ] Plan regular integration points to combine individual efforts
- [ ] Define communication protocols for dependencies and blockers
- [ ] Consider pairing for knowledge transfer and quality assurance`
	}

	promptText += `

### 5. Subtask Definition Template

**For each identified subtask:**

**Subtask Name:** [Clear, action-oriented name]

**Objective:** [What specific outcome should be achieved]

**Deliverables:**
- [ ] [Concrete output 1]
- [ ] [Concrete output 2]

**Acceptance Criteria:**
- [ ] [How to know when this subtask is complete]
- [ ] [Quality standards or measurements]

**Dependencies:**
- **Requires:** [What must be done first]
- **Provides to:** [What other subtasks need this work]

**Effort Estimate:** [Time investment required]

**Skills Required:** [Expertise or knowledge needed]

**Potential Risks:** [What could go wrong]

### 6. Integration Planning
- [ ] **Integration Points:** When and how will subtask outputs be combined?
- [ ] **Quality Gates:** What reviews or checkpoints ensure standards?
- [ ] **Communication Plan:** How will the team coordinate and share progress?
- [ ] **Change Management:** How will scope changes be handled?

### 7. Recommended Subtask Structure

**Phase 1: Foundation/Setup**
- Research and requirements gathering
- Environment setup and tool preparation
- Stakeholder alignment and planning finalization

**Phase 2: Core Development/Execution**
- Primary work packages (the main effort)
- Regular progress reviews and adjustments
- Quality assurance and testing

**Phase 3: Integration/Completion**
- Component integration and system testing
- Documentation and knowledge transfer
- Stakeholder review and acceptance

## Next Steps Checklist
- [ ] Create individual tasks for each identified subtask
- [ ] Set up dependency relationships in the task management system
- [ ] Assign owners and estimated timelines for each subtask
- [ ] Schedule integration points and team check-ins
- [ ] Document the overall breakdown approach for future reference

---
*This breakdown ensures complex work is manageable while maintaining visibility into progress and dependencies.*`

	return &mcp.GetPromptResult{
		Description: "Comprehensive guide for breaking down complex tasks into manageable subtasks",
		Messages: []*mcp.PromptMessage{
			{
				Role: "user",
				Content: &mcp.TextContent{
					Text: promptText,
				},
			},
		},
	}, nil
}