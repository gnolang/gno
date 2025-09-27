# `watchdog` - Service Health Monitoring

A simple watchdog timer implementation for monitoring service health and uptime. Tracks when a service was last seen alive and provides status reporting.

## Features

- **Health monitoring**: Track if a service is alive based on periodic updates
- **Configurable timeout**: Set custom duration for considering service down
- **Status reporting**: Get current status and uptime information
- **Downtime tracking**: Track when service went down

## Usage

```go
import (
    "time"
    "gno.land/p/nt/watchdog"
)

// Create watchdog with 30-second timeout
wd := &watchdog.Watchdog{
    Duration: 30 * time.Second,
}

// Service calls this periodically to indicate it's alive
wd.Alive()

// Check if service is currently alive
if wd.IsAlive() {
    // Service is healthy
} else {
    // Service is down or unresponsive
}

// Get status string
status := wd.Status() // "OK" or "KO"
```

## Monitoring Example

```go
// Health monitoring for a background service
type ServiceMonitor struct {
    watchdog *watchdog.Watchdog
}

func NewServiceMonitor() *ServiceMonitor {
    return &ServiceMonitor{
        watchdog: &watchdog.Watchdog{
            Duration: 60 * time.Second, // 1 minute timeout
        },
    }
}

func (sm *ServiceMonitor) Heartbeat() {
    sm.watchdog.Alive()
}

func (sm *ServiceMonitor) HealthCheck() string {
    if sm.watchdog.IsAlive() {
        upSince := sm.watchdog.UpSince()
        return "Service healthy, up since: " + upSince.String()
    }
    
    downSince := sm.watchdog.DownSince()
    return "Service down since: " + downSince.String()
}
```

## API

```go
type Watchdog struct {
    Duration   time.Duration // Timeout duration for considering service down
    // private fields for tracking state
}

// Service lifecycle
func (w *Watchdog) Alive()            // Mark service as alive (call periodically)

// Status checking  
func (w Watchdog) IsAlive() bool      // Check if service is currently alive
func (w Watchdog) Status() string     // Get status string ("OK" or "KO")

// Timing information
func (w Watchdog) UpSince() time.Time   // When service came back up
func (w Watchdog) DownSince() time.Time // When service went down
```

## Multi-Service Monitoring

```go
type MultiServiceMonitor struct {
    services map[string]*watchdog.Watchdog
}

func NewMultiServiceMonitor() *MultiServiceMonitor {
    return &MultiServiceMonitor{
        services: make(map[string]*watchdog.Watchdog),
    }
}

func (msm *MultiServiceMonitor) RegisterService(name string, timeout time.Duration) {
    msm.services[name] = &watchdog.Watchdog{
        Duration: timeout,
    }
}

func (msm *MultiServiceMonitor) Heartbeat(serviceName string) {
    if wd, exists := msm.services[serviceName]; exists {
        wd.Alive()
    }
}

func (msm *MultiServiceMonitor) GetSystemStatus() map[string]string {
    status := make(map[string]string)
    for name, wd := range msm.services {
        status[name] = wd.Status()
    }
    return status
}
```

## Integration with Gno Contracts

```go
var serviceWatchdog *watchdog.Watchdog

func init() {
    serviceWatchdog = &watchdog.Watchdog{
        Duration: 5 * time.Minute,
    }
}

// Call this from your service operations
func ProcessData() {
    defer serviceWatchdog.Alive() // Mark alive after processing
    
    // Process data...
}

// Render health status
func Render(path string) string {
    if path == "health" {
        if serviceWatchdog.IsAlive() {
            return "Service: " + serviceWatchdog.Status() + 
                   ", Up since: " + serviceWatchdog.UpSince().String()
        }
        return "Service: " + serviceWatchdog.Status() + 
               ", Down since: " + serviceWatchdog.DownSince().String()
    }
    return "Unknown path"
}
```

## Use Cases

- **Background job monitoring**: Track if scheduled tasks are running
- **API health checks**: Monitor if external services are responding
- **Service discovery**: Determine which services are currently available
- **Automated failover**: Trigger backup systems when primary service is down
- **Performance monitoring**: Track service reliability and uptime

This package provides a simple but effective way to monitor service health and implement basic reliability patterns in Gno applications.
