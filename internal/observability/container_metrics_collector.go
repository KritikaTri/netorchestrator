package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"netorchestrator/internal/models"
)

// ContainerStats represents Docker/Podman container statistics
type ContainerStats struct {
	ContainerID   string
	ContainerName string
	CPUPercent    float64
	MemoryUsageMB float64
	MemoryLimitMB float64
	MemoryPercent float64
	NetworkRxMB   float64
	NetworkTxMB   float64
	BlockReadMB   float64
	BlockWriteMB  float64
	PIDs          int
}

// ContainerMetricsCollector collects real-time metrics from Docker/Podman containers
type ContainerMetricsCollector struct {
	db              *gorm.DB
	logger          *zap.Logger
	containerEngine string // "docker" or "podman"
	ticker          *time.Ticker
	stopChan        chan struct{}
}

// NewContainerMetricsCollector creates a new container metrics collector
func NewContainerMetricsCollector(db *gorm.DB, logger *zap.Logger) *ContainerMetricsCollector {
	// Detect container engine (prefer docker, fallback to podman)
	engine := "docker"
	if err := exec.Command("docker", "ps").Run(); err != nil {
		logger.Info("Docker not available, trying Podman")
		if err := exec.Command("podman", "ps").Run(); err != nil {
			logger.Warn("Neither Docker nor Podman is available")
		} else {
			engine = "podman"
		}
	}

	return &ContainerMetricsCollector{
		db:              db,
		logger:          logger,
		containerEngine: engine,
		stopChan:        make(chan struct{}),
	}
}

// Start begins collecting container metrics every 10 seconds
func (c *ContainerMetricsCollector) Start(ctx context.Context) {
	c.logger.Info("Starting container metrics collector",
		zap.String("engine", c.containerEngine),
		zap.Duration("interval", 10*time.Second))

	c.ticker = time.NewTicker(10 * time.Second)

	// Collect immediately on start
	go c.collectAndStore(ctx)

	// Then collect every 10 seconds
	go func() {
		for {
			select {
			case <-c.ticker.C:
				c.collectAndStore(ctx)
			case <-c.stopChan:
				c.logger.Info("Stopping container metrics collector")
				return
			case <-ctx.Done():
				c.logger.Info("Context cancelled, stopping container metrics collector")
				return
			}
		}
	}()
}

// Stop stops the metrics collector
func (c *ContainerMetricsCollector) Stop() {
	if c.ticker != nil {
		c.ticker.Stop()
	}
	close(c.stopChan)
}

// collectAndStore collects metrics from all containers and stores them in DB
func (c *ContainerMetricsCollector) collectAndStore(ctx context.Context) {
	startTime := time.Now()

	// Get all running containers
	containerStats, err := c.collectContainerStats()
	if err != nil {
		c.logger.Error("Failed to collect container stats", zap.Error(err))
		return
	}

	if len(containerStats) == 0 {
		c.logger.Debug("No running containers found")
		return
	}

	// Store metrics in database
	stored := 0
	for _, stats := range containerStats {
		if err := c.storeContainerMetrics(ctx, stats); err != nil {
			c.logger.Error("Failed to store container metrics",
				zap.String("container", stats.ContainerName),
				zap.Error(err))
		} else {
			stored++
		}
	}

	duration := time.Since(startTime)
	c.logger.Info("Container metrics collection completed",
		zap.Int("containers_found", len(containerStats)),
		zap.Int("metrics_stored", stored),
		zap.Duration("duration", duration))
}

// collectContainerStats collects statistics from all running containers
func (c *ContainerMetricsCollector) collectContainerStats() ([]ContainerStats, error) {
	// Get list of running containers
	containers, err := c.getRunningContainers()
	if err != nil {
		return nil, fmt.Errorf("failed to get running containers: %w", err)
	}

	var allStats []ContainerStats

	// Collect stats for each container
	for _, container := range containers {
		stats, err := c.getContainerStats(container.ID, container.Name)
		if err != nil {
			c.logger.Warn("Failed to get stats for container",
				zap.String("container", container.Name),
				zap.Error(err))
			continue
		}
		allStats = append(allStats, stats)
	}

	return allStats, nil
}

// ContainerInfo holds basic container information
type ContainerInfo struct {
	ID   string
	Name string
}

// getRunningContainers returns list of running containers with full container IDs
func (c *ContainerMetricsCollector) getRunningContainers() ([]ContainerInfo, error) {
	cmd := exec.Command(c.containerEngine, "ps", "--no-trunc", "--format", "{{.ID}}||{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var containers []ContainerInfo

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "||")
		if len(parts) == 2 {
			containers = append(containers, ContainerInfo{
				ID:   parts[0],
				Name: parts[1],
			})
		}
	}

	return containers, nil
}

// getContainerStats gets statistics for a specific container
func (c *ContainerMetricsCollector) getContainerStats(containerID, containerName string) (ContainerStats, error) {
	stats := ContainerStats{
		ContainerID:   containerID,
		ContainerName: containerName,
	}

	// Use docker/podman stats with --no-stream to get instant snapshot
	cmd := exec.Command(c.containerEngine, "stats", containerID, "--no-stream", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		return stats, fmt.Errorf("failed to get stats: %w", err)
	}

	// Parse the JSON output - Docker returns an array of stats objects
	var rawStatsArray []map[string]interface{}
	if err := json.Unmarshal(output, &rawStatsArray); err != nil {
		return stats, fmt.Errorf("failed to parse stats JSON: %w", err)
	}

	// Get the first (and should be only) stats object
	if len(rawStatsArray) == 0 {
		return stats, fmt.Errorf("no stats data returned for container %s", containerID)
	}

	rawStats := rawStatsArray[0]

	// Extract CPU percentage (try both old and new Docker format)
	if cpuPerc, ok := rawStats["CPUPerc"].(string); ok {
		stats.CPUPercent = parsePercent(cpuPerc)
	} else if cpuPerc, ok := rawStats["cpu_percent"].(string); ok {
		stats.CPUPercent = parsePercent(cpuPerc)
	}

	// Extract memory usage and limit (try both formats)
	if memUsage, ok := rawStats["MemUsage"].(string); ok {
		usage, limit := parseMemoryUsage(memUsage)
		stats.MemoryUsageMB = usage
		stats.MemoryLimitMB = limit
	} else if memUsage, ok := rawStats["mem_usage"].(string); ok {
		usage, limit := parseMemoryUsage(memUsage)
		stats.MemoryUsageMB = usage
		stats.MemoryLimitMB = limit
	}

	// Extract memory percentage (try both formats)
	if memPerc, ok := rawStats["MemPerc"].(string); ok {
		stats.MemoryPercent = parsePercent(memPerc)
	} else if memPerc, ok := rawStats["mem_percent"].(string); ok {
		stats.MemoryPercent = parsePercent(memPerc)
	}

	// Extract network I/O (try both formats)
	if netIO, ok := rawStats["NetIO"].(string); ok {
		rx, tx := parseNetworkIO(netIO)
		stats.NetworkRxMB = rx
		stats.NetworkTxMB = tx
	} else if netIO, ok := rawStats["net_io"].(string); ok {
		rx, tx := parseNetworkIO(netIO)
		stats.NetworkRxMB = rx
		stats.NetworkTxMB = tx
	}

	// Extract block I/O (try both formats)
	if blockIO, ok := rawStats["BlockIO"].(string); ok {
		read, write := parseBlockIO(blockIO)
		stats.BlockReadMB = read
		stats.BlockWriteMB = write
	} else if blockIO, ok := rawStats["block_io"].(string); ok {
		read, write := parseBlockIO(blockIO)
		stats.BlockReadMB = read
		stats.BlockWriteMB = write
	}

	// Extract PIDs (try both formats)
	if pids, ok := rawStats["PIDs"].(string); ok {
		stats.PIDs = parsePIDs(pids)
	} else if pidsFloat, ok := rawStats["PIDs"].(float64); ok {
		stats.PIDs = int(pidsFloat)
	} else if pids, ok := rawStats["pids"].(string); ok {
		stats.PIDs = parsePIDs(pids)
	} else if pidsFloat, ok := rawStats["pids"].(float64); ok {
		stats.PIDs = int(pidsFloat)
	}

	return stats, nil
}

// storeContainerMetrics stores container metrics in the database
func (c *ContainerMetricsCollector) storeContainerMetrics(ctx context.Context, stats ContainerStats) error {
	// Try to find the node associated with this container
	var node models.Node
	err := c.db.WithContext(ctx).
		Where("name = ?", stats.ContainerName).
		First(&node).Error

	// Determine entity type and ID
	entityType := "container"
	var entityID string

	if err == nil {
		// Found matching node
		entityID = node.ID.String()
		entityType = "node"
	} else {
		// No matching node, use container ID
		entityID = stats.ContainerID
	}

	// Insert metrics into database
	metrics := []models.Metric{
		{
			EntityType:  entityType,
			EntityID:    entityID,
			MetricName:  "cpu_percent",
			MetricValue: stats.CPUPercent,
			Unit:        "percent",
			Metadata:    models.MetricMetadata{"container_name": stats.ContainerName},
		},
		{
			EntityType:  entityType,
			EntityID:    entityID,
			MetricName:  "memory_usage_mb",
			MetricValue: stats.MemoryUsageMB,
			Unit:        "mb",
			Metadata:    models.MetricMetadata{"container_name": stats.ContainerName, "limit_mb": stats.MemoryLimitMB},
		},
		{
			EntityType:  entityType,
			EntityID:    entityID,
			MetricName:  "memory_percent",
			MetricValue: stats.MemoryPercent,
			Unit:        "percent",
			Metadata:    models.MetricMetadata{"container_name": stats.ContainerName},
		},
		{
			EntityType:  entityType,
			EntityID:    entityID,
			MetricName:  "network_rx_mb",
			MetricValue: stats.NetworkRxMB,
			Unit:        "mb",
			Metadata:    models.MetricMetadata{"container_name": stats.ContainerName},
		},
		{
			EntityType:  entityType,
			EntityID:    entityID,
			MetricName:  "network_tx_mb",
			MetricValue: stats.NetworkTxMB,
			Unit:        "mb",
			Metadata:    models.MetricMetadata{"container_name": stats.ContainerName},
		},
		{
			EntityType:  entityType,
			EntityID:    entityID,
			MetricName:  "block_read_mb",
			MetricValue: stats.BlockReadMB,
			Unit:        "mb",
			Metadata:    models.MetricMetadata{"container_name": stats.ContainerName},
		},
		{
			EntityType:  entityType,
			EntityID:    entityID,
			MetricName:  "block_write_mb",
			MetricValue: stats.BlockWriteMB,
			Unit:        "mb",
			Metadata:    models.MetricMetadata{"container_name": stats.ContainerName},
		},
		{
			EntityType:  entityType,
			EntityID:    entityID,
			MetricName:  "pids",
			MetricValue: float64(stats.PIDs),
			Unit:        "count",
			Metadata:    models.MetricMetadata{"container_name": stats.ContainerName},
		},
	}

	// Use UPSERT to update existing metrics or insert new ones
	// This ensures only latest values are kept (no historical data)
	for _, metric := range metrics {
		result := c.db.WithContext(ctx).
			Exec(`
				INSERT INTO metrics (entity_type, entity_id, metric_name, metric_value, unit, timestamp, metadata)
				VALUES (?, ?, ?, ?, ?, NOW(), ?)
				ON CONFLICT (entity_id, metric_name) 
				DO UPDATE SET 
					metric_value = EXCLUDED.metric_value,
					timestamp = EXCLUDED.timestamp,
					metadata = EXCLUDED.metadata
			`,
				metric.EntityType,
				metric.EntityID,
				metric.MetricName,
				metric.MetricValue,
				metric.Unit,
				metric.Metadata,
			)

		if result.Error != nil {
			return fmt.Errorf("failed to upsert metric %s: %w", metric.MetricName, result.Error)
		}
	}

	return nil
}

// Helper functions to parse Docker stats output

func parsePercent(s string) float64 {
	s = strings.TrimSuffix(s, "%")
	val, _ := strconv.ParseFloat(s, 64)
	return val
}

func parseMemoryUsage(s string) (usage float64, limit float64) {
	// Format: "123.4MiB / 2GiB" or "123.4MB / 2GB"
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return 0, 0
	}

	usage = parseBytes(strings.TrimSpace(parts[0]))
	limit = parseBytes(strings.TrimSpace(parts[1]))
	return
}

func parseNetworkIO(s string) (rx float64, tx float64) {
	// Format: "1.2MB / 3.4MB"
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return 0, 0
	}

	rx = parseBytes(strings.TrimSpace(parts[0]))
	tx = parseBytes(strings.TrimSpace(parts[1]))
	return
}

func parseBlockIO(s string) (read float64, write float64) {
	// Format: "1.2MB / 3.4MB"
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return 0, 0
	}

	read = parseBytes(strings.TrimSpace(parts[0]))
	write = parseBytes(strings.TrimSpace(parts[1]))
	return
}

func parseBytes(s string) float64 {
	// Handle different units: B, KB, MB, GB, KiB, MiB, GiB
	s = strings.TrimSpace(s)

	multiplier := 1.0
	if strings.HasSuffix(s, "GiB") {
		multiplier = 1024.0
		s = strings.TrimSuffix(s, "GiB")
	} else if strings.HasSuffix(s, "GB") {
		multiplier = 1000.0
		s = strings.TrimSuffix(s, "GB")
	} else if strings.HasSuffix(s, "MiB") {
		multiplier = 1.0
		s = strings.TrimSuffix(s, "MiB")
	} else if strings.HasSuffix(s, "MB") {
		multiplier = 1.0
		s = strings.TrimSuffix(s, "MB")
	} else if strings.HasSuffix(s, "KiB") {
		multiplier = 1.0 / 1024.0
		s = strings.TrimSuffix(s, "KiB")
	} else if strings.HasSuffix(s, "KB") {
		multiplier = 1.0 / 1000.0
		s = strings.TrimSuffix(s, "KB")
	} else if strings.HasSuffix(s, "B") {
		multiplier = 1.0 / (1024.0 * 1024.0)
		s = strings.TrimSuffix(s, "B")
	}

	val, _ := strconv.ParseFloat(s, 64)
	return val * multiplier // Return in MB
}

func parsePIDs(s string) int {
	val, _ := strconv.Atoi(s)
	return val
}
