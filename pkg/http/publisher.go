package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/felixgateru/multiprotocol-benchmark/pkg"
)

type ClientConfig struct {
	BaseURL  string
	Endpoint string
	Headers  map[string]string
}

type PublishConfig struct {
	MessageCount int
	MessageSize  int
	Delay        time.Duration
	Timeout      time.Duration
	Method       string
	TLSConfig    *tls.Config
}

// MessageStats tracks the message statistics
type MessageStats struct {
	mu              sync.Mutex
	publishedCount  int
	successCount    int
	failureCount    int
	publishedMsgIDs map[int]bool
	startTime       time.Time
	endTime         time.Time
	responseTimes   []time.Duration
}

func NewMessageStats() *MessageStats {
	return &MessageStats{
		publishedMsgIDs: make(map[int]bool),
		startTime:       time.Now(),
		responseTimes:   make([]time.Duration, 0),
	}
}

func (ms *MessageStats) IncrementPublished(msgID int, success bool, responseTime time.Duration) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.publishedCount++
	ms.publishedMsgIDs[msgID] = true
	if success {
		ms.successCount++
	} else {
		ms.failureCount++
	}
	ms.responseTimes = append(ms.responseTimes, responseTime)
}

func (ms *MessageStats) GetCounts() (published, success, failure int) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.publishedCount, ms.successCount, ms.failureCount
}

func (ms *MessageStats) SetEndTime() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.endTime = time.Now()
}

func (ms *MessageStats) GetDuration() time.Duration {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if ms.endTime.IsZero() {
		return time.Since(ms.startTime)
	}
	return ms.endTime.Sub(ms.startTime)
}

func (ms *MessageStats) GetAverageResponseTime() time.Duration {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if len(ms.responseTimes) == 0 {
		return 0
	}
	var total time.Duration
	for _, rt := range ms.responseTimes {
		total += rt
	}
	return total / time.Duration(len(ms.responseTimes))
}

func (ms *MessageStats) GetMinMaxResponseTime() (min, max time.Duration) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if len(ms.responseTimes) == 0 {
		return 0, 0
	}
	min = ms.responseTimes[0]
	max = ms.responseTimes[0]
	for _, rt := range ms.responseTimes {
		if rt < min {
			min = rt
		}
		if rt > max {
			max = rt
		}
	}
	return min, max
}

type EndpointStats struct {
	mu              sync.Mutex
	publishedCount  int
	successCount    int
	failureCount    int
	publishedMsgIDs map[string]bool
	responseTimes   []time.Duration
}

func NewEndpointStats() *EndpointStats {
	return &EndpointStats{
		publishedMsgIDs: make(map[string]bool),
		responseTimes:   make([]time.Duration, 0),
	}
}

func (es *EndpointStats) IncrementPublished(clientID string, msgID int, success bool, responseTime time.Duration) {
	es.mu.Lock()
	defer es.mu.Unlock()
	es.publishedCount++
	es.publishedMsgIDs[fmt.Sprintf("%s-%d", clientID, msgID)] = true
	if success {
		es.successCount++
	} else {
		es.failureCount++
	}
	es.responseTimes = append(es.responseTimes, responseTime)
}

func (es *EndpointStats) GetCounts() (published, success, failure int) {
	es.mu.Lock()
	defer es.mu.Unlock()
	return es.publishedCount, es.successCount, es.failureCount
}

func (es *EndpointStats) GetAverageResponseTime() time.Duration {
	es.mu.Lock()
	defer es.mu.Unlock()
	if len(es.responseTimes) == 0 {
		return 0
	}
	var total time.Duration
	for _, rt := range es.responseTimes {
		total += rt
	}
	return total / time.Duration(len(es.responseTimes))
}

type ClientEndpointConfig struct {
	ClientConfig ClientConfig
	Endpoint     string
	ClientNum    int
}

func createHTTPClient(timeout time.Duration, tlsConfig *tls.Config) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	// Set TLS configuration if provided
	if tlsConfig != nil {
		transport.TLSClientConfig = tlsConfig
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

func publishMessages(ctx context.Context, client *http.Client, config ClientConfig, pubConfig PublishConfig, stats *MessageStats, wg *sync.WaitGroup, logger slog.Logger) {
	defer wg.Done()

	url := config.BaseURL + "/" + config.Endpoint
	logger.Debug("url", "url", url)
	logger.Debug("Starting publisher", "message_count", pubConfig.MessageCount, "url", url, "delay", pubConfig.Delay)

	retryConfig := pkg.RetryConfig{
		MaxRetries: pkg.MaxRetries,
		Timeout:    pubConfig.Timeout,
		Logger:     logger,
	}

	for i := 1; i <= pubConfig.MessageCount; i++ {
		select {
		case <-ctx.Done():
			logger.Debug("Publisher: context cancelled, stopping")
			return
		default:
			// Create SenML message with variable size padding
			baseName := fmt.Sprintf("msg_%d", i)
			payload, err := pkg.CreateSenMLMessage(baseName, i, pubConfig.MessageSize)
			if err != nil {
				logger.Warn("Failed to create SenML message", "message_id", i, "error", err)
				stats.IncrementPublished(i, false, 0)
				continue
			}

			startTime := time.Now()
			success := false

			// Retry logic for publishing
			err = pkg.RetryWithExponentialBackoff(retryConfig, func() error {
				req, reqErr := http.NewRequestWithContext(ctx, pubConfig.Method, url, bytes.NewBufferString(payload))
				if reqErr != nil {
					return fmt.Errorf("failed to create request: %w", reqErr)
				}

				req.Header.Set("Content-Type", "application/senml+json")
				for key, value := range config.Headers {
					req.Header.Set(key, value)
				}

				resp, doErr := client.Do(req)
				if doErr != nil {
					return fmt.Errorf("failed to send request: %w", doErr)
				}
				defer resp.Body.Close()

				body, _ := io.ReadAll(resp.Body)

				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					success = true
					return nil
				}

				return fmt.Errorf("got status code %d: %s", resp.StatusCode, string(body))
			}, fmt.Sprintf("HTTP publish message %d", i))

			responseTime := time.Since(startTime)

			if err != nil {
				logger.Warn("Failed to publish message after retries", "message_id", i, "error", err)
			} else {
				logger.Debug("Published message", "message_id", i, "response_time", responseTime)
			}

			stats.IncrementPublished(i, success, responseTime)

			if i < pubConfig.MessageCount {
				time.Sleep(pubConfig.Delay)
			}
		}
	}

	logger.Debug("Publisher: all messages sent")
}

func RunPublish(clientConfig ClientConfig, pubConfig PublishConfig, logger slog.Logger) error {
	stats := NewMessageStats()

	httpClient := createHTTPClient(10*time.Second, pubConfig.TLSConfig)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(1)
	go publishMessages(ctx, httpClient, clientConfig, pubConfig, stats, &wg, logger)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Debug("All operations completed successfully")
	case <-ctx.Done():
		logger.Debug("Operations timed out")
	}

	stats.SetEndTime()

	// Print final statistics with enhanced output
	published, success, failure := stats.GetCounts()
	duration := stats.GetDuration()
	avgResponseTime := stats.GetAverageResponseTime()
	minRT, maxRT := stats.GetMinMaxResponseTime()

	logger.Info("=== HTTP Final Statistics ===")
	logger.Info("Publishing Statistics:",
		"attempted", fmt.Sprintf("%d/%d", published, pubConfig.MessageCount),
		"successful", success,
		"failed", failure,
		"success_rate", fmt.Sprintf("%.2f%%", float64(success)/float64(published)*100))
	logger.Info("Performance Metrics:",
		"duration", duration,
		"throughput", fmt.Sprintf("%.2f msgs/sec", float64(published)/duration.Seconds()),
		"avg_response_time", avgResponseTime,
		"min_response_time", minRT,
		"max_response_time", maxRT)

	if published == pubConfig.MessageCount && success == pubConfig.MessageCount {
		logger.Info("Status: SUCCESS - All messages published successfully")
		return nil
	} else if failure > 0 {
		return fmt.Errorf("incomplete: %d messages published but %d failed", published, failure)
	} else if published < pubConfig.MessageCount {
		return fmt.Errorf("incomplete: only %d/%d messages published", published, pubConfig.MessageCount)
	}

	return nil
}
