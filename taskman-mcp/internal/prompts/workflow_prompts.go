package prompts

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// CreateWorkflowPrompts returns all workflow-related prompt definitions
func CreateWorkflowPrompts() []*mcp.ServerPrompt {
	return []*mcp.ServerPrompt{
		{
			Prompt: &mcp.Prompt{
				Name:        "daily_standup",
				Description: "Generate daily work summaries and planning templates for team standups",
				Arguments: []*mcp.PromptArgument{
					{
						Name:        "user_id",
						Description: "User ID for personalized daily summary",
						Required:    true,
					},
					{
						Name:        "standup_type",
						Description: "Type of standup (individual, team, cross-team)",
						Required:    false,
					},
				},
			},
			Handler: handleDailyStandupPrompt,
		},
		{
			Prompt: &mcp.Prompt{
				Name:        "weekly_planning",
				Description: "Weekly priority and capacity planning template for effective work organization",
				Arguments: []*mcp.PromptArgument{
					{
						Name:        "user_id",
						Description: "User ID for personalized weekly planning",
						Required:    true,
					},
					{
						Name:        "planning_horizon",
						Description: "Planning timeframe (this_week, next_week, upcoming)",
						Required:    false,
					},
				},
			},
			Handler: handleWeeklyPlanningPrompt,
		},
		{
			Prompt: &mcp.Prompt{
				Name:        "task_handoff",
				Description: "Template for transferring tasks between team members with complete context",
				Arguments: []*mcp.PromptArgument{
					{
						Name:        "task_id",
						Description: "ID of the task being handed off",
						Required:    true,
					},
					{
						Name:        "from_user",
						Description: "Current task owner who is handing off",
						Required:    true,
					},
					{
						Name:        "to_user",
						Description: "New task owner receiving the handoff",
						Required:    true,
					},
				},
			},
			Handler: handleTaskHandoffPrompt,
		},
	}
}

// handleDailyStandupPrompt generates a daily work summary template
func handleDailyStandupPrompt(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.GetPromptParams,
) (*mcp.GetPromptResult, error) {
	slog.Info("Generating daily_standup prompt", "name", params.Name)

	// Extract arguments
	userID := ""
	standupType := "individual"

	if params.Arguments != nil {
		if user, ok := params.Arguments["user_id"]; ok {
			userID = user
		}
		if stype, ok := params.Arguments["standup_type"]; ok {
			standupType = stype
		}
	}

	promptText := fmt.Sprintf(`# Daily Standup Report - %s

## Standup Overview
**User:** %s
**Date:** [Today's Date]
**Standup Type:** %s

## Yesterday's Accomplishments

### Completed Tasks
**What did you finish yesterday?**
- [ ] [Task name] - [Brief description of what was completed]
- [ ] [Task name] - [Key outcome or deliverable]
- [ ] [Task name] - [Progress made or milestone reached]

### Key Achievements
**Significant progress or wins from yesterday:**
- [ ] [Major milestone reached or problem solved]
- [ ] [Important decision made or approval received]
- [ ] [Successful collaboration or meeting outcome]
- [ ] [Learning or insight that will help future work]

### Stakeholder Interactions
- [ ] **Meetings Attended:** [Key meetings and outcomes]
- [ ] **Communications Sent:** [Important updates or decisions communicated]
- [ ] **Feedback Received:** [Input from stakeholders or team members]

## Today's Plan

### Priority Tasks
**What are your top 3 priorities for today?**
1. **High Priority:** [Most critical task - what specifically will you accomplish?]
2. **Medium Priority:** [Important task - what progress will you make?]
3. **Low Priority:** [If time permits - what would you like to start/continue?]

### Scheduled Activities
- [ ] **Meetings:** [List key meetings and their purpose]
- [ ] **Deadlines:** [Any deliverables due today]
- [ ] **Collaborative Work:** [Work requiring coordination with others]
- [ ] **Deep Work:** [Focused individual work planned]

### Expected Outcomes
**By end of day, what will be accomplished?**
- [ ] [Specific deliverable or milestone to complete]
- [ ] [Progress measurement or completion percentage expected]
- [ ] [Decision to make or communication to send]

## Blockers and Challenges

### Current Blockers
**What is preventing you from making progress?**
- [ ] **Blocker:** [Specific obstacle]
  - **Impact:** [How this affects your work]
  - **Help Needed:** [What assistance would resolve this]
  - **Urgency:** [How quickly this needs resolution]

### Potential Challenges
**What might slow you down today?**
- [ ] [Dependency on someone else's work or decision]
- [ ] [Resource or access limitation]
- [ ] [Technical challenge or unknown complexity]
- [ ] [Competing priority or time constraint]

### Support Requests
**What help do you need from the team?**
- [ ] **From [Person/Team]:** [Specific request and timeline]
- [ ] **Technical Support:** [System access, tool help, expertise needed]
- [ ] **Decision Needed:** [What decision is blocking progress and who can make it]`, userID, userID, standupType)

	// Add standup-type specific guidance
	switch standupType {
	case "team":
		promptText += `

## Team Coordination

### Collaboration Opportunities
- [ ] **Can Help With:** [Areas where you can assist teammates]
- [ ] **Need Collaboration:** [Work that requires team coordination]
- [ ] **Knowledge Sharing:** [Insights or solutions to share]

### Team Dependencies
- [ ] **Waiting For:** [Work or decisions needed from teammates]
- [ ] **Providing To:** [Deliverables or support you're providing to others]
- [ ] **Coordination Needed:** [Activities requiring team synchronization]`

	case "cross-team":
		promptText += `

## Cross-Team Updates

### External Dependencies
- [ ] **Other Teams:** [Work dependencies on other teams]
- [ ] **External Stakeholders:** [Updates needed from outside the immediate team]
- [ ] **System Dependencies:** [Infrastructure or platform requirements]

### Communication Priorities
- [ ] **Updates to Share:** [Information other teams need to know]
- [ ] **Decisions Pending:** [Cross-team decisions affecting your work]
- [ ] **Escalations Needed:** [Issues requiring management attention]`

	default: // individual
		promptText += `

## Personal Planning

### Time Management
- [ ] **Deep Work Blocks:** [When will you focus on complex tasks?]
- [ ] **Communication Windows:** [When will you check messages/emails?]
- [ ] **Break Schedule:** [Planned breaks to maintain productivity]

### Learning and Development
- [ ] **Skill Building:** [Any learning opportunities in today's work?]
- [ ] **Process Improvement:** [Ways to work more effectively?]
- [ ] **Documentation:** [Knowledge to capture or share?]`
	}

	promptText += `

## Looking Ahead

### Tomorrow's Preparation
- [ ] **Setup for Tomorrow:** [What can you prepare today for tomorrow's work?]
- [ ] **Dependencies to Resolve:** [What needs completion to unblock tomorrow?]
- [ ] **Communications to Send:** [Updates or requests to send before EOD]

### Weekly Progress Check
- [ ] **Weekly Goals Status:** [How are you tracking against weekly objectives?]
- [ ] **Adjustments Needed:** [Any priority or timeline changes for the week?]
- [ ] **Support Required:** [Help needed to stay on track for weekly goals?]

## Standup Summary

### Key Message for Team
**In one sentence, what's your status and main focus?**
[Example: "Completed user authentication module yesterday, focusing on database integration today, blocked on API specification approval."]

### Energy and Capacity
- [ ] **Energy Level:** [High/Medium/Low - how are you feeling?]
- [ ] **Capacity:** [Normal/Reduced/Extra - what's your bandwidth today?]
- [ ] **Availability:** [Any schedule constraints or out-of-office time?]

### Success Metrics for Today
**How will you know today was successful?**
- [ ] [Specific accomplishment that would make today a win]
- [ ] [Progress measurement that indicates good momentum]
- [ ] [Problem solved or decision made that removes friction]

---
*Use this template to ensure comprehensive daily planning and effective team communication during standups.*`

	return &mcp.GetPromptResult{
		Description: "Daily standup preparation and work planning template",
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

// handleWeeklyPlanningPrompt generates a weekly planning template
func handleWeeklyPlanningPrompt(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.GetPromptParams,
) (*mcp.GetPromptResult, error) {
	slog.Info("Generating weekly_planning prompt", "name", params.Name)

	// Extract arguments
	userID := ""
	planningHorizon := "this_week"

	if params.Arguments != nil {
		if user, ok := params.Arguments["user_id"]; ok {
			userID = user
		}
		if horizon, ok := params.Arguments["planning_horizon"]; ok {
			planningHorizon = horizon
		}
	}

	promptText := fmt.Sprintf(`# Weekly Planning Session - %s

## Planning Overview
**User:** %s
**Planning Period:** %s
**Planning Date:** [Current Date]

## 1. Previous Week Review (if applicable)

### Accomplishments Assessment
**Major achievements from last week:**
- [ ] [Key project milestone or deliverable completed]
- [ ] [Important problem solved or decision made]
- [ ] [Successful collaboration or stakeholder engagement]
- [ ] [Process improvement or efficiency gain]

### Goals vs. Reality Check
- [ ] **Goals Met:** [Which planned objectives were accomplished?]
- [ ] **Goals Missed:** [What didn't get done and why?]
- [ ] **Unexpected Work:** [Unplanned tasks that consumed time]
- [ ] **Efficiency Insights:** [What worked well vs. what hindered progress?]

## 2. Current Workload Analysis

### Active Project Portfolio
**Current projects and their status:**
| Project | Priority | Status | This Week's Goal | Time Allocation |
|---------|----------|--------|------------------|-----------------|
| 1. | High/Med/Low | %% Complete | | X hours |
| 2. | High/Med/Low | %% Complete | | X hours |
| 3. | High/Med/Low | %% Complete | | X hours |

### Task Inventory
**All pending tasks by category:**

**High Priority (Must Do This Week):**
- [ ] [Critical task with deadline or blocker impact]
- [ ] [Important deliverable with stakeholder dependency]
- [ ] [Time-sensitive opportunity or commitment]

**Medium Priority (Should Do This Week):**
- [ ] [Important but flexible timing]
- [ ] [Preparation for next week's critical work]
- [ ] [Improvement or optimization opportunity]

**Low Priority (Nice to Do This Week):**
- [ ] [Enhancement or learning opportunity]
- [ ] [Documentation or cleanup task]
- [ ] [Exploration or research activity]

## 3. Capacity Planning

### Time Availability Assessment
**Total available hours this week:** [X hours]

**Fixed commitments:**
- [ ] **Meetings:** [X hours] - [List key recurring or scheduled meetings]
- [ ] **Deadlines:** [X hours] - [Time needed for specific deliverables]
- [ ] **Administrative:** [X hours] - [Email, reports, routine tasks]
- [ ] **Buffer/Unexpected:** [X hours] - [Reserve for urgent issues]

**Available for planned work:** [Remaining hours]

### Energy and Focus Planning
**When are you most productive?**
- [ ] **Deep Work Windows:** [Best times for complex, focused tasks]
- [ ] **Collaborative Work:** [Optimal times for meetings and teamwork]
- [ ] **Routine Tasks:** [When to handle administrative work]
- [ ] **Creative Work:** [Best times for innovation and problem-solving]

## 4. Strategic Priority Setting`, userID, userID, planningHorizon)

	// Add horizon-specific guidance
	switch planningHorizon {
	case "next_week":
		promptText += `

### Next Week Preparation Focus
**What must be set up for success next week?**
- [ ] **Prerequisites:** [What needs completion this week to enable next week?]
- [ ] **Resource Arrangement:** [Tools, access, or support to secure]
- [ ] **Stakeholder Alignment:** [Communications or decisions needed]
- [ ] **Dependency Management:** [Coordination with others required]`

	case "upcoming":
		promptText += `

### Medium-term Strategic Planning
**Looking ahead 2-4 weeks, what preparation is needed?**
- [ ] **Project Milestones:** [Major deliverables approaching]
- [ ] **Resource Planning:** [Skills, tools, or support to arrange]
- [ ] **Stakeholder Engagement:** [Key relationships to nurture]
- [ ] **Process Optimization:** [Workflows to improve for upcoming demands]`

	default: // this_week
		promptText += `

### This Week's Strategic Focus
**What will make this week highly impactful?**
- [ ] **Core Objectives:** [2-3 main goals that drive significant value]
- [ ] **Stakeholder Priorities:** [What matters most to key stakeholders?]
- [ ] **Project Momentum:** [What will keep important projects moving?]
- [ ] **Problem Resolution:** [Key blockers or issues to solve]`
	}

	promptText += `

### Impact vs. Effort Analysis
**High Impact, Low Effort (Quick Wins):**
- [ ] [Task that delivers significant value with minimal time investment]
- [ ] [Process improvement that saves time long-term]

**High Impact, High Effort (Strategic Work):**
- [ ] [Major deliverable or milestone requiring significant focus]
- [ ] [Complex problem requiring deep work and expertise]

**Low Impact, Low Effort (Fill-in Work):**
- [ ] [Administrative task for spare time]
- [ ] [Documentation or cleanup when energy is low]

**Low Impact, High Effort (Avoid/Defer):**
- [ ] [Work to delegate, defer, or eliminate if possible]

## 5. Daily Planning and Distribution

### Monday Planning
**Week kickoff priorities:**
- [ ] **Primary Focus:** [Most important task to start strong]
- [ ] **Team Coordination:** [Key meetings or collaborative work]
- [ ] **Weekly Setup:** [Administrative tasks to enable the week]

### Tuesday-Wednesday (Peak Productivity)
**Deep work and major deliverables:**
- [ ] **Tuesday Focus:** [Complex task requiring fresh energy]
- [ ] **Wednesday Focus:** [Continuation of complex work or major deliverable]

### Thursday Planning
**Collaboration and refinement:**
- [ ] **Team Work:** [Collaborative tasks or stakeholder engagement]
- [ ] **Quality Review:** [Review and refinement of week's work]

### Friday Planning
**Completion and preparation:**
- [ ] **Week Closure:** [Finishing deliverables and communications]
- [ ] **Next Week Setup:** [Preparation for following week]

## 6. Risk Management and Contingency

### Potential Challenges
**What could disrupt this week's plan?**
- [ ] **External Dependencies:** [Delays from others that could affect your work]
- [ ] **Technical Risks:** [System issues or complex problems that might emerge]
- [ ] **Competing Priorities:** [Urgent requests that might redirect focus]
- [ ] **Capacity Constraints:** [Energy or time limitations to consider]

### Mitigation Strategies
- [ ] **Plan B Tasks:** [Alternative work if blocked on primary tasks]
- [ ] **Early Warning Systems:** [How to detect problems early]
- [ ] **Communication Protocols:** [When and how to escalate issues]
- [ ] **Buffer Management:** [How to use reserved time effectively]

## 7. Success Metrics and Accountability

### Weekly Success Criteria
**How will you know this week was successful?**
- [ ] **Completion Metrics:** [Specific deliverables or milestones to achieve]
- [ ] **Progress Metrics:** [Percentage progress on key projects]
- [ ] **Quality Metrics:** [Standards for work quality or stakeholder satisfaction]
- [ ] **Learning Metrics:** [Skills developed or knowledge gained]

### Check-in Schedule
- [ ] **Monday:** [Quick plan review and adjustment]
- [ ] **Wednesday:** [Mid-week progress assessment]
- [ ] **Friday:** [Week completion review and next week preparation]

## 8. Communication and Coordination

### Stakeholder Updates
**Who needs to know your plans and progress?**
- [ ] **Manager/Team Lead:** [Key priorities and any support needed]
- [ ] **Team Members:** [Coordination points and dependencies]
- [ ] **Project Stakeholders:** [Deliverable timelines and progress updates]

### Proactive Communications
- [ ] **Status Updates:** [Regular progress reports to send]
- [ ] **Decision Requests:** [Approvals or input needed from others]
- [ ] **Coordination Messages:** [Scheduling or planning communications]

## 9. Personal Development Integration

### Learning Opportunities
**How will this week contribute to your growth?**
- [ ] **Skill Development:** [New capabilities to practice or develop]
- [ ] **Knowledge Building:** [Expertise to deepen or areas to explore]
- [ ] **Network Building:** [Relationships to develop or strengthen]

### Process Improvement
**What will you experiment with this week?**
- [ ] **Productivity Methods:** [New tools or techniques to try]
- [ ] **Communication Approaches:** [Ways to improve collaboration]
- [ ] **Time Management:** [Scheduling or focus techniques to test]

## 10. Weekly Plan Summary

### Top 3 Weekly Objectives
1. **Primary Goal:** [Most critical outcome for the week]
2. **Secondary Goal:** [Important supporting objective]
3. **Development Goal:** [Learning or improvement focus]

### Key Success Factors
- [ ] **Focus Areas:** [Where to concentrate energy for maximum impact]
- [ ] **Critical Dependencies:** [What must go right for success]
- [ ] **Resource Requirements:** [Support or tools needed]

### Commitment Statement
**One-sentence commitment for the week:**
[Example: "This week I will complete the user interface mockups, resolve the database performance issue, and coordinate the team review session."]

---
*This planning template ensures strategic weekly focus while maintaining flexibility for adaptation and continuous improvement.*`

	return &mcp.GetPromptResult{
		Description: "Comprehensive weekly planning and priority setting template",
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

// handleTaskHandoffPrompt generates a task transfer template
func handleTaskHandoffPrompt(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.GetPromptParams,
) (*mcp.GetPromptResult, error) {
	slog.Info("Generating task_handoff prompt", "name", params.Name)

	// Extract arguments
	taskID := ""
	fromUser := ""
	toUser := ""

	if params.Arguments != nil {
		if id, ok := params.Arguments["task_id"]; ok {
			taskID = id
		}
		if from, ok := params.Arguments["from_user"]; ok {
			fromUser = from
		}
		if to, ok := params.Arguments["to_user"]; ok {
			toUser = to
		}
	}

	promptText := fmt.Sprintf(`# Task Handoff Documentation

## Handoff Overview
**Task ID:** %s
**From:** %s
**To:** %s
**Handoff Date:** [Current Date]

## 1. Task Context and Background

### Task Summary
**Task Name:** [Current task name and brief description]
**Project Context:** [Which project or initiative does this belong to?]
**Original Objective:** [What was this task meant to accomplish?]
**Current Status:** [What's the current state of completion?]

### Business Context
**Why This Task Matters:**
- [ ] **Business Value:** [What problem does this solve or opportunity does it create?]
- [ ] **Stakeholder Impact:** [Who benefits from this work and how?]
- [ ] **Timeline Importance:** [Why does timing matter for this task?]
- [ ] **Dependencies:** [What other work depends on completing this task?]

### Historical Context
**Task Evolution:**
- [ ] **Original Scope:** [What was initially planned?]
- [ ] **Scope Changes:** [How has the scope evolved and why?]
- [ ] **Previous Challenges:** [What obstacles have been encountered?]
- [ ] **Decisions Made:** [Key decisions that shaped the current approach]

## 2. Current State Assessment

### Work Completed So Far
**Deliverables Created:**
- [ ] [Specific output or artifact 1 - location and status]
- [ ] [Specific output or artifact 2 - location and status]
- [ ] [Specific output or artifact 3 - location and status]

**Progress Made:**
- [ ] **Research/Analysis:** [What groundwork has been laid?]
- [ ] **Design/Planning:** [What approaches have been developed?]
- [ ] **Implementation:** [What has been built or created?]
- [ ] **Testing/Validation:** [What has been verified or validated?]

### Current Challenges and Blockers
**Active Issues:**
- [ ] **Technical Challenge:** [Specific technical problem and current approach]
- [ ] **Resource Constraint:** [Missing tools, access, or expertise]
- [ ] **Dependency Blocker:** [Waiting for external input or decision]
- [ ] **Scope Ambiguity:** [Areas where requirements need clarification]

### Quality and Standards
- [ ] **Code Quality:** [Current state of code quality and standards compliance]
- [ ] **Documentation Status:** [What documentation exists and what's missing]
- [ ] **Testing Coverage:** [What testing has been done and what's needed]
- [ ] **Review Status:** [What reviews have occurred and what's pending]

## 3. Knowledge Transfer

### Technical Knowledge
**Architecture and Approach:**
- [ ] **System Design:** [How does this fit into the larger system?]
- [ ] **Technical Approach:** [What methodology or framework is being used?]
- [ ] **Key Technologies:** [What tools, languages, or platforms are involved?]
- [ ] **Performance Considerations:** [Any scalability or efficiency requirements?]

**Implementation Details:**
- [ ] **Code Structure:** [How is the code organized and why?]
- [ ] **Key Functions/Modules:** [What are the main components and their purposes?]
- [ ] **Configuration:** [What settings or environment setup is required?]
- [ ] **Dependencies:** [What external libraries or services are used?]

### Domain Knowledge
**Business Rules and Logic:**
- [ ] **Core Requirements:** [What are the essential functional requirements?]
- [ ] **Business Rules:** [What logic or constraints must be followed?]
- [ ] **Edge Cases:** [What unusual scenarios need to be handled?]
- [ ] **User Experience:** [What user interactions or workflows are involved?]

**Data and Integration:**
- [ ] **Data Sources:** [Where does information come from?]
- [ ] **Data Transformations:** [How is data processed or modified?]
- [ ] **External Systems:** [What other systems does this interact with?]
- [ ] **Security Requirements:** [What security considerations apply?]

## 4. Stakeholder and Communication Context

### Key Stakeholders
**Primary Contacts:**
| Role | Name | Involvement | Contact Info | Preferences |
|------|------|-------------|--------------|-------------|
| Project Owner | | Final decisions, requirements | | |
| Technical Lead | | Architecture guidance | | |
| End User Rep | | Requirements validation | | |
| QA/Testing | | Quality assurance | | |

### Communication History
**Recent Important Conversations:**
- [ ] **Stakeholder Meeting:** [Date, participants, key decisions or feedback]
- [ ] **Technical Discussion:** [Key technical decisions or approaches agreed upon]
- [ ] **Status Update:** [Latest status communication and stakeholder response]

**Pending Communications:**
- [ ] **Status Report Due:** [When and to whom]
- [ ] **Decision Needed:** [What requires stakeholder input or approval]
- [ ] **Feedback Expected:** [What reviews or input are anticipated]

## 5. Next Steps and Immediate Actions

### Immediate Priorities (Next 1-2 Days)
1. [ ] **Critical First Step:** [Most urgent action to maintain momentum]
2. [ ] **Knowledge Verification:** [Confirm understanding of key concepts]
3. [ ] **Stakeholder Contact:** [Who to reach out to first and why]

### Short-term Goals (Next 1-2 Weeks)
**Primary Objectives:**
- [ ] [Next major milestone or deliverable to focus on]
- [ ] [Key problem to solve or decision to make]
- [ ] [Important stakeholder engagement or approval to secure]

**Success Criteria:**
- [ ] [How to know you're making good progress]
- [ ] [What constitutes completion of the next phase]
- [ ] [Key metrics or feedback to track]

### Recommended Approach
**Strategic Suggestions:**
- [ ] **Technical Strategy:** [Recommended technical approach based on experience]
- [ ] **Stakeholder Strategy:** [How to best work with key stakeholders]
- [ ] **Risk Mitigation:** [What to watch out for and how to handle issues]
- [ ] **Quality Approach:** [How to ensure work meets standards]

## 6. Resources and Documentation

### Essential Resources
**Documentation:**
- [ ] **Requirements:** [Location of current requirements documentation]
- [ ] **Technical Specs:** [Architecture or design documentation]
- [ ] **Previous Work:** [Related projects or reference implementations]
- [ ] **Standards:** [Coding standards, design guidelines, or compliance requirements]

**Tools and Access:**
- [ ] **Development Environment:** [How to set up and access development tools]
- [ ] **Testing Environment:** [Access to testing systems or data]
- [ ] **Communication Tools:** [Team chat channels, project boards, etc.]
- [ ] **Repository Access:** [Code repositories, documentation wikis, etc.]

### Learning Resources
**Getting Up to Speed:**
- [ ] **Domain Knowledge:** [Best sources for understanding the problem space]
- [ ] **Technical Knowledge:** [Documentation or tutorials for key technologies]
- [ ] **Process Knowledge:** [How this team or organization typically works]

## 7. Support and Escalation

### Available Support
**Technical Support:**
- [ ] **Primary Technical Contact:** [Who to ask for technical guidance]
- [ ] **Secondary Support:** [Alternative technical resources]
- [ ] **Subject Matter Expert:** [Domain expert for business questions]

**Process Support:**
- [ ] **Project Management:** [Who handles project coordination and planning]
- [ ] **Quality Assurance:** [Who provides testing or review support]
- [ ] **DevOps/Infrastructure:** [Who handles deployment and operational issues]

### Escalation Paths
**When to Escalate:**
- [ ] **Technical Blockers:** [When technical challenges exceed your expertise]
- [ ] **Scope Questions:** [When requirements or priorities are unclear]
- [ ] **Resource Needs:** [When additional support or tools are needed]
- [ ] **Timeline Concerns:** [When deadlines appear at risk]

## 8. Handoff Verification

### Knowledge Check
**Verification Questions:**
- [ ] Do you understand the business context and importance of this task?
- [ ] Are you clear on the current technical approach and architecture?
- [ ] Do you know who to contact for different types of questions or support?
- [ ] Can you access all necessary tools, systems, and documentation?
- [ ] Do you understand the immediate next steps and priorities?

### Transition Plan
**Handoff Activities:**
- [ ] **Code Walkthrough:** [Schedule technical walkthrough session]
- [ ] **Stakeholder Introduction:** [Introduce new owner to key stakeholders]
- [ ] **Documentation Review:** [Review all relevant documentation together]
- [ ] **Environment Setup:** [Ensure new owner can access all necessary tools]

### Success Metrics
**Handoff Completion Criteria:**
- [ ] New owner can independently access all necessary resources
- [ ] Key stakeholders have been introduced and are comfortable with transition
- [ ] Next immediate steps are clear and actionable
- [ ] Support structure is established and understood

## 9. Historical Reference

### Previous Owner's Insights
**Lessons Learned:**
- [ ] **What Worked Well:** [Successful approaches or strategies]
- [ ] **What Was Challenging:** [Difficult aspects and how they were handled]
- [ ] **Stakeholder Insights:** [Key learnings about working with specific people]
- [ ] **Technical Insights:** [Gotchas, shortcuts, or important technical details]

### Recommendations for Success
**Top Advice from Previous Owner:**
1. [Most important insight for success with this task]
2. [Key relationship or communication strategy]
3. [Technical approach or tool that's particularly effective]

---
*This handoff documentation ensures complete knowledge transfer and sets up the new task owner for success while preserving institutional knowledge.*`, taskID, fromUser, toUser)

	return &mcp.GetPromptResult{
		Description: "Comprehensive task handoff documentation and knowledge transfer template",
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