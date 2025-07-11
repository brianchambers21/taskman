package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/bchamber/taskman-mcp/internal/client"
	"github.com/bchamber/taskman-mcp/internal/config"
	"github.com/bchamber/taskman-mcp/internal/prompts"
	"github.com/bchamber/taskman-mcp/internal/resources"
	"github.com/bchamber/taskman-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Server struct {
	mcpServer  *mcp.Server
	apiClient  *client.APIClient
	config     *config.Config
	httpServer *http.Server
}

func NewServer(cfg *config.Config) *Server {
	slog.Info("Creating new MCP server",
		"server_name", cfg.ServerName,
		"server_version", cfg.ServerVersion,
	)

	// Create MCP server instance with proper options
	serverOptions := &mcp.ServerOptions{
		Instructions: "Taskman MCP Server - Provides task management tools, resources, and prompts",
		InitializedHandler: func(ctx context.Context, session *mcp.ServerSession, params *mcp.InitializedParams) {
			slog.Info("MCP client initialized", "session", session)
		},
		PageSize:  100,
		KeepAlive: 0, // Disable keep-alive to prevent ping timeout issues
	}

	mcpServer := mcp.NewServer(cfg.ServerName, cfg.ServerVersion, serverOptions)

	// Create API client
	apiClient := client.NewAPIClient(cfg.APIBaseURL, cfg.APITimeout)

	server := &Server{
		mcpServer: mcpServer,
		apiClient: apiClient,
		config:    cfg,
	}

	// Set up HTTP server if needed
	if cfg.TransportMode == "http" || cfg.TransportMode == "both" {
		server.setupHTTPServer()
	}

	// Register tools, resources, and prompts
	server.registerTools()
	server.registerResources()
	server.registerPrompts()

	// Add comprehensive logging middleware
	server.setupLogging()

	slog.Info("MCP server created successfully")
	return server
}

func (s *Server) registerTools() {
	slog.Info("Registering MCP tools")

	// Register a basic health check tool to demonstrate functionality
	healthTool := mcp.NewServerTool(
		"health_check",
		"Check the health of the taskman API server",
		s.handleHealthCheck,
		// No input parameters needed for health check
	)

	// Create task tools handler
	taskTools := tools.NewTaskTools(s.apiClient)

	// Create project tools handler
	projectTools := tools.NewProjectTools(s.apiClient)

	// Create user tools handler
	userTools := tools.NewUserTools(s.apiClient)

	// Register task management tools
	getTaskOverviewTool := mcp.NewServerTool(
		"get_task_overview",
		"Get a dashboard overview of tasks with status breakdown, overdue tasks, and recent activity",
		taskTools.HandleGetTaskOverview,
	)

	createTaskWithContextTool := mcp.NewServerTool(
		"create_task_with_context",
		"Create a new task with context and add an initial planning note. Valid statuses: 'Not Started', 'In Progress', 'Blocked', 'Review', 'Complete'. Valid priorities: 'Low', 'Medium', 'High'",
		taskTools.HandleCreateTaskWithContext,
	)

	getTaskDetailsTool := mcp.NewServerTool(
		"get_task_details",
		"Get complete task details including notes and project information for decision-making",
		taskTools.HandleGetTaskDetails,
	)

	updateTaskProgressTool := mcp.NewServerTool(
		"update_task_progress",
		"Update task status/progress and add a progress note. Valid statuses: 'Not Started', 'In Progress', 'Blocked', 'Review', 'Complete'. Valid priorities: 'Low', 'Medium', 'High'",
		taskTools.HandleUpdateTaskProgress,
	)

	searchTasksTool := mcp.NewServerTool(
		"search_tasks",
		"Search tasks with advanced filtering. Filter by status ('Not Started', 'In Progress', 'Blocked', 'Review', 'Complete'), priority ('Low', 'Medium', 'High'), assignee, project, creator, dates, and text",
		taskTools.HandleSearchTasks,
	)

	// Register project management tools
	getProjectStatusTool := mcp.NewServerTool(
		"get_project_status",
		"Get project overview with task breakdown, progress metrics, and insights",
		projectTools.HandleGetProjectStatus,
	)

	createProjectWithInitialTasksTool := mcp.NewServerTool(
		"create_project_with_initial_tasks",
		"Create a new project and populate it with initial tasks in one operation",
		projectTools.HandleCreateProjectWithInitialTasks,
	)

	getAllProjectsTool := mcp.NewServerTool(
		"get_all_projects",
		"Get a list of all projects in the system",
		projectTools.HandleGetAllProjects,
	)

	getAllTasksTool := mcp.NewServerTool(
		"get_all_tasks",
		"Get a list of all tasks in the system with status breakdown and insights",
		taskTools.HandleGetAllTasks,
	)

	addTaskNoteTool := mcp.NewServerTool(
		"add_task_note",
		"Add a note to an existing task without requiring status or other changes",
		taskTools.HandleAddTaskNote,
	)

	// Register user-focused tools
	getMyWorkTool := mcp.NewServerTool(
		"get_my_work",
		"Get personalized work queue with prioritized tasks and workload insights",
		userTools.HandleGetMyWork,
	)

	s.mcpServer.AddTools(
		healthTool,
		getTaskOverviewTool,
		createTaskWithContextTool,
		getTaskDetailsTool,
		updateTaskProgressTool,
		searchTasksTool,
		getProjectStatusTool,
		createProjectWithInitialTasksTool,
		getAllProjectsTool,
		getAllTasksTool,
		addTaskNoteTool,
		getMyWorkTool,
	)

	slog.Info("Tools registration completed", "tool_count", 12)
}

// Health check tool handler
func (s *Server) handleHealthCheck(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.CallToolParamsFor[map[string]any],
) (*mcp.CallToolResultFor[map[string]any], error) {
	slog.Info("Executing health_check tool")

	// Make API call to health endpoint
	resp, err := s.apiClient.Get(ctx, "/health")
	if err != nil {
		slog.Error("Health check failed", "error", err)
		return &mcp.CallToolResultFor[map[string]any]{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Health check failed: %v", err),
				},
			},
		}, nil
	}

	result := map[string]any{
		"status":       "healthy",
		"api_response": string(resp),
		"timestamp":    time.Now().Format(time.RFC3339),
	}

	slog.Info("Health check completed successfully")

	return &mcp.CallToolResultFor[map[string]any]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("API Health Check: %s", result["status"]),
			},
		},
		Meta: result,
	}, nil
}

func (s *Server) registerResources() {
	slog.Info("Registering MCP resources")

	// Create resource handlers
	taskResources := resources.NewTaskResources(s.apiClient)
	projectResources := resources.NewProjectResources(s.apiClient)
	dashboardResources := resources.NewDashboardResources(s.apiClient)

	// Register API status resource
	statusResource := &mcp.ServerResource{
		Resource: &mcp.Resource{
			URI:         "taskman://api/status",
			Name:        "API Status",
			Description: "Current status of the taskman API server",
			MIMEType:    "application/json",
		},
		Handler: s.handleStatusResource,
	}

	// Register task resources
	taskResource := &mcp.ServerResource{
		Resource: &mcp.Resource{
			URI:         "taskman://task/{task_id}",
			Name:        "Task Details",
			Description: "Individual task with complete details, notes, and project information",
			MIMEType:    "text/plain",
		},
		Handler: taskResources.HandleTaskResource,
	}

	tasksOverviewResource := &mcp.ServerResource{
		Resource: &mcp.Resource{
			URI:         "taskman://tasks/overview",
			Name:        "Tasks Overview",
			Description: "Overview of all tasks with status and priority breakdowns",
			MIMEType:    "text/plain",
		},
		Handler: taskResources.HandleTasksOverviewResource,
	}

	userTasksResource := &mcp.ServerResource{
		Resource: &mcp.Resource{
			URI:         "taskman://tasks/user/{user_id}",
			Name:        "User Tasks",
			Description: "All tasks assigned to a specific user, organized by status",
			MIMEType:    "text/plain",
		},
		Handler: taskResources.HandleUserTasksResource,
	}

	// Register project resources
	projectResource := &mcp.ServerResource{
		Resource: &mcp.Resource{
			URI:         "taskman://project/{project_id}",
			Name:        "Project Details",
			Description: "Individual project with task summary and progress metrics",
			MIMEType:    "text/plain",
		},
		Handler: projectResources.HandleProjectResource,
	}

	projectsOverviewResource := &mcp.ServerResource{
		Resource: &mcp.Resource{
			URI:         "taskman://projects/overview",
			Name:        "Projects Overview",
			Description: "Overview of all projects with creation statistics",
			MIMEType:    "text/plain",
		},
		Handler: projectResources.HandleProjectsOverviewResource,
	}

	projectTasksResource := &mcp.ServerResource{
		Resource: &mcp.Resource{
			URI:         "taskman://project/{project_id}/tasks",
			Name:        "Project Tasks",
			Description: "All tasks within a specific project, organized by status",
			MIMEType:    "text/plain",
		},
		Handler: projectResources.HandleProjectTasksResource,
	}

	// Register dashboard resources
	systemDashboardResource := &mcp.ServerResource{
		Resource: &mcp.Resource{
			URI:         "taskman://dashboard/system",
			Name:        "System Dashboard",
			Description: "System-wide dashboard with overall statistics and insights",
			MIMEType:    "text/plain",
		},
		Handler: dashboardResources.HandleSystemDashboardResource,
	}

	userDashboardResource := &mcp.ServerResource{
		Resource: &mcp.Resource{
			URI:         "taskman://dashboard/user/{user_id}",
			Name:        "User Dashboard",
			Description: "Personalized dashboard for a specific user with workload and deadlines",
			MIMEType:    "text/plain",
		},
		Handler: dashboardResources.HandleUserDashboardResource,
	}

	projectDashboardResource := &mcp.ServerResource{
		Resource: &mcp.Resource{
			URI:         "taskman://dashboard/project/{project_id}",
			Name:        "Project Dashboard",
			Description: "Project-specific dashboard with team workload and critical tasks",
			MIMEType:    "text/plain",
		},
		Handler: dashboardResources.HandleProjectDashboardResource,
	}

	// Add all resources to the server
	s.mcpServer.AddResources(
		statusResource,
		taskResource,
		tasksOverviewResource,
		userTasksResource,
		projectResource,
		projectsOverviewResource,
		projectTasksResource,
		systemDashboardResource,
		userDashboardResource,
		projectDashboardResource,
	)

	slog.Info("Resources registration completed", "resource_count", 10)
}

// Status resource handler
func (s *Server) handleStatusResource(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.ReadResourceParams,
) (*mcp.ReadResourceResult, error) {
	slog.Info("Reading status resource", "uri", params.URI)

	// Make API call to health endpoint
	resp, err := s.apiClient.Get(ctx, "/health")
	if err != nil {
		slog.Error("Failed to read API status", "error", err)
		return nil, fmt.Errorf("failed to read API status: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      params.URI,
				MIMEType: "application/json",
				Text:     string(resp),
			},
		},
	}, nil
}

func (s *Server) registerPrompts() {
	slog.Info("Registering MCP prompts")

	// Register a basic task creation prompt template
	taskPrompt := &mcp.ServerPrompt{
		Prompt: &mcp.Prompt{
			Name:        "create_task",
			Description: "Template for creating a new task with proper context",
			Arguments: []*mcp.PromptArgument{
				{
					Name:        "task_name",
					Description: "Name of the task to create",
					Required:    true,
				},
				{
					Name:        "project_id",
					Description: "Optional project ID to associate with the task",
					Required:    false,
				},
			},
		},
		Handler: s.handleCreateTaskPrompt,
	}

	// Collect all prompts from different modules
	allPrompts := []*mcp.ServerPrompt{taskPrompt}

	// Add task-related prompts
	taskPrompts := prompts.CreateTaskPrompts()
	allPrompts = append(allPrompts, taskPrompts...)

	// Add project-related prompts
	projectPrompts := prompts.CreateProjectPrompts()
	allPrompts = append(allPrompts, projectPrompts...)

	// Add workflow-related prompts
	workflowPrompts := prompts.CreateWorkflowPrompts()
	allPrompts = append(allPrompts, workflowPrompts...)

	// Register all prompts with the MCP server
	s.mcpServer.AddPrompts(allPrompts...)

	slog.Info("Prompts registration completed", "prompt_count", len(allPrompts))
}

// Create task prompt handler
func (s *Server) handleCreateTaskPrompt(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.GetPromptParams,
) (*mcp.GetPromptResult, error) {
	slog.Info("Generating create_task prompt", "name", params.Name)

	// Extract arguments
	taskName := ""
	projectID := ""

	if params.Arguments != nil {
		if name, ok := params.Arguments["task_name"]; ok {
			taskName = name
		}
		if pid, ok := params.Arguments["project_id"]; ok {
			projectID = pid
		}
	}

	// Generate prompt text
	promptText := fmt.Sprintf(`Create a new task with the following details:

Task Name: %s`, taskName)

	if projectID != "" {
		promptText += fmt.Sprintf(`
Project ID: %s`, projectID)
	}

	promptText += `

Please provide:
1. A detailed description for this task
2. Appropriate priority level (Low, Medium, High)
3. Estimated completion timeline
4. Any dependencies or prerequisites
5. Success criteria for completion`

	return &mcp.GetPromptResult{
		Description: "Task creation guidance prompt",
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

func (s *Server) setupHTTPServer() {
	mux := http.NewServeMux()

	// Create SSE handler that provides access to our server
	sseHandler := mcp.NewSSEHandler(func(r *http.Request) *mcp.Server {
		return s.mcpServer
	})

	// Set up SSE endpoint for streaming connections
	mux.Handle("/sse", sseHandler)

	// Set up streamable HTTP handler for HTTP transport
	streamableHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return s.mcpServer
	}, nil)

	// Set up streamable HTTP endpoint
	mux.Handle("/mcp", streamableHandler)

	addr := fmt.Sprintf("%s:%s", s.config.HTTPHost, s.config.HTTPPort)
	s.httpServer = &http.Server{
		Addr:           addr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	slog.Info("HTTP server configured",
		"address", addr,
		"sse_endpoint", "/sse",
		"http_endpoint", "/mcp",
	)
}

func (s *Server) Run(ctx context.Context) error {
	slog.Info("Starting MCP server", "transport_mode", s.config.TransportMode)

	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	// Start stdio transport if needed
	if s.config.TransportMode == "stdio" || s.config.TransportMode == "both" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			slog.Info("Starting MCP server with stdio transport")

			transport := mcp.NewStdioTransport()
			if err := s.mcpServer.Run(ctx, transport); err != nil {
				slog.Error("Stdio MCP server failed", "error", err)
				errCh <- fmt.Errorf("stdio transport error: %w", err)
			}
		}()
	}

	// Start HTTP server if needed
	if s.config.TransportMode == "http" || s.config.TransportMode == "both" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			slog.Info("Starting HTTP server", "address", s.httpServer.Addr)

			if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("HTTP server failed", "error", err)
				errCh <- fmt.Errorf("HTTP server error: %w", err)
			}
		}()

		// Handle graceful HTTP server shutdown
		go func() {
			<-ctx.Done()
			if s.httpServer != nil {
				slog.Info("Shutting down HTTP server")
				if err := s.httpServer.Shutdown(context.Background()); err != nil {
					slog.Error("HTTP server shutdown error", "error", err)
				}
			}
		}()
	}

	// Wait for either completion or error
	go func() {
		wg.Wait()
		close(errCh)
	}()

	// Return first error if any
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		slog.Info("Server stopped by context cancellation")
	}

	slog.Info("MCP server stopped")
	return nil
}

// setupLogging configures comprehensive logging for the MCP server
func (s *Server) setupLogging() {
	slog.Info("Setting up comprehensive MCP request/response logging")
	
	// Create logging middleware for incoming requests
	loggingMiddleware := s.createLoggingMiddleware()
	
	// Add middleware for both receiving and sending
	s.mcpServer.AddReceivingMiddleware(loggingMiddleware)
	s.mcpServer.AddSendingMiddleware(loggingMiddleware)
	
	slog.Info("MCP logging middleware configured")
}

// createLoggingMiddleware creates middleware that logs all MCP requests and responses
func (s *Server) createLoggingMiddleware() mcp.Middleware[*mcp.ServerSession] {
	return func(next mcp.MethodHandler[*mcp.ServerSession]) mcp.MethodHandler[*mcp.ServerSession] {
		return func(ctx context.Context, session *mcp.ServerSession, method string, params mcp.Params) (mcp.Result, error) {
			// Handle ping requests with minimal logging
			if method == "ping" {
				start := time.Now()
				result, err := next(ctx, session, method, params)
				duration := time.Since(start)
				
				if err != nil {
					slog.Error("PING FAILED", "error", err, "duration_ms", duration.Milliseconds())
				} else {
					slog.Info("PING OK", "duration_ms", duration.Milliseconds())
				}
				return result, err
			}
			
			start := time.Now()
			
			// Log incoming request with full details
			slog.Info("=== MCP REQUEST START ===",
				"method", method,
				"timestamp", start.Format(time.RFC3339Nano),
				"session_info", fmt.Sprintf("%+v", session),
			)
			
			// Log parameters in detail
			if params != nil {
				slog.Info("MCP Request Parameters",
					"method", method,
					"params_type", fmt.Sprintf("%T", params),
					"params_value", fmt.Sprintf("%+v", params),
				)
				
				// Try to marshal params to see raw JSON
				if paramsJSON, err := json.Marshal(params); err == nil {
					slog.Info("MCP Request Parameters JSON",
						"method", method,
						"params_json", string(paramsJSON),
					)
				}
			} else {
				slog.Info("MCP Request Parameters",
					"method", method,
					"params", "null",
				)
			}
			
			// Execute the handler
			result, err := next(ctx, session, method, params)
			
			duration := time.Since(start)
			
			// Log response with full details
			if err != nil {
				slog.Error("=== MCP REQUEST FAILED ===",
					"method", method,
					"error", err.Error(),
					"error_type", fmt.Sprintf("%T", err),
					"duration_ms", duration.Milliseconds(),
					"timestamp", time.Now().Format(time.RFC3339Nano),
				)
			} else {
				slog.Info("=== MCP REQUEST SUCCESS ===",
					"method", method,
					"result_type", fmt.Sprintf("%T", result),
					"duration_ms", duration.Milliseconds(),
					"timestamp", time.Now().Format(time.RFC3339Nano),
				)
				
				// Log result details
				if result != nil {
					slog.Info("MCP Response Result",
						"method", method,
						"result_value", fmt.Sprintf("%+v", result),
					)
					
					// Try to marshal result to see raw JSON
					if resultJSON, err := json.Marshal(result); err == nil {
						slog.Info("MCP Response Result JSON",
							"method", method,
							"result_json", string(resultJSON),
						)
					}
				}
			}
			
			return result, err
		}
	}
}
