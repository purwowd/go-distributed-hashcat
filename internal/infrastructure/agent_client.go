package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// CircuitState represents the circuit breaker state
type CircuitState int

const (
	Closed CircuitState = iota
	Open
	HalfOpen
)

// CircuitBreaker implements circuit breaker pattern for agent communication
type CircuitBreaker struct {
	mu           sync.RWMutex
	state        CircuitState
	failures     int
	lastFailTime time.Time
	timeout      time.Duration
	maxFailures  int
}

func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:       Closed,
		maxFailures: maxFailures,
		timeout:     timeout,
	}
}

func (cb *CircuitBreaker) CanExecute() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case Closed:
		return true
	case Open:
		// Check if we should transition to half-open
		if time.Since(cb.lastFailTime) > cb.timeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			if cb.state == Open {
				cb.state = HalfOpen
			}
			cb.mu.Unlock()
			cb.mu.RLock()
			return cb.state == HalfOpen
		}
		return false
	case HalfOpen:
		return true
	default:
		return false
	}
}

func (cb *CircuitBreaker) OnSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	cb.state = Closed
}

func (cb *CircuitBreaker) OnFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailTime = time.Now()

	if cb.failures >= cb.maxFailures {
		cb.state = Open
	}
}

// RetryConfig configures retry behavior with exponential backoff
type RetryConfig struct {
	MaxRetries    int           `json:"max_retries"`
	BaseDelay     time.Duration `json:"base_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
	JitterEnabled bool          `json:"jitter_enabled"`
}

// AgentClient handles resilient communication with the server
type AgentClient struct {
	httpClient     *http.Client
	serverURL      string
	agentID        uuid.UUID
	circuitBreaker *CircuitBreaker
	retryConfig    RetryConfig
	metrics        *ClientMetrics
}

type ClientMetrics struct {
	mu                  sync.RWMutex
	TotalRequests       int64         `json:"total_requests"`
	SuccessfulRequests  int64         `json:"successful_requests"`
	FailedRequests      int64         `json:"failed_requests"`
	CircuitOpenEvents   int64         `json:"circuit_open_events"`
	RetryAttempts       int64         `json:"retry_attempts"`
	LastRequestTime     time.Time     `json:"last_request_time"`
	AverageResponseTime time.Duration `json:"average_response_time"`
}

func NewAgentClient(serverURL string, agentID uuid.UUID) *AgentClient {
	return &AgentClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		serverURL:      serverURL,
		agentID:        agentID,
		circuitBreaker: NewCircuitBreaker(3, 30*time.Second), // 3 failures, 30s timeout
		retryConfig: RetryConfig{
			MaxRetries:    3,
			BaseDelay:     1 * time.Second,
			MaxDelay:      30 * time.Second,
			BackoffFactor: 2.0,
			JitterEnabled: true,
		},
		metrics: &ClientMetrics{},
	}
}

// SendHeartbeatWithRetry sends heartbeat with circuit breaker and retry logic
func (c *AgentClient) SendHeartbeatWithRetry(ctx context.Context) error {
	return c.executeWithRetry(ctx, "heartbeat", func() error {
		return c.sendHeartbeat(ctx)
	})
}

// UpdateStatusWithRetry updates agent status with retry logic
func (c *AgentClient) UpdateStatusWithRetry(ctx context.Context, status string) error {
	return c.executeWithRetry(ctx, "update_status", func() error {
		return c.updateStatus(ctx, status)
	})
}

func (c *AgentClient) executeWithRetry(ctx context.Context, operation string, fn func() error) error {
	c.metrics.mu.Lock()
	c.metrics.TotalRequests++
	c.metrics.LastRequestTime = time.Now()
	c.metrics.mu.Unlock()

	// Check circuit breaker
	if !c.circuitBreaker.CanExecute() {
		c.metrics.mu.Lock()
		c.metrics.FailedRequests++
		c.metrics.CircuitOpenEvents++
		c.metrics.mu.Unlock()
		return fmt.Errorf("circuit breaker is open for operation: %s", operation)
	}

	start := time.Now()
	var lastErr error

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate delay with exponential backoff
			delay := c.calculateBackoffDelay(attempt)

			c.metrics.mu.Lock()
			c.metrics.RetryAttempts++
			c.metrics.mu.Unlock()

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		err := fn()
		if err == nil {
			// Success
			c.circuitBreaker.OnSuccess()

			duration := time.Since(start)
			c.metrics.mu.Lock()
			c.metrics.SuccessfulRequests++
			// Update average response time (simple moving average)
			if c.metrics.AverageResponseTime == 0 {
				c.metrics.AverageResponseTime = duration
			} else {
				c.metrics.AverageResponseTime = (c.metrics.AverageResponseTime + duration) / 2
			}
			c.metrics.mu.Unlock()

			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !c.isRetryableError(err) {
			break
		}
	}

	// All attempts failed
	c.circuitBreaker.OnFailure()

	c.metrics.mu.Lock()
	c.metrics.FailedRequests++
	c.metrics.mu.Unlock()

	return fmt.Errorf("operation %s failed after %d attempts: %w",
		operation, c.retryConfig.MaxRetries+1, lastErr)
}

func (c *AgentClient) calculateBackoffDelay(attempt int) time.Duration {
	delay := float64(c.retryConfig.BaseDelay) * math.Pow(c.retryConfig.BackoffFactor, float64(attempt))

	if delay > float64(c.retryConfig.MaxDelay) {
		delay = float64(c.retryConfig.MaxDelay)
	}

	// Add jitter to prevent thundering herd
	if c.retryConfig.JitterEnabled {
		jitter := delay * 0.1 * (2*rand.Float64() - 1) // ±10% jitter
		delay += jitter
	}

	return time.Duration(delay)
}

func (c *AgentClient) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Network errors are retryable
	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout() || netErr.Temporary()
	}

	// Context errors
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}

	// HTTP status code errors
	if httpErr, ok := err.(*url.Error); ok {
		return c.isRetryableError(httpErr.Err)
	}

	// Connection errors
	errStr := err.Error()
	retryableErrors := []string{
		"connection refused",
		"connection reset",
		"no such host",
		"network is unreachable",
		"temporary failure",
	}

	for _, retryableErr := range retryableErrors {
		if strings.Contains(strings.ToLower(errStr), retryableErr) {
			return true
		}
	}

	return false
}

func (c *AgentClient) sendHeartbeat(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/v1/agents/%s/heartbeat", c.serverURL, c.agentID.String())

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("heartbeat failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *AgentClient) updateStatus(ctx context.Context, status string) error {
	req := struct {
		Status string `json:"status"`
	}{Status: status}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/agents/%s/status", c.serverURL, c.agentID.String())
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status update failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetMetrics returns current client metrics
func (c *AgentClient) GetMetrics() ClientMetrics {
	c.metrics.mu.RLock()
	defer c.metrics.mu.RUnlock()

	// Return a copy without the mutex to avoid copying lock value
	return ClientMetrics{
		TotalRequests:       c.metrics.TotalRequests,
		SuccessfulRequests:  c.metrics.SuccessfulRequests,
		FailedRequests:      c.metrics.FailedRequests,
		CircuitOpenEvents:   c.metrics.CircuitOpenEvents,
		RetryAttempts:       c.metrics.RetryAttempts,
		LastRequestTime:     c.metrics.LastRequestTime,
		AverageResponseTime: c.metrics.AverageResponseTime,
	}
}

// HealthCheck performs a health check to the server
func (c *AgentClient) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", c.serverURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	return nil
}
