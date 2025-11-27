package websocket

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/felixgateru/multiprotocol-benchmark/pkg"
	"github.com/gorilla/websocket"
)

type ClientConfig struct {
	ServerURL string
	Path      string
	Headers   map[string]string
}

type PubSubConfig struct {
	MessageCount int
	MessageSize  int
	Delay        time.Duration
	Timeout      time.Duration
	TLSConfig    *tls.Config
}

type MessageStats struct {
	mu              sync.Mutex
	publishedCount  int
	receivedCount   int
	publishedMsgIDs map[int]bool
	receivedMsgIDs  map[int]bool
	startTime       time.Time
	endTime         time.Time
}

func NewMessageStats() *MessageStats {
	return &MessageStats{
		publishedMsgIDs: make(map[int]bool),
		receivedMsgIDs:  make(map[int]bool),
		startTime:       time.Now(),
	}
}

func (ms *MessageStats) IncrementPublished(msgID int) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.publishedCount++
	ms.publishedMsgIDs[msgID] = true
}

func (ms *MessageStats) IncrementReceived(msgID int) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.receivedCount++
	ms.receivedMsgIDs[msgID] = true
}

func (ms *MessageStats) GetCounts() (published, received int) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.publishedCount, ms.receivedCount
}

func (ms *MessageStats) Finalize() {
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

type PathStats struct {
	mu              sync.Mutex
	publishedCount  int
	receivedCount   int
	publishedMsgIDs map[string]bool // clientID-msgID
	receivedMsgIDs  map[string]bool // clientID-msgID
}

func NewPathStats() *PathStats {
	return &PathStats{
		publishedMsgIDs: make(map[string]bool),
		receivedMsgIDs:  make(map[string]bool),
	}
}

func (ps *PathStats) IncrementPublished(clientID string, msgID int) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.publishedCount++
	ps.publishedMsgIDs[fmt.Sprintf("%s-%d", clientID, msgID)] = true
}

func (ps *PathStats) IncrementReceived(clientID string, msgID int) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.receivedCount++
	ps.receivedMsgIDs[fmt.Sprintf("%s-%d", clientID, msgID)] = true
}

func (ps *PathStats) GetCounts() (published, received int) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return ps.publishedCount, ps.receivedCount
}

type ClientPathConfig struct {
	ClientID string
	Config   ClientConfig
}

// RunPubSub runs a single publisher and subscriber for WebSocket
func RunPubSub(clientConfig ClientConfig, pubSubConfig PubSubConfig, logger *slog.Logger) (published, received int, err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stats := NewMessageStats()

	var wg sync.WaitGroup
	errChan := make(chan error, 2)
	readyChan := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := subscribeMessages(ctx, clientConfig, pubSubConfig, stats, readyChan, logger); err != nil {
			errChan <- fmt.Errorf("subscriber error: %w", err)
		}
	}()

	select {
	case <-readyChan:
		logger.Debug("Subscriber ready, starting publisher")
	case <-time.After(10 * time.Second):
		published, received = stats.GetCounts()
		return published, received, fmt.Errorf("timeout waiting for subscriber to be ready")
	case <-ctx.Done():
		published, received = stats.GetCounts()
		return published, received, ctx.Err()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := publishMessages(ctx, clientConfig, pubSubConfig, stats, logger); err != nil {
			errChan <- fmt.Errorf("publisher error: %w", err)
		}
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(pubSubConfig.Timeout):
		logger.Debug("Timeout reached")
	case <-ctx.Done():
		logger.Debug("Context cancelled")
	}

	stats.Finalize()

	select {
	case err := <-errChan:
		published, received = stats.GetCounts()
		return published, received, err
	default:
	}

	// Print final statistics with separated publish/subscribe counts
	published, received = stats.GetCounts()
	duration := stats.GetDuration()

	if pubSubConfig.MessageCount == 0 {
		logger.Info("=== WebSocket Final Statistics ===")
		logger.Info("No messages configured")
		return published, received, nil
	}

	logger.Info("=== WebSocket Final Statistics ===")
	logger.Info("Publishing Statistics:",
		"published", fmt.Sprintf("%d/%d", published, pubSubConfig.MessageCount),
		"success_rate", fmt.Sprintf("%.2f%%", float64(published)/float64(pubSubConfig.MessageCount)*100))
	logger.Info("Subscription Statistics:",
		"received", fmt.Sprintf("%d/%d", received, pubSubConfig.MessageCount),
		"success_rate", fmt.Sprintf("%.2f%%", float64(received)/float64(pubSubConfig.MessageCount)*100))
	logger.Info("Overall:", "duration", duration)

	if published == pubSubConfig.MessageCount && received == pubSubConfig.MessageCount {
		logger.Info("Status: SUCCESS - All messages published and received")
		return published, received, nil
	}

	if published != pubSubConfig.MessageCount {
		return published, received, fmt.Errorf("incomplete: %d messages published, expected %d", published, pubSubConfig.MessageCount)
	}
	if received != pubSubConfig.MessageCount {
		return published, received, fmt.Errorf("incomplete: %d messages received, expected %d", received, pubSubConfig.MessageCount)
	}

	return published, received, nil
}

func publishMessages(ctx context.Context, clientConfig ClientConfig, pubSubConfig PubSubConfig, stats *MessageStats, logger *slog.Logger) error {
	u, err := url.Parse(clientConfig.ServerURL)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}
	u.Path = "api/ws/" + clientConfig.Path

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
		TLSClientConfig:  pubSubConfig.TLSConfig,
	}

	var httpHeaders = http.Header{}
	for k, v := range clientConfig.Headers {
		httpHeaders.Set(k, v)
	}

	conn, _, err := dialer.Dial(u.String(), httpHeaders)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket server: %w", err)
	}
	defer conn.Close()

	logger.Debug(fmt.Sprintf("Publisher connected to %s", u.String()))

	for i := 1; i <= pubSubConfig.MessageCount; i++ {
		select {
		case <-ctx.Done():
			logger.Debug("Publisher: context cancelled")
			return ctx.Err()
		default:
		}

		// Create SenML message with variable size padding
		baseName := fmt.Sprintf("msg_%d", i)
		message, err := pkg.CreateSenMLMessage(baseName, i, pubSubConfig.MessageSize)
		if err != nil {
			logger.Warn("Failed to create SenML message", "message_id", i, "error", err)
			continue
		}

		err = conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			logger.Warn("Failed to publish message", "message_id", i, "error", err)
			return err
		}

		stats.IncrementPublished(i)
		logger.Debug(fmt.Sprintf("Published message %d, size: %d bytes", i, len(message)))

		if i < pubSubConfig.MessageCount {
			time.Sleep(pubSubConfig.Delay)
		}
	}

	logger.Debug("Publisher: all messages sent")
	return nil
}

func subscribeMessages(ctx context.Context, clientConfig ClientConfig, pubSubConfig PubSubConfig, stats *MessageStats, readyChan chan struct{}, logger *slog.Logger) error {
	u, err := url.Parse(clientConfig.ServerURL)
	if err != nil {
		close(readyChan)
		return fmt.Errorf("invalid server URL: %w", err)
	}
	u.Path = "api/ws/" + clientConfig.Path

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
		TLSClientConfig:  pubSubConfig.TLSConfig,
	}

	var httpHeaders = http.Header{}
	for k, v := range clientConfig.Headers {
		httpHeaders.Set(k, v)
	}

	conn, _, err := dialer.Dial(u.String(), httpHeaders)
	if err != nil {
		close(readyChan)
		return fmt.Errorf("failed to connect to WebSocket server: %w", err)
	}
	defer conn.Close()

	logger.Debug(fmt.Sprintf("Subscriber connected to %s", u.String()))

	close(readyChan)

	for i := 1; i <= pubSubConfig.MessageCount; i++ {
		select {
		case <-ctx.Done():
			logger.Debug("Subscriber: context cancelled")
			return ctx.Err()
		default:
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			logger.Warn(fmt.Sprintf("Failed to read message: %v", err))
			return err
		}

		// Parse SenML message
		records, err := pkg.ParseSenMLMessage(string(message))
		if err != nil {
			logger.Warn("Failed to parse SenML message", "error", err)
			continue
		}

		// Extract message number from SenML
		msgID, err := pkg.ExtractMessageNumber(records)
		if err != nil {
			logger.Warn("Failed to extract message number", "error", err)
			continue
		}

		stats.IncrementReceived(msgID)
		logger.Debug(fmt.Sprintf("Received message %d, size: %d bytes", msgID, len(message)))
	}

	logger.Debug("Subscriber: all messages received")
	return nil
}
