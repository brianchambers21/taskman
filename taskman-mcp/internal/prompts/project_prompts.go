package prompts

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// CreateProjectPrompts returns all project-related prompt definitions
func CreateProjectPrompts() []*mcp.ServerPrompt {
	return []*mcp.ServerPrompt{
		{
			Prompt: &mcp.Prompt{
				Name:        "create_project_plan",
				Description: "Guide comprehensive project planning with goals, scope, and initial task structure",
				Arguments: []*mcp.PromptArgument{
					{
						Name:        "project_name",
						Description: "Name of the project to plan",
						Required:    true,
					},
					{
						Name:        "project_type",
						Description: "Type or category of project (Development, Research, Process, etc.)",
						Required:    false,
					},
					{
						Name:        "duration",
						Description: "Expected project duration or deadline",
						Required:    false,
					},
				},
			},
			Handler: handleCreateProjectPlanPrompt,
		},
		{
			Prompt: &mcp.Prompt{
				Name:        "project_status_review",
				Description: "Template for regular project health checks and status assessments",
				Arguments: []*mcp.PromptArgument{
					{
						Name:        "project_id",
						Description: "ID of the project to review",
						Required:    true,
					},
					{
						Name:        "review_period",
						Description: "Time period for this review (weekly, monthly, milestone)",
						Required:    false,
					},
				},
			},
			Handler: handleProjectStatusReviewPrompt,
		},
		{
			Prompt: &mcp.Prompt{
				Name:        "project_retrospective",
				Description: "Comprehensive post-project analysis template for lessons learned and improvement",
				Arguments: []*mcp.PromptArgument{
					{
						Name:        "project_id",
						Description: "ID of the completed project to analyze",
						Required:    true,
					},
					{
						Name:        "project_outcome",
						Description: "Overall outcome (Success, Partial Success, Challenge)",
						Required:    false,
					},
				},
			},
			Handler: handleProjectRetrospectivePrompt,
		},
	}
}

// handleCreateProjectPlanPrompt generates a comprehensive project planning guide
func handleCreateProjectPlanPrompt(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.GetPromptParams,
) (*mcp.GetPromptResult, error) {
	slog.Info("Generating create_project_plan prompt", "name", params.Name)

	// Extract arguments
	projectName := ""
	projectType := ""
	duration := ""

	if params.Arguments != nil {
		if name, ok := params.Arguments["project_name"]; ok {
			projectName = name
		}
		if ptype, ok := params.Arguments["project_type"]; ok {
			projectType = ptype
		}
		if dur, ok := params.Arguments["duration"]; ok {
			duration = dur
		}
	}

	promptText := fmt.Sprintf(`# Project Planning Guide: %s

## Project Overview
**Project Name:** %s`, projectName, projectName)

	if projectType != "" {
		promptText += fmt.Sprintf(`
**Project Type:** %s`, projectType)
	}

	if duration != "" {
		promptText += fmt.Sprintf(`
**Target Duration:** %s`, duration)
	}

	promptText += `

## 1. Project Definition & Scope

### Vision and Objectives
- [ ] **Project Vision:** What is the ultimate goal or desired outcome?
- [ ] **Success Criteria:** How will you know the project is successful?
- [ ] **Key Performance Indicators:** What metrics will track progress?
- [ ] **Business Justification:** Why is this project important now?

### Scope Definition
**In Scope (What will be delivered):**
- [ ] Core feature or deliverable #1
- [ ] Core feature or deliverable #2
- [ ] Core feature or deliverable #3

**Out of Scope (What will NOT be included):**
- [ ] Related but separate initiative #1
- [ ] Future enhancement #1
- [ ] Alternative approach that was considered but rejected

**Assumptions:**
- [ ] What conditions or resources are assumed to be available?
- [ ] What external factors are assumed to remain stable?
- [ ] What stakeholder behaviors or decisions are assumed?

## 2. Stakeholder Analysis

### Primary Stakeholders
| Role | Name/Group | Responsibilities | Communication Needs |
|------|------------|------------------|-------------------|
| Project Sponsor | | Funding, strategic direction | Monthly executive updates |
| Project Owner | | Day-to-day decisions, requirements | Weekly progress reports |
| End Users | | Requirements input, testing | Demo sessions, feedback loops |
| Technical Team | | Implementation, architecture | Daily standups, technical specs |

### Communication Plan
- [ ] **Kickoff Meeting:** When and who should attend?
- [ ] **Regular Updates:** What frequency and format?
- [ ] **Decision Points:** How will key decisions be made and communicated?
- [ ] **Issue Escalation:** What's the escalation path for problems?

## 3. Risk Assessment & Mitigation

### High-Level Risk Analysis
**Technical Risks:**
- [ ] **Risk:** [Describe potential technical challenge]
  - **Impact:** [High/Medium/Low]
  - **Probability:** [High/Medium/Low] 
  - **Mitigation:** [How to reduce or manage this risk]

**Resource Risks:**
- [ ] **Risk:** [Describe potential resource constraint]
  - **Impact:** [High/Medium/Low]
  - **Probability:** [High/Medium/Low]
  - **Mitigation:** [How to reduce or manage this risk]

**Timeline Risks:**
- [ ] **Risk:** [Describe potential schedule challenge]
  - **Impact:** [High/Medium/Low]
  - **Probability:** [High/Medium/Low]
  - **Mitigation:** [How to reduce or manage this risk]

## 4. Resource Planning

### Team Structure
- [ ] **Project Lead/Manager:** [Role responsibilities]
- [ ] **Technical Lead:** [If different from project lead]
- [ ] **Core Team Members:** [List key roles and skills needed]
- [ ] **Subject Matter Experts:** [Specialized knowledge required]
- [ ] **External Dependencies:** [Vendors, other teams, approvals needed]

### Budget Considerations
- [ ] **Personnel Costs:** Team time investment
- [ ] **Technology Costs:** Tools, licenses, infrastructure
- [ ] **External Costs:** Consultants, vendors, services
- [ ] **Contingency:** Buffer for unexpected expenses

## 5. High-Level Timeline & Phases

### Project Phases`

	// Add phase guidance based on project type
	switch projectType {
	case "Development", "Software":
		promptText += `
**Phase 1: Planning & Design** (Week 1-2)
- Requirements gathering and analysis
- Technical architecture and design
- Detailed project planning and task breakdown

**Phase 2: Core Development** (Week 3-6)
- Core feature implementation
- Regular testing and integration
- Stakeholder demos and feedback

**Phase 3: Testing & Refinement** (Week 7-8)
- Comprehensive testing and bug fixes
- User acceptance testing
- Performance optimization

**Phase 4: Deployment & Closure** (Week 9-10)
- Production deployment
- Documentation and training
- Project retrospective and closure`

	case "Research":
		promptText += `
**Phase 1: Research Design** (Week 1-2)
- Literature review and background research
- Methodology design and validation
- Data collection plan and resource setup

**Phase 2: Data Collection** (Week 3-6)
- Primary data gathering
- Secondary source compilation
- Regular progress assessment and course correction

**Phase 3: Analysis & Synthesis** (Week 7-8)
- Data analysis and pattern identification
- Findings synthesis and validation
- Preliminary conclusions and peer review

**Phase 4: Reporting & Presentation** (Week 9-10)
- Final report preparation
- Stakeholder presentations
- Recommendations and next steps`

	default:
		promptText += `
**Phase 1: Planning & Preparation** (20% of timeline)
- Detailed planning and resource allocation
- Stakeholder alignment and approval
- Setup and preparation activities

**Phase 2: Core Execution** (60% of timeline)
- Primary work and deliverable creation
- Regular progress reviews and adjustments
- Stakeholder communication and feedback

**Phase 3: Completion & Closure** (20% of timeline)
- Final deliverable preparation
- Quality assurance and stakeholder acceptance
- Project closure and lessons learned`
	}

	promptText += `

### Key Milestones
- [ ] **Milestone 1:** [Phase 1 completion criteria and date]
- [ ] **Milestone 2:** [Phase 2 completion criteria and date]
- [ ] **Milestone 3:** [Phase 3 completion criteria and date]
- [ ] **Final Delivery:** [Project completion criteria and date]

## 6. Initial Task Breakdown

### Immediate Setup Tasks (Week 1)
1. **Project Setup**
   - Create project in task management system
   - Set up communication channels and documentation space
   - Schedule kickoff meeting with key stakeholders

2. **Requirements Gathering**
   - Conduct stakeholder interviews
   - Document functional and non-functional requirements
   - Create acceptance criteria for major deliverables

3. **Planning Deep Dive**
   - Break down phases into detailed tasks
   - Estimate effort and assign initial responsibilities
   - Identify and document all dependencies

### Critical Path Identification
- [ ] What sequence of tasks determines the minimum project duration?
- [ ] Where are the bottlenecks or single points of failure?
- [ ] What tasks have the most uncertainty or risk?
- [ ] Which deliverables are prerequisites for multiple other tasks?

## 7. Success Metrics & Monitoring

### Tracking Mechanisms
- [ ] **Progress Metrics:** How will completion be measured?
- [ ] **Quality Metrics:** How will deliverable quality be assessed?
- [ ] **Resource Metrics:** How will effort and budget be tracked?
- [ ] **Stakeholder Satisfaction:** How will stakeholder approval be measured?

### Review Schedule
- [ ] **Daily:** Team standups for active phases
- [ ] **Weekly:** Progress review and stakeholder updates
- [ ] **Phase Gates:** Formal milestone reviews before phase transitions
- [ ] **Monthly:** Strategic review with sponsors and key stakeholders

## 8. Next Steps

### Immediate Actions (Next 48 Hours)
1. [ ] Create the project record with basic information
2. [ ] Schedule stakeholder kickoff meeting
3. [ ] Set up project workspace and communication channels
4. [ ] Begin detailed task breakdown for Phase 1

### Week 1 Deliverables
1. [ ] Completed project charter with stakeholder sign-off
2. [ ] Detailed task breakdown for Phase 1
3. [ ] Resource allocation and team commitments
4. [ ] Communication plan implementation

---
*This planning template ensures thorough preparation while maintaining flexibility for project-specific needs.*`

	return &mcp.GetPromptResult{
		Description: "Comprehensive project planning guidance and template",
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

// handleProjectStatusReviewPrompt generates a project health check template
func handleProjectStatusReviewPrompt(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.GetPromptParams,
) (*mcp.GetPromptResult, error) {
	slog.Info("Generating project_status_review prompt", "name", params.Name)

	// Extract arguments
	projectID := ""
	reviewPeriod := "weekly"

	if params.Arguments != nil {
		if id, ok := params.Arguments["project_id"]; ok {
			projectID = id
		}
		if period, ok := params.Arguments["review_period"]; ok {
			reviewPeriod = period
		}
	}

	promptText := fmt.Sprintf(`# Project Status Review: %s

## Review Overview
**Project ID:** %s
**Review Period:** %s
**Review Date:** [Current Date]

## 1. Progress Assessment

### Milestone Status
- [ ] **Current Phase:** [Which phase is the project in?]
- [ ] **Phase Progress:** [X%% complete - what's been accomplished?]
- [ ] **Next Milestone:** [What's the next major deliverable and target date?]
- [ ] **On Track:** [Yes/No - are we meeting planned timelines?]

### Task Completion Analysis
**Completed Since Last Review:**
- [ ] [Major task or deliverable 1]
- [ ] [Major task or deliverable 2]
- [ ] [Major task or deliverable 3]

**Currently In Progress:**
- [ ] [Active task 1 - owner and expected completion]
- [ ] [Active task 2 - owner and expected completion]
- [ ] [Active task 3 - owner and expected completion]

**Upcoming This Period:**
- [ ] [Planned task 1 - priority and owner]
- [ ] [Planned task 2 - priority and owner]
- [ ] [Planned task 3 - priority and owner]

## 2. Health Check Indicators

### Schedule Performance
- [ ] **Timeline Status:** [Ahead/On Track/Behind - by how much?]
- [ ] **Critical Path Impact:** [Any delays affecting critical path?]
- [ ] **Milestone Confidence:** [High/Medium/Low confidence in next milestone date]
- [ ] **Recovery Plan:** [If behind, what's the plan to recover?]

### Resource Utilization
- [ ] **Team Capacity:** [Is the team over/under/appropriately utilized?]
- [ ] **Skill Gaps:** [Any expertise shortages identified?]
- [ ] **Budget Status:** [%% of budget used vs %% of project complete]
- [ ] **External Dependencies:** [Any vendor or external team delays?]

### Quality Assessment
- [ ] **Deliverable Quality:** [Are outputs meeting quality standards?]
- [ ] **Technical Debt:** [Any shortcuts that need future attention?]
- [ ] **Stakeholder Feedback:** [What's the satisfaction level?]
- [ ] **Testing Status:** [Are quality gates being met?]

## 3. Issue and Risk Management

### Current Issues
**Active Issues Requiring Attention:**
| Issue | Impact | Owner | Target Resolution |
|-------|--------|-------|------------------|
| 1. | High/Med/Low | | |
| 2. | High/Med/Low | | |
| 3. | High/Med/Low | | |

### Emerging Risks
**New Risks Identified:**
- [ ] **Risk:** [Description]
  - **Probability:** [High/Medium/Low]
  - **Impact:** [High/Medium/Low]
  - **Mitigation Plan:** [Action to reduce risk]

**Risk Status Changes:**
- [ ] **Escalated Risks:** [Risks that have increased in probability/impact]
- [ ] **Resolved Risks:** [Risks that are no longer concerns]
- [ ] **New Dependencies:** [External factors that could affect project]

## 4. Stakeholder Communication

### Stakeholder Pulse Check
- [ ] **Sponsor Satisfaction:** [Feedback on project direction and progress]
- [ ] **End User Engagement:** [Participation in reviews and feedback quality]
- [ ] **Team Morale:** [Team satisfaction and any concerns]
- [ ] **Communication Effectiveness:** [Are updates reaching the right people?]

### Key Messages for Stakeholders
**Executive Summary (for sponsors):**
- Current status: [One sentence on overall health]
- Key achievements: [Top 2-3 accomplishments this period]
- Main concerns: [Top 1-2 issues requiring attention]
- Next steps: [What will happen in the next period]

## 5. Adaptation and Course Correction`, projectID, projectID, reviewPeriod)

	// Add period-specific guidance
	switch reviewPeriod {
	case "daily":
		promptText += `

### Daily Focus Areas
- [ ] **Immediate Blockers:** What's preventing progress today?
- [ ] **Team Coordination:** Any communication or collaboration issues?
- [ ] **Quick Wins:** Small improvements that can be made immediately?
- [ ] **Tomorrow's Priorities:** What are the top 3 tasks for tomorrow?`

	case "monthly", "milestone":
		promptText += `

### Strategic Assessment
- [ ] **Scope Changes:** Any requirements that have evolved?
- [ ] **Resource Reallocation:** Should team assignments or focus areas change?
- [ ] **Timeline Adjustments:** Are any major schedule changes needed?
- [ ] **Success Criteria:** Are the original project goals still appropriate?

### Deep Dive Analysis
- [ ] **Lessons Learned:** What insights have emerged about the work or process?
- [ ] **Process Improvements:** What working methods could be enhanced?
- [ ] **Stakeholder Evolution:** Have stakeholder needs or priorities shifted?
- [ ] **Market/External Changes:** Any external factors affecting the project?`

	default: // weekly
		promptText += `

### Weekly Adjustments
- [ ] **Priority Shifts:** Should any tasks be reprioritized for next week?
- [ ] **Resource Needs:** Any additional support or expertise needed?
- [ ] **Process Tweaks:** Small improvements to working methods?
- [ ] **Communication Updates:** Any changes needed to update frequency or format?`
	}

	promptText += `

### Recommendations and Actions
**Process Improvements:**
- [ ] [Specific change to improve efficiency or quality]
- [ ] [Communication or coordination enhancement]
- [ ] [Tool or method adjustment]

**Resource Adjustments:**
- [ ] [Team capacity or skill changes needed]
- [ ] [Budget reallocation or additional resources]
- [ ] [External support or expertise required]

**Timeline Modifications:**
- [ ] [Milestone date adjustments with rationale]
- [ ] [Task prioritization changes]
- [ ] [Scope adjustments to maintain schedule]

## 6. Action Items and Next Steps

### Immediate Actions (Next 24-48 Hours)
1. [ ] [Critical action requiring immediate attention]
2. [ ] [Important communication or decision needed]
3. [ ] [Resource issue to resolve]

### This Period's Focus
1. [ ] [Primary goal for the next review period]
2. [ ] [Secondary priority requiring attention]
3. [ ] [Preparation needed for upcoming milestones]

### Follow-up Required
- [ ] **Decisions Needed:** [What approvals or choices are pending?]
- [ ] **External Coordination:** [What requires action from others?]
- [ ] **Documentation Updates:** [What project artifacts need updating?]

## 7. Review Summary

**Overall Project Health:** [Green/Yellow/Red with brief explanation]

**Confidence Level:** [High/Medium/Low confidence in meeting project objectives]

**Key Message:** [One-sentence summary of project status for stakeholders]

**Next Review Date:** [When is the next formal review scheduled?]

---
*Regular status reviews ensure early identification of issues and maintain project momentum through proactive management.*`

	return &mcp.GetPromptResult{
		Description: "Comprehensive project status review and health check template",
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

// handleProjectRetrospectivePrompt generates a post-project analysis template
func handleProjectRetrospectivePrompt(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.GetPromptParams,
) (*mcp.GetPromptResult, error) {
	slog.Info("Generating project_retrospective prompt", "name", params.Name)

	// Extract arguments
	projectID := ""
	projectOutcome := ""

	if params.Arguments != nil {
		if id, ok := params.Arguments["project_id"]; ok {
			projectID = id
		}
		if outcome, ok := params.Arguments["project_outcome"]; ok {
			projectOutcome = outcome
		}
	}

	promptText := fmt.Sprintf(`# Project Retrospective: %s

## Retrospective Overview
**Project ID:** %s`, projectID, projectID)

	if projectOutcome != "" {
		promptText += fmt.Sprintf(`
**Project Outcome:** %s`, projectOutcome)
	}

	promptText += `
**Retrospective Date:** [Current Date]
**Participants:** [Who is contributing to this retrospective?]

## 1. Project Summary and Outcomes

### Final Results Assessment
- [ ] **Original Objectives:** [List the initial project goals]
- [ ] **Delivered Outcomes:** [What was actually accomplished]
- [ ] **Success Criteria Met:** [Which success criteria were achieved]
- [ ] **Unmet Expectations:** [What goals were not fully achieved]

### Quantitative Metrics
**Schedule Performance:**
- Planned Duration: [Original timeline]
- Actual Duration: [How long it actually took]
- Variance: [Ahead/behind by how much?]

**Resource Utilization:**
- Budget Performance: [% of planned budget used]
- Team Effort: [Planned vs. actual person-hours]
- External Costs: [Vendor/consultant expenses vs. plan]

**Quality Metrics:**
- Defect Rate: [Issues found post-delivery]
- Stakeholder Satisfaction: [Feedback scores or qualitative assessment]
- Performance Standards: [Met/exceeded/below expectations]

## 2. What Went Well (Continue)

### Process Successes
**Planning and Organization:**
- [ ] [Specific planning practice that worked well]
- [ ] [Organizational approach that was effective]
- [ ] [Documentation or communication method that succeeded]

**Team Collaboration:**
- [ ] [Team dynamic or working method that was successful]
- [ ] [Communication practice that enhanced collaboration]
- [ ] [Decision-making process that worked efficiently]

**Technical Execution:**
- [ ] [Technical approach or methodology that was effective]
- [ ] [Tool or technology that performed well]
- [ ] [Quality assurance practice that prevented issues]

### Standout Achievements
- [ ] **Innovation:** [Creative solutions or novel approaches used]
- [ ] **Efficiency:** [Processes or methods that saved time/effort]
- [ ] **Quality:** [Deliverables that exceeded expectations]
- [ ] **Collaboration:** [Exceptional teamwork or stakeholder engagement]

## 3. What Could Be Improved (Change)

### Process Challenges
**Planning and Scope Management:**
- [ ] [Planning practice that was insufficient or ineffective]
- [ ] [Scope management issue that caused problems]
- [ ] [Requirement gathering or analysis weakness]

**Communication and Coordination:**
- [ ] [Communication breakdown or inefficiency]
- [ ] [Coordination challenge between team members or stakeholders]
- [ ] [Information sharing or documentation issue]

**Resource and Timeline Management:**
- [ ] [Resource allocation or utilization problem]
- [ ] [Timeline estimation or management issue]
- [ ] [External dependency that was poorly managed]

### Technical and Quality Issues
- [ ] [Technical approach that created difficulties]
- [ ] [Quality assurance gap that allowed issues through]
- [ ] [Tool or technology that hindered rather than helped]

## 4. Lessons Learned and Insights

### Key Learning Themes
**About the Problem Domain:**
- [ ] [New understanding about the subject matter or requirements]
- [ ] [Insights about user needs or business context]
- [ ] [Discoveries about technical constraints or opportunities]

**About Team Dynamics:**
- [ ] [Insights about team composition or role definition]
- [ ] [Learning about communication styles or preferences]
- [ ] [Understanding about motivation and engagement factors]

**About Process and Methods:**
- [ ] [Process improvements that would enhance future projects]
- [ ] [Methodological insights about project management]
- [ ] [Tool or technique discoveries for future use]

### Unexpected Discoveries
**Positive Surprises:**
- [ ] [Capabilities or resources that exceeded expectations]
- [ ] [Opportunities or benefits that weren't initially anticipated]
- [ ] [Solutions or approaches that worked better than expected]

**Unforeseen Challenges:**
- [ ] [Obstacles or complexities that weren't anticipated]
- [ ] [External factors that impacted the project unexpectedly]
- [ ] [Assumptions that proved incorrect]

## 5. Impact Assessment

### Business and Organizational Impact
- [ ] **Value Delivered:** [Concrete benefits achieved for the organization]
- [ ] **User Impact:** [How end users or customers were affected]
- [ ] **Process Improvement:** [Organizational processes enhanced or created]
- [ ] **Knowledge Building:** [Expertise or capabilities developed]

### Team and Individual Development
- [ ] **Skill Development:** [New capabilities acquired by team members]
- [ ] **Career Growth:** [Professional development opportunities created]
- [ ] **Network Building:** [New relationships or collaborations established]
- [ ] **Confidence Building:** [Areas where team gained confidence or expertise]

## 6. Actionable Recommendations

### For Future Similar Projects
**Process Recommendations:**
1. [ ] **Planning:** [Specific improvement to planning approach]
2. [ ] **Execution:** [Method or practice to adopt for implementation]
3. [ ] **Quality:** [Quality assurance practice to implement]

**Resource and Management Recommendations:**
1. [ ] **Team Structure:** [Optimal team composition or role definition]
2. [ ] **Timeline:** [Approach to estimation and schedule management]
3. [ ] **Stakeholder Management:** [Strategy for stakeholder engagement]

### For Organizational Process Improvement
**Template and Standard Updates:**
- [ ] [Project template modifications based on lessons learned]
- [ ] [Standard operating procedures to update or create]
- [ ] [Training or knowledge sharing needs identified]

**Tool and Technology Recommendations:**
- [ ] [Tools that should be adopted organization-wide]
- [ ] [Technologies to investigate or invest in]
- [ ] [Infrastructure or support system improvements needed]

## 7. Knowledge Transfer and Documentation

### Project Artifacts and Outputs
**Documentation Inventory:**
- [ ] [Technical documentation created and where it's stored]
- [ ] [Process documentation or procedures developed]
- [ ] [Training materials or user guides produced]

**Code and Deliverable Repository:**
- [ ] [Where final deliverables are stored and how to access them]
- [ ] [Version control or backup procedures in place]
- [ ] [Maintenance or support procedures documented]

### Institutional Knowledge Capture
**Expertise and Contacts:**
- [ ] [Subject matter experts identified and their areas of expertise]
- [ ] [External contacts or resources valuable for future projects]
- [ ] [Internal champions or stakeholders to maintain relationships with]

**Lessons Learned Database:**
- [ ] [Key insights added to organizational knowledge base]
- [ ] [Templates or checklists created for future use]
- [ ] [Best practices documented for organizational use]

## 8. Follow-up Actions and Next Steps

### Immediate Post-Project Actions
1. [ ] **Documentation Completion:** [Any final documentation or cleanup needed]
2. [ ] **Stakeholder Communication:** [Final project closure communications]
3. [ ] **Resource Transition:** [Team member transitions or reassignments]

### Longer-term Improvement Implementation
1. [ ] **Process Updates:** [When and how process improvements will be implemented]
2. [ ] **Training Development:** [Knowledge sharing or training programs to create]
3. [ ] **Tool Adoption:** [Timeline for adopting new tools or technologies]

### Future Project Connections
- [ ] **Follow-up Projects:** [Related projects that should build on this work]
- [ ] **Maintenance Requirements:** [Ongoing support or updates needed]
- [ ] **Scaling Opportunities:** [How this project's success could be replicated]

## 9. Retrospective Summary

**Overall Assessment:** [High-level summary of project success and learning]

**Top 3 Successes to Replicate:**
1. [Most important success factor to continue]
2. [Second most important success factor]
3. [Third most important success factor]

**Top 3 Improvements for Next Time:**
1. [Most critical improvement opportunity]
2. [Second most critical improvement]
3. [Third most critical improvement]

**Key Insight:** [Most important single learning from this project]

**Recommendation for Future Projects:** [Most important advice for similar future work]

---
*This retrospective captures institutional knowledge and ensures continuous improvement in project execution and organizational capability.*`

	return &mcp.GetPromptResult{
		Description: "Comprehensive post-project retrospective analysis and lessons learned template",
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
