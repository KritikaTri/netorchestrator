package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ServiceRegistry manages microservice registration and discovery
type ServiceRegistry struct {
	services map[string]*ServiceInfo
	mu       sync.RWMutex
	logger   *zap.Logger
}

// ServiceInfo represents a registered service
type ServiceInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Address     string            `json:"address"`
	Port        int               `json:"port"`
	Health      string            `json:"health"`      // healthy, unhealthy, unknown
	LastSeen    time.Time         `json:"last_seen"`
	Metadata    map[string]string `json:"metadata"`
	Endpoints   []string          `json:"endpoints"`
	Tags        []string          `json:"tags"`
}

// HealthCheck represents a service health check configuration
type HealthCheck struct {
	Endpoint string        `json:"endpoint"`
	Interval time.Duration `json:"interval"`
	Timeout  time.Duration `json:"timeout"`
	Method   string        `json:"method"`
}

func NewServiceRegistry(logger *zap.Logger) *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string]*ServiceInfo),
		logger:   logger,
	}
}

// RegisterService registers a new service in the registry
func (sr *ServiceRegistry) RegisterService(service *ServiceInfo) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	service.ID = uuid.New().String()
	service.LastSeen = time.Now()
	service.Health = "unknown"

	sr.services[service.ID] = service

	sr.logger.Info("Service registered",
		zap.String("service_id", service.ID),
		zap.String("service_name", service.Name),
		zap.String("address", service.Address),
		zap.Int("port", service.Port))

	return nil
}

// DeregisterService removes a service from the registry
func (sr *ServiceRegistry) DeregisterService(serviceID string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	if _, exists := sr.services[serviceID]; !exists {
		return fmt.Errorf("service not found: %s", serviceID)
	}

	delete(sr.services, serviceID)
	sr.logger.Info("Service deregistered", zap.String("service_id", serviceID))
	return nil
}

// DiscoverServices returns all healthy services of a given type
func (sr *ServiceRegistry) DiscoverServices(serviceName string) []*ServiceInfo {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	var services []*ServiceInfo
	for _, service := range sr.services {
		if service.Name == serviceName && service.Health == "healthy" {
			services = append(services, service)
		}
	}

	return services
}

// GetAllServices returns all registered services with their status
func (sr *ServiceRegistry) GetAllServices() []*ServiceInfo {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	services := make([]*ServiceInfo, 0, len(sr.services))
	for _, service := range sr.services {
		services = append(services, service)
	}

	return services
}

// StartHealthChecks begins periodic health checking for all services
func (sr *ServiceRegistry) StartHealthChecks(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sr.performHealthChecks()
		}
	}
}

// performHealthChecks checks the health of all registered services
func (sr *ServiceRegistry) performHealthChecks() {
	sr.mu.Lock()
	services := make([]*ServiceInfo, 0, len(sr.services))
	for _, service := range sr.services {
		services = append(services, service)
	}
	sr.mu.Unlock()

	for _, service := range services {
		go sr.checkServiceHealth(service)
	}
}

// checkServiceHealth performs a health check on a single service
func (sr *ServiceRegistry) checkServiceHealth(service *ServiceInfo) {
	healthURL := fmt.Sprintf("http://%s:%d/health", service.Address, service.Port)
	
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(healthURL)
	if err != nil {
		sr.updateServiceHealth(service.ID, "unhealthy")
		sr.logger.Warn("Health check failed",
			zap.String("service_id", service.ID),
			zap.String("service_name", service.Name),
			zap.Error(err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		sr.updateServiceHealth(service.ID, "healthy")
	} else {
		sr.updateServiceHealth(service.ID, "unhealthy")
		sr.logger.Warn("Health check returned non-200",
			zap.String("service_id", service.ID),
			zap.String("service_name", service.Name),
			zap.Int("status_code", resp.StatusCode))
	}
}

// updateServiceHealth updates the health status of a service
func (sr *ServiceRegistry) updateServiceHealth(serviceID, health string) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	if service, exists := sr.services[serviceID]; exists {
		service.Health = health
		service.LastSeen = time.Now()
	}
}

// GetServiceStats returns statistics about the service registry
func (sr *ServiceRegistry) GetServiceStats() map[string]interface{} {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	stats := map[string]interface{}{
		"total_services":     len(sr.services),
		"healthy_services":   0,
		"unhealthy_services": 0,
		"unknown_services":   0,
		"services_by_type":   make(map[string]int),
	}

	servicesByType := make(map[string]int)
	
	for _, service := range sr.services {
		switch service.Health {
		case "healthy":
			stats["healthy_services"] = stats["healthy_services"].(int) + 1
		case "unhealthy":
			stats["unhealthy_services"] = stats["unhealthy_services"].(int) + 1
		default:
			stats["unknown_services"] = stats["unknown_services"].(int) + 1
		}
		
		servicesByType[service.Name]++
	}
	
	stats["services_by_type"] = servicesByType
	return stats
}

// LoadBalancer provides simple round-robin load balancing
type LoadBalancer struct {
	services []*ServiceInfo
	current  int
	mu       sync.Mutex
}

// NewLoadBalancer creates a new load balancer for a service type
func (sr *ServiceRegistry) NewLoadBalancer(serviceName string) *LoadBalancer {
	services := sr.DiscoverServices(serviceName)
	return &LoadBalancer{
		services: services,
		current:  0,
	}
}

// NextService returns the next service in round-robin fashion
func (lb *LoadBalancer) NextService() *ServiceInfo {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if len(lb.services) == 0 {
		return nil
	}

	service := lb.services[lb.current]
	lb.current = (lb.current + 1) % len(lb.services)
	return service
}

// ServiceRegistryHandler provides HTTP endpoints for service registry operations
type ServiceRegistryHandler struct {
	registry *ServiceRegistry
}

func NewServiceRegistryHandler(registry *ServiceRegistry) *ServiceRegistryHandler {
	return &ServiceRegistryHandler{registry: registry}
}

// HandleServiceRegistration handles service registration requests
func (h *ServiceRegistryHandler) HandleServiceRegistration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var service ServiceInfo
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.registry.RegisterService(&service); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"service_id": service.ID,
		"status":     "registered",
	})
}

// HandleServiceDiscovery handles service discovery requests
func (h *ServiceRegistryHandler) HandleServiceDiscovery(w http.ResponseWriter, r *http.Request) {
	serviceName := r.URL.Query().Get("service")
	if serviceName == "" {
		// Return all services
		services := h.registry.GetAllServices()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(services)
		return
	}

	// Return specific service type
	services := h.registry.DiscoverServices(serviceName)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

// HandleRegistryStats handles registry statistics requests
func (h *ServiceRegistryHandler) HandleRegistryStats(w http.ResponseWriter, r *http.Request) {
	stats := h.registry.GetServiceStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}