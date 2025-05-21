package middleware

import (
	"database/sql"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// DatabaseHealthChecker is middleware that periodically checks database connectivity
type DatabaseHealthChecker struct {
	db            *sql.DB
	healthStatus  bool
	lastChecked   time.Time
	checkInterval time.Duration
	mu            sync.RWMutex
}

// NewDatabaseHealthChecker creates a new database health checker middleware
func NewDatabaseHealthChecker(db *sql.DB, checkInterval time.Duration) *DatabaseHealthChecker {
	if checkInterval <= 0 {
		checkInterval = 30 * time.Second
	}

	checker := &DatabaseHealthChecker{
		db:            db,
		healthStatus:  true, // Assume healthy initially
		lastChecked:   time.Now(),
		checkInterval: checkInterval,
	}

	// Do initial health check
	checker.updateHealth()

	// Start background health check
	go checker.startBackgroundChecks()

	return checker
}

// startBackgroundChecks periodically checks database health
func (d *DatabaseHealthChecker) startBackgroundChecks() {
	ticker := time.NewTicker(d.checkInterval)
	defer ticker.Stop()

	for range ticker.C {
		d.updateHealth()
	}
}

// updateHealth performs a health check and updates status
func (d *DatabaseHealthChecker) updateHealth() {
	err := d.db.Ping()

	d.mu.Lock()
	defer d.mu.Unlock()

	d.lastChecked = time.Now()
	if err != nil {
		if d.healthStatus {
			log.Printf("Database health check failed: %v", err)
		}
		d.healthStatus = false
	} else {
		if !d.healthStatus {
			log.Printf("Database connection restored")
		}
		d.healthStatus = true
	}
}

// Middleware creates the gin middleware function
func (d *DatabaseHealthChecker) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		d.mu.RLock()
		healthy := d.healthStatus
		lastChecked := d.lastChecked
		d.mu.RUnlock()

		// If the last check was too long ago or unhealthy, do a fresh check
		if time.Since(lastChecked) > d.checkInterval || !healthy {
			d.updateHealth()
			d.mu.RLock()
			healthy = d.healthStatus
			d.mu.RUnlock()
		}

		if !healthy {
			log.Printf("Request aborted due to unhealthy database connection")
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "Database connection unavailable",
			})
			return
		}

		c.Next()
	}
}
