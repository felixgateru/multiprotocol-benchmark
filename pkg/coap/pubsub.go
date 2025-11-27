package coap

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/felixgateru/multiprotocol-benchmark/pkg"
	piondtls "github.com/pion/dtls/v3"
	"github.com/plgd-dev/go-coap/v3/dtls"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/message/pool"
	"github.com/plgd-dev/go-coap/v3/udp/client"
)

type ClientConfig struct {
	Host string
	Port int
	Path string
	Opts []message.Option
}

type PubSubConfig struct {
	MessageCount int
	MessageSize  int
	Delay        time.Duration
	Timeout      time.Duration
	TLSConfig    *tls.Config
}

// MessageStats tracks the message statistics
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

type TopicStats struct {
	mu              sync.Mutex
	publishedCount  int
	receivedCount   int
	publishedMsgIDs map[string]bool
	receivedMsgIDs  map[string]bool
}

func NewTopicStats() *TopicStats {
	return &TopicStats{
		publishedMsgIDs: make(map[string]bool),
		receivedMsgIDs:  make(map[string]bool),
	}
}

func (ts *TopicStats) IncrementPublished(clientID string, msgID int) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.publishedCount++
	ts.publishedMsgIDs[fmt.Sprintf("%s-%d", clientID, msgID)] = true
}

func (ts *TopicStats) IncrementReceived(clientID string, msgID int) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.receivedCount++
	ts.receivedMsgIDs[fmt.Sprintf("%s-%d", clientID, msgID)] = true
}

func (ts *TopicStats) GetCounts() (published, received int) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.publishedCount, ts.receivedCount
}

type ClientPathConfig struct {
	ClientConfig ClientConfig
	Path         string
	ClientNum    int
}

func createCoapClient(config ClientConfig, pubSubConfig PubSubConfig, logger *slog.Logger) (*client.Conn, error) {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	dtlsCfg := &piondtls.Config{
		InsecureSkipVerify: true,
		RootCAs:            pubSubConfig.TLSConfig.RootCAs,
	}
	conn, err := dtls.Dial(addr, dtlsCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to dial CoAP server: %w", err)
	}

	logger.Debug("Connected to CoAP server", "addr", addr)
	return conn, nil
}

func publishMessages(ctx context.Context, conn *client.Conn, config ClientConfig, pubSubConfig PubSubConfig, stats *MessageStats, wg *sync.WaitGroup, logger *slog.Logger) {
	defer wg.Done()

	logger.Debug("Starting publisher", "messageCount", pubSubConfig.MessageCount, "delay", pubSubConfig.Delay)

	for i := 1; i <= pubSubConfig.MessageCount; i++ {
		select {
		case <-ctx.Done():
			logger.Debug("Publisher: context cancelled, stopping")
			return
		default:
			// Create SenML message with variable size padding
			baseName := fmt.Sprintf("msg_%d", i)
			payload, err := pkg.CreateSenMLMessage(baseName, i, pubSubConfig.MessageSize)
			if err != nil {
				logger.Warn("Failed to create SenML message", "message_id", i, "error", err)
				continue
			}

			resp, err := conn.Post(ctx, config.Path, message.AppJSON, bytes.NewReader([]byte(payload)), config.Opts...)
			if err != nil {
				logger.Warn("Failed to publish message", "id", i, "error", err)
				continue
			}

			if resp != nil && (resp.Code() == codes.Changed || resp.Code() == codes.Created || resp.Code() == codes.Content) {
				stats.IncrementPublished(i)
				logger.Debug("Published message", "id", i, "payload", payload, "response", resp.Code())
			} else if resp != nil {
				logger.Debug("Published message but got unexpected response", "id", i, "response", resp.Code())
			}

			if i < pubSubConfig.MessageCount {
				time.Sleep(pubSubConfig.Delay)
			}
		}
	}

	logger.Debug("Publisher: all messages sent")
}

func observeMessages(ctx context.Context, conn *client.Conn, config ClientConfig, pubSubConfig PubSubConfig, stats *MessageStats, readyChan chan<- struct{}, wg *sync.WaitGroup, logger *slog.Logger) {
	defer wg.Done()

	receivedChan := make(chan struct{}, pubSubConfig.MessageCount)

	obs, err := conn.Observe(ctx, config.Path, func(req *pool.Message) {
		body := req.Body()
		if body == nil {
			return
		}

		payload := make([]byte, 1024*10) // Increased buffer for larger SenML messages
		n, readErr := body.Read(payload)
		if readErr != nil && readErr.Error() != "EOF" {
			logger.Warn("Failed to read notification body", "error", readErr)
			return
		}
		payload = payload[:n]

		// Parse SenML message
		records, err := pkg.ParseSenMLMessage(string(payload))
		if err != nil {
			logger.Warn("Failed to parse SenML message", "error", err)
			return
		}

		// Extract message number from SenML
		msgID, err := pkg.ExtractMessageNumber(records)
		if err != nil {
			logger.Warn("Failed to extract message number", "error", err)
			return
		}

		stats.IncrementReceived(msgID)

		_, received := stats.GetCounts()
		logger.Debug("Received notification", "message_num", msgID, "received", received, "expected", pubSubConfig.MessageCount)

		if received >= pubSubConfig.MessageCount {
			select {
			case receivedChan <- struct{}{}:
			default:
			}
		}
	}, config.Opts...)

	if err != nil {
		logger.Warn("Failed to observe", "error", err)
		if readyChan != nil {
			close(readyChan)
		}
		return
	}
	defer obs.Cancel(ctx)

	logger.Debug("Observing path", "path", config.Path)

	if readyChan != nil {
		close(readyChan)
	}

	select {
	case <-receivedChan:
		logger.Debug("Observer: all messages received")
	case <-ctx.Done():
		logger.Debug("Observer: context cancelled")
	}
}

func RunPubSub(clientConfig ClientConfig, pubSubConfig PubSubConfig, logger slog.Logger) (published, received int, err error) {
	stats := NewMessageStats()

	pubConn, createErr := createCoapClient(clientConfig, pubSubConfig, &logger)
	if createErr != nil {
		return 0, 0, fmt.Errorf("failed to create publisher connection: %w", createErr)
	}
	defer pubConn.Close()

	obsConn, createErr := createCoapClient(clientConfig, pubSubConfig, &logger)
	if createErr != nil {
		return 0, 0, fmt.Errorf("failed to create observer connection: %w", createErr)
	}
	defer obsConn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	obsReady := make(chan struct{})

	wg.Add(1)
	go observeMessages(ctx, obsConn, clientConfig, pubSubConfig, stats, obsReady, &wg, &logger)

	logger.Debug("Waiting for observer to be ready...")
	select {
	case <-obsReady:
		logger.Debug("Observer is ready, starting publisher")
	case <-time.After(5 * time.Second):
		logger.Warn("Observer took too long to be ready, starting publisher anyway")
	case <-ctx.Done():
		published, received = stats.GetCounts()
		return published, received, fmt.Errorf("context cancelled while waiting for observer")
	}

	wg.Add(1)
	go publishMessages(ctx, pubConn, clientConfig, pubSubConfig, stats, &wg, &logger)

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

	published, received = stats.GetCounts()
	duration := stats.GetDuration()

	logger.Info("=== CoAP Final Statistics ===")
	logger.Info("Publishing Statistics:",
		"published", fmt.Sprintf("%d/%d", published, pubSubConfig.MessageCount),
		"success_rate", fmt.Sprintf("%.2f%%", float64(published)/float64(pubSubConfig.MessageCount)*100))
	logger.Info("Observation Statistics:",
		"received", fmt.Sprintf("%d/%d", received, pubSubConfig.MessageCount),
		"success_rate", fmt.Sprintf("%.2f%%", float64(received)/float64(pubSubConfig.MessageCount)*100))
	logger.Info("Overall:", "duration", duration)

	if published == pubSubConfig.MessageCount && received == pubSubConfig.MessageCount {
		logger.Info("Status: SUCCESS - All messages published and received")
		return published, received, nil
	} else if received < published {
		return published, received, fmt.Errorf("incomplete: %d messages published but only %d received", published, received)
	} else if published < pubSubConfig.MessageCount {
		return published, received, fmt.Errorf("incomplete: only %d/%d messages published", published, pubSubConfig.MessageCount)
	}

	return published, received, nil
}
