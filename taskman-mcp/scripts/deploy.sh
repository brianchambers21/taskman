#!/bin/bash

# Task Management MCP Server - Deployment Script
# This script handles deployment of the taskman-mcp server

set -e

# Configuration
BINARY_NAME="taskman-mcp"
BUILD_DIR="./bin"
LOG_DIR="./logs"
PID_FILE="./taskman-mcp.pid"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
    exit 1
}

# Check if .env file exists
check_environment() {
    if [ ! -f "../.env" ]; then
        error ".env file not found. Please create it from .env.example"
    fi
    log "Environment configuration found"
}

# Check API server connectivity
check_api_server() {
    API_URL="${TASKMAN_API_BASE_URL:-http://localhost:8080}"
    
    log "Checking API server connectivity at $API_URL..."
    
    if curl -s -f "${API_URL}/api/v1/tasks" > /dev/null; then
        log "✓ API server is accessible"
    else
        error "✗ API server is not accessible at $API_URL. Please start the API server first."
    fi
}

# Build the application
build_application() {
    log "Building taskman-mcp..."
    
    # Create build directory
    mkdir -p "${BUILD_DIR}"
    
    # Build the MCP server
    go build -o "${BUILD_DIR}/${BINARY_NAME}" cmd/server/main.go
    
    log "Build completed successfully"
}

# Start the MCP server
start_server() {
    log "Starting taskman-mcp server..."
    
    # Create log directory
    mkdir -p "${LOG_DIR}"
    
    # Check if server is already running
    if [ -f "${PID_FILE}" ]; then
        PID=$(cat "${PID_FILE}")
        if ps -p "$PID" > /dev/null 2>&1; then
            warn "Server is already running with PID $PID"
            return 0
        else
            log "Removing stale PID file"
            rm -f "${PID_FILE}"
        fi
    fi
    
    # Determine transport mode
    TRANSPORT_MODE="${TASKMAN_TRANSPORT_MODE:-stdio}"
    
    case "$TRANSPORT_MODE" in
        stdio)
            log "Starting MCP server in stdio mode..."
            log "Note: stdio mode requires connection from MCP client. Server will wait for input."
            "${BUILD_DIR}/${BINARY_NAME}" > "${LOG_DIR}/mcp.log" 2>&1 &
            ;;
        http)
            log "Starting MCP server in HTTP mode..."
            nohup "${BUILD_DIR}/${BINARY_NAME}" > "${LOG_DIR}/mcp.log" 2>&1 &
            ;;
        both)
            log "Starting MCP server in both stdio and HTTP modes..."
            nohup "${BUILD_DIR}/${BINARY_NAME}" > "${LOG_DIR}/mcp.log" 2>&1 &
            ;;
        *)
            error "Invalid transport mode: $TRANSPORT_MODE. Valid options: stdio, http, both"
            ;;
    esac
    
    echo $! > "${PID_FILE}"
    sleep 2
    
    # Verify server started
    PID=$(cat "${PID_FILE}")
    if ps -p "$PID" > /dev/null 2>&1; then
        log "Server started successfully with PID $PID"
        log "Transport mode: $TRANSPORT_MODE"
        log "Logs available at: ${LOG_DIR}/mcp.log"
        
        if [ "$TRANSPORT_MODE" = "http" ] || [ "$TRANSPORT_MODE" = "both" ]; then
            HTTP_PORT="${TASKMAN_HTTP_PORT:-8081}"
            HTTP_HOST="${TASKMAN_HTTP_HOST:-localhost}"
            log "HTTP endpoint: http://${HTTP_HOST}:${HTTP_PORT}"
        fi
    else
        error "Failed to start server"
    fi
}

# Stop the MCP server
stop_server() {
    log "Stopping taskman-mcp server..."
    
    if [ ! -f "${PID_FILE}" ]; then
        warn "PID file not found. Server may not be running."
        return 0
    fi
    
    PID=$(cat "${PID_FILE}")
    if ps -p "$PID" > /dev/null 2>&1; then
        kill "$PID"
        sleep 2
        
        # Force kill if still running
        if ps -p "$PID" > /dev/null 2>&1; then
            warn "Force killing server process"
            kill -9 "$PID"
        fi
        
        rm -f "${PID_FILE}"
        log "Server stopped successfully"
    else
        warn "Server process $PID not found"
        rm -f "${PID_FILE}"
    fi
}

# Check server status
check_status() {
    if [ -f "${PID_FILE}" ]; then
        PID=$(cat "${PID_FILE}")
        if ps -p "$PID" > /dev/null 2>&1; then
            log "Server is running with PID $PID"
            
            TRANSPORT_MODE="${TASKMAN_TRANSPORT_MODE:-stdio}"
            log "Transport mode: $TRANSPORT_MODE"
            
            # Test HTTP endpoint if applicable
            if [ "$TRANSPORT_MODE" = "http" ] || [ "$TRANSPORT_MODE" = "both" ]; then
                HTTP_PORT="${TASKMAN_HTTP_PORT:-8081}"
                HTTP_HOST="${TASKMAN_HTTP_HOST:-localhost}"
                
                if curl -s "http://${HTTP_HOST}:${HTTP_PORT}" > /dev/null; then
                    log "HTTP endpoint is responding"
                else
                    warn "HTTP endpoint is not responding"
                fi
            fi
        else
            warn "PID file exists but process is not running"
        fi
    else
        log "Server is not running"
    fi
}

# Run tests
run_tests() {
    log "Running tests..."
    go test ./... -v
    log "Tests completed"
}

# Health check
health_check() {
    log "Performing health check..."
    
    # Check if process is running
    if [ -f "${PID_FILE}" ]; then
        PID=$(cat "${PID_FILE}")
        if ps -p "$PID" > /dev/null 2>&1; then
            log "✓ MCP server process is running"
        else
            error "✗ MCP server process is not running"
        fi
    else
        error "✗ MCP server is not started"
    fi
    
    # Check API connectivity
    API_URL="${TASKMAN_API_BASE_URL:-http://localhost:8080}"
    if curl -s -f "${API_URL}/api/v1/tasks" > /dev/null; then
        log "✓ API server connectivity is healthy"
    else
        error "✗ Cannot connect to API server at $API_URL"
    fi
    
    # Check HTTP endpoint if in HTTP mode
    TRANSPORT_MODE="${TASKMAN_TRANSPORT_MODE:-stdio}"
    if [ "$TRANSPORT_MODE" = "http" ] || [ "$TRANSPORT_MODE" = "both" ]; then
        HTTP_PORT="${TASKMAN_HTTP_PORT:-8081}"
        HTTP_HOST="${TASKMAN_HTTP_HOST:-localhost}"
        
        if curl -s "http://${HTTP_HOST}:${HTTP_PORT}" > /dev/null; then
            log "✓ HTTP transport is healthy"
        else
            warn "✗ HTTP transport health check failed"
        fi
    fi
}

# Test MCP functionality
test_mcp_functionality() {
    log "Testing MCP functionality..."
    
    # Check if API server is available
    check_api_server
    
    # Run MCP-specific tests
    log "Running MCP integration tests..."
    go test ./test/... -v
    
    log "MCP functionality tests completed"
}

# Show usage
usage() {
    echo "Usage: $0 {build|deploy|start|stop|restart|status|test|health|test-mcp}"
    echo ""
    echo "Commands:"
    echo "  build     - Build the application"
    echo "  deploy    - Full deployment (build + start)"
    echo "  start     - Start the server"
    echo "  stop      - Stop the server"
    echo "  restart   - Restart the server"
    echo "  status    - Check server status"
    echo "  test      - Run tests"
    echo "  health    - Perform health check"
    echo "  test-mcp  - Test MCP functionality with API"
    echo ""
    echo "Environment Variables:"
    echo "  TASKMAN_TRANSPORT_MODE - stdio|http|both (default: stdio)"
    echo "  TASKMAN_API_BASE_URL   - API server URL (default: http://localhost:8080)"
    echo "  TASKMAN_HTTP_PORT      - HTTP port for MCP server (default: 8081)"
    echo "  TASKMAN_HTTP_HOST      - HTTP host for MCP server (default: localhost)"
    exit 1
}

# Main script logic
case "$1" in
    build)
        check_environment
        build_application
        ;;
    deploy)
        check_environment
        check_api_server
        build_application
        start_server
        sleep 3
        health_check
        ;;
    start)
        check_environment
        check_api_server
        start_server
        ;;
    stop)
        stop_server
        ;;
    restart)
        stop_server
        sleep 2
        start_server
        ;;
    status)
        check_status
        ;;
    test)
        run_tests
        ;;
    health)
        health_check
        ;;
    test-mcp)
        test_mcp_functionality
        ;;
    *)
        usage
        ;;
esac