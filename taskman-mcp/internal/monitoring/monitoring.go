package monitoring

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	monitoringpb "cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"google.golang.org/api/option"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Monitor handles Google Cloud monitoring integration for MCP server
type Monitor struct {
	client    *monitoring.MetricClient
	projectID string
	enabled   bool
	mutex     sync.RWMutex
	metrics   map[string]*MetricTracker
}

// MetricTracker tracks individual metrics
type MetricTracker struct {
	Name        string
	Type        string
	Count       int64
	LastValue   float64
	LastUpdated time.Time
}

// Config for monitoring setup
type Config struct {
	ProjectID           string
	Enabled             bool
	CredentialsFile     string
	MetricPrefix        string
	FlushIntervalMinutes int
}

var (
	defaultMonitor *Monitor
	once           sync.Once
)

// Initialize sets up Google Cloud monitoring for MCP server
func Initialize(config Config) error {
	var err error
	once.Do(func() {
		defaultMonitor, err = NewMonitor(config)
	})
	return err
}

// NewMonitor creates a new monitoring instance
func NewMonitor(config Config) (*Monitor, error) {
	if !config.Enabled {
		slog.Info("Google Cloud monitoring disabled for MCP server")
		return &Monitor{
			enabled: false,
			metrics: make(map[string]*MetricTracker),
		}, nil
	}

	ctx := context.Background()
	var opts []option.ClientOption

	if config.CredentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(config.CredentialsFile))
	}

	client, err := monitoring.NewMetricClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create monitoring client: %w", err)
	}

	monitor := &Monitor{
		client:    client,
		projectID: config.ProjectID,
		enabled:   true,
		metrics:   make(map[string]*MetricTracker),
	}

	slog.Info("Google Cloud monitoring initialized for MCP server", "project", config.ProjectID)

	// Start background metric flusher
	if config.FlushIntervalMinutes > 0 {
		go monitor.startMetricFlusher(time.Duration(config.FlushIntervalMinutes) * time.Minute)
	}

	return monitor, nil
}

// GetDefault returns the default monitor instance
func GetDefault() *Monitor {
	if defaultMonitor == nil {
		return &Monitor{
			enabled: false,
			metrics: make(map[string]*MetricTracker),
		}
	}
	return defaultMonitor
}

// IncrementCounter increments a counter metric
func (m *Monitor) IncrementCounter(metricName string) {
	if !m.enabled {
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if tracker, exists := m.metrics[metricName]; exists {
		tracker.Count++
		tracker.LastUpdated = time.Now()
	} else {
		m.metrics[metricName] = &MetricTracker{
			Name:        metricName,
			Type:        "counter",
			Count:       1,
			LastUpdated: time.Now(),
		}
	}
}

// SetGauge sets a gauge metric value
func (m *Monitor) SetGauge(metricName string, value float64) {
	if !m.enabled {
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if tracker, exists := m.metrics[metricName]; exists {
		tracker.LastValue = value
		tracker.LastUpdated = time.Now()
	} else {
		m.metrics[metricName] = &MetricTracker{
			Name:        metricName,
			Type:        "gauge",
			LastValue:   value,
			LastUpdated: time.Now(),
		}
	}
}

// RecordLatency records a latency metric
func (m *Monitor) RecordLatency(metricName string, duration time.Duration) {
	if !m.enabled {
		return
	}

	latencyMs := float64(duration.Nanoseconds()) / 1e6
	m.SetGauge(metricName+"_latency_ms", latencyMs)
}

// SendMetric sends a single metric to Google Cloud Monitoring
func (m *Monitor) SendMetric(ctx context.Context, metricType, metricName string, value float64, labels map[string]string) error {
	if !m.enabled {
		return nil
	}

	now := &timestamppb.Timestamp{
		Seconds: time.Now().Unix(),
	}

	// Create metric descriptor name
	metricDescriptor := fmt.Sprintf("custom.googleapis.com/taskman-mcp/%s", metricName)

	// Create the data point
	dataPoint := &monitoringpb.Point{
		Interval: &monitoringpb.TimeInterval{
			EndTime: now,
		},
		Value: &monitoringpb.TypedValue{
			Value: &monitoringpb.TypedValue_DoubleValue{
				DoubleValue: value,
			},
		},
	}

	// Create metric labels
	metricLabels := make(map[string]string)
	for k, v := range labels {
		metricLabels[k] = v
	}
	metricLabels["service"] = "taskman-mcp"
	metricLabels["version"] = getVersion()

	// Create the time series
	timeSeries := &monitoringpb.TimeSeries{
		Metric: &metricpb.Metric{
			Type:   metricDescriptor,
			Labels: metricLabels,
		},
		Resource: &monitoredrespb.MonitoredResource{
			Type: "generic_task",
			Labels: map[string]string{
				"project_id": m.projectID,
				"location":   "global",
				"namespace":  "taskman-mcp",
				"task_id":    "mcp-server",
			},
		},
		Points: []*monitoringpb.Point{dataPoint},
	}

	// Send the metric
	req := &monitoringpb.CreateTimeSeriesRequest{
		Name:       "projects/" + m.projectID,
		TimeSeries: []*monitoringpb.TimeSeries{timeSeries},
	}

	err := m.client.CreateTimeSeries(ctx, req)
	if err != nil {
		slog.Error("Failed to send metric to Google Cloud", "metric", metricName, "error", err)
		return err
	}

	slog.Debug("Metric sent to Google Cloud", "metric", metricName, "value", value)
	return nil
}

// FlushMetrics sends all accumulated metrics to Google Cloud
func (m *Monitor) FlushMetrics(ctx context.Context) error {
	if !m.enabled {
		return nil
	}

	m.mutex.RLock()
	metricsToSend := make(map[string]*MetricTracker)
	for k, v := range m.metrics {
		metricsToSend[k] = &MetricTracker{
			Name:        v.Name,
			Type:        v.Type,
			Count:       v.Count,
			LastValue:   v.LastValue,
			LastUpdated: v.LastUpdated,
		}
	}
	m.mutex.RUnlock()

	var errors []error
	for _, tracker := range metricsToSend {
		var value float64
		if tracker.Type == "counter" {
			value = float64(tracker.Count)
		} else {
			value = tracker.LastValue
		}

		err := m.SendMetric(ctx, tracker.Type, tracker.Name, value, nil)
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		slog.Error("Some metrics failed to send", "count", len(errors))
		return errors[0]
	}

	slog.Info("Metrics flushed to Google Cloud", "count", len(metricsToSend))
	return nil
}

// startMetricFlusher starts a background goroutine to periodically flush metrics
func (m *Monitor) startMetricFlusher(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := m.FlushMetrics(ctx); err != nil {
				slog.Error("Failed to flush metrics", "error", err)
			}
			cancel()
		}
	}
}

// Close closes the monitoring client
func (m *Monitor) Close() error {
	if !m.enabled || m.client == nil {
		return nil
	}

	// Final flush before closing
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := m.FlushMetrics(ctx); err != nil {
		slog.Error("Failed to flush metrics before closing", "error", err)
	}

	return m.client.Close()
}

// MCP-specific helper functions
func IncrementMCPToolCall(toolName string) {
	monitor := GetDefault()
	monitor.IncrementCounter("mcp_tool_calls_total")
	monitor.IncrementCounter(fmt.Sprintf("mcp_tool_calls_%s", toolName))
}

func IncrementMCPToolError(toolName string) {
	monitor := GetDefault()
	monitor.IncrementCounter("mcp_tool_errors_total")
	monitor.IncrementCounter(fmt.Sprintf("mcp_tool_errors_%s", toolName))
}

func RecordMCPToolLatency(toolName string, duration time.Duration) {
	monitor := GetDefault()
	monitor.RecordLatency(fmt.Sprintf("mcp_tool_latency_%s", toolName), duration)
}

func IncrementMCPResourceAccess(resourceType string) {
	monitor := GetDefault()
	monitor.IncrementCounter("mcp_resource_access_total")
	monitor.IncrementCounter(fmt.Sprintf("mcp_resource_access_%s", resourceType))
}

func IncrementMCPPromptGeneration(promptName string) {
	monitor := GetDefault()
	monitor.IncrementCounter("mcp_prompt_generation_total")
	monitor.IncrementCounter(fmt.Sprintf("mcp_prompt_generation_%s", promptName))
}

func IncrementAPICall(endpoint string) {
	monitor := GetDefault()
	monitor.IncrementCounter("api_calls_total")
	monitor.IncrementCounter(fmt.Sprintf("api_calls_%s", endpoint))
}

func IncrementAPIError(endpoint string, statusCode int) {
	monitor := GetDefault()
	monitor.IncrementCounter("api_errors_total")
	monitor.IncrementCounter(fmt.Sprintf("api_errors_%s_%d", endpoint, statusCode))
}

func RecordAPILatency(endpoint string, duration time.Duration) {
	monitor := GetDefault()
	monitor.RecordLatency(fmt.Sprintf("api_latency_%s", endpoint), duration)
}

func SetMCPConnections(count int) {
	monitor := GetDefault()
	monitor.SetGauge("mcp_active_connections", float64(count))
}

func getVersion() string {
	if version := os.Getenv("TASKMAN_VERSION"); version != "" {
		return version
	}
	return "1.0.0"
}