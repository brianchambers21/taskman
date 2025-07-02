#!/bin/bash

# Task Management System - Full System Deployment Script
# This script orchestrates the deployment of the complete taskman system

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $1${NC}"
}

# Check prerequisites
check_prerequisites() {
    log "Checking system prerequisites..."
    
    # Check Go installation
    if ! command -v go &> /dev/null; then
        error "Go is not installed. Please install Go 1.21 or higher."
    fi
    
    GO_VERSION=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | sed 's/go//')
    info "Go version: $GO_VERSION"
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        error "Docker is not installed. Please install Docker."
    fi
    
    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        error "Docker Compose is not installed. Please install Docker Compose."
    fi
    
    # Check curl
    if ! command -v curl &> /dev/null; then
        error "curl is not installed. Please install curl."
    fi
    
    log "Prerequisites check completed"
}

# Setup environment
setup_environment() {
    log "Setting up environment..."
    
    if [ ! -f ".env" ]; then
        if [ -f ".env.example" ]; then
            cp .env.example .env
            log "Created .env from .env.example"
            warn "Please review and update .env file with your settings"
        else
            error ".env.example file not found"
        fi
    else
        log "Environment file .env already exists"
    fi
}

# Start database
start_database() {
    log "Starting PostgreSQL database..."
    
    if docker-compose ps | grep -q postgres; then
        log "Database is already running"
    else
        docker-compose up -d postgres
        
        # Wait for database to be ready
        log "Waiting for database to be ready..."
        for i in {1..30}; do
            if docker-compose exec -T postgres pg_isready -U taskman_user -d taskman &> /dev/null; then
                log "Database is ready"
                break
            fi
            
            if [ $i -eq 30 ]; then
                error "Database failed to start within 30 seconds"
            fi
            
            sleep 1
        done
    fi
}

# Build and deploy API server
deploy_api() {
    log "Deploying API server..."
    
    cd taskman-api
    
    # Build and deploy
    ./scripts/deploy.sh deploy
    
    # Wait for API to be ready
    log "Waiting for API server to be ready..."
    for i in {1..30}; do
        if curl -s -f "http://localhost:${TASKMAN_API_PORT:-8080}/api/v1/tasks" > /dev/null; then
            log "API server is ready"
            break
        fi
        
        if [ $i -eq 30 ]; then
            error "API server failed to start within 30 seconds"
        fi
        
        sleep 1
    done
    
    cd ..
}

# Build and deploy MCP server
deploy_mcp() {
    log "Deploying MCP server..."
    
    cd taskman-mcp
    
    # Build and deploy
    ./scripts/deploy.sh deploy
    
    cd ..
}

# Run system tests
run_system_tests() {
    log "Running system-wide tests..."
    
    # Run API tests
    log "Running API tests..."
    cd taskman-api
    go test ./... -v
    cd ..
    
    # Run MCP tests
    log "Running MCP tests..."
    cd taskman-mcp
    go test ./... -v
    cd ..
    
    # Run E2E tests
    log "Running end-to-end tests..."
    go test ./test/... -v
    
    log "All tests completed successfully"
}

# System health check
system_health_check() {
    log "Performing system-wide health check..."
    
    # Check database
    if docker-compose exec -T postgres pg_isready -U taskman_user -d taskman &> /dev/null; then
        log "✓ Database is healthy"
    else
        error "✗ Database health check failed"
    fi
    
    # Check API server
    API_PORT="${TASKMAN_API_PORT:-8080}"
    if curl -s -f "http://localhost:${API_PORT}/api/v1/tasks" > /dev/null; then
        log "✓ API server is healthy"
    else
        error "✗ API server health check failed"
    fi
    
    # Check MCP server (via deployment script)
    cd taskman-mcp
    ./scripts/deploy.sh health
    cd ..
    
    log "System health check completed successfully"
}

# Stop all services
stop_services() {
    log "Stopping all services..."
    
    # Stop MCP server
    cd taskman-mcp
    ./scripts/deploy.sh stop
    cd ..
    
    # Stop API server
    cd taskman-api
    ./scripts/deploy.sh stop
    cd ..
    
    # Stop database
    docker-compose down
    
    log "All services stopped"
}

# Show system status
show_status() {
    log "System Status:"
    
    # Database status
    if docker-compose ps | grep -q postgres; then
        log "✓ Database: Running"
    else
        log "✗ Database: Stopped"
    fi
    
    # API server status
    cd taskman-api
    ./scripts/deploy.sh status
    cd ..
    
    # MCP server status
    cd taskman-mcp
    ./scripts/deploy.sh status
    cd ..
}

# Show system information
show_info() {
    echo ""
    info "=== Task Management System ==="
    info "API Server: http://localhost:${TASKMAN_API_PORT:-8080}"
    info "MCP Server: Transport mode ${TASKMAN_TRANSPORT_MODE:-stdio}"
    
    if [ "${TASKMAN_TRANSPORT_MODE}" = "http" ] || [ "${TASKMAN_TRANSPORT_MODE}" = "both" ]; then
        info "MCP HTTP: http://localhost:${TASKMAN_HTTP_PORT:-8081}"
    fi
    
    info "Database: PostgreSQL on port ${TASKMAN_DB_PORT:-5433}"
    echo ""
    info "Logs:"
    info "  API: taskman-api/logs/api.log"
    info "  MCP: taskman-mcp/logs/mcp.log"
    echo ""
}

# Show usage
usage() {
    echo "Usage: $0 {deploy|start|stop|restart|status|test|health|info}"
    echo ""
    echo "Commands:"
    echo "  deploy    - Full system deployment (database + API + MCP)"
    echo "  start     - Start all services (database + API + MCP)"
    echo "  stop      - Stop all services"
    echo "  restart   - Restart all services"
    echo "  status    - Show status of all services"
    echo "  test      - Run all system tests"
    echo "  health    - Perform system health check"
    echo "  info      - Show system information"
    echo ""
    echo "Prerequisites:"
    echo "  - Go 1.21 or higher"
    echo "  - Docker and Docker Compose"
    echo "  - curl"
    exit 1
}

# Main script logic
case "$1" in
    deploy)
        check_prerequisites
        setup_environment
        start_database
        deploy_api
        deploy_mcp
        system_health_check
        show_info
        ;;
    start)
        start_database
        deploy_api
        deploy_mcp
        show_info
        ;;
    stop)
        stop_services
        ;;
    restart)
        stop_services
        sleep 3
        start_database
        deploy_api
        deploy_mcp
        show_info
        ;;
    status)
        show_status
        ;;
    test)
        run_system_tests
        ;;
    health)
        system_health_check
        ;;
    info)
        show_info
        ;;
    *)
        usage
        ;;
esac