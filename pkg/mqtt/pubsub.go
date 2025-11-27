package mqtt

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/felixgateru/multiprotocol-benchmark/pkg"
)

type ClientConfig struct {
	Broker   string
	ClientID string
	Username string
	Password string
	Topic    string
}

type PubSubConfig struct {
	MessageCount int
	MessageSize  int
	Delay        time.Duration
	Timeout      time.Duration
	QoS          byte
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

type TopicStats struct {
	mu              sync.Mutex
	publishedCount  int
	receivedCount   int
	publishedMsgIDs map[string]bool // clientID-msgID
	receivedMsgIDs  map[string]bool // clientID-msgID
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

type ClientTopicConfig struct {
	ClientConfig ClientConfig
	Topic        string
	ClientNum    int
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

func createMQTTClient(config ClientConfig, pubSubConfig PubSubConfig, logger *slog.Logger) (mqtt.Client, error) {
	var client mqtt.Client

	retryConfig := pkg.RetryConfig{
		Timeout:    pubSubConfig.Timeout,
		Logger:     logger,
	}

	err := pkg.RetryWithExponentialBackoff(retryConfig, func() error {
		opts := mqtt.NewClientOptions()
		opts.AddBroker(config.Broker)
		opts.SetClientID(config.ClientID)
		opts.SetUsername(config.Username)
		opts.SetPassword(config.Password)
		opts.SetCleanSession(true)
		opts.SetAutoReconnect(true)

		// Set TLS configuration if provided
		if pubSubConfig.TLSConfig != nil {
			opts.SetTLSConfig(pubSubConfig.TLSConfig)
		}

		opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
			logger.Warn("MQTT connection lost", "error", err)
		})
		opts.SetOnConnectHandler(func(client mqtt.Client) {
			logger.Debug("MQTT connected to broker", "broker", config.Broker)
		})

		client = mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			return fmt.Errorf("failed to connect to broker: %w", token.Error())
		}

		return nil
	}, fmt.Sprintf("MQTT connection to %s", config.Broker))

	if err != nil {
		return nil, err
	}

	return client, nil
}

func publishMessages(ctx context.Context, client mqtt.Client, config ClientConfig, pubSubConfig PubSubConfig, stats *MessageStats, wg *sync.WaitGroup, logger *slog.Logger) {
	defer wg.Done()

	logger.Debug("MQTT publisher starting", "message_count", pubSubConfig.MessageCount, "delay", pubSubConfig.Delay)

	for i := 1; i <= pubSubConfig.MessageCount; i++ {
		select {
		case <-ctx.Done():
			logger.Debug("MQTT publisher context cancelled, stopping")
			return
		default:
			// Create SenML message with variable size padding
			baseName := fmt.Sprintf("msg_%d", i)
			message, err := pkg.CreateSenMLMessage(baseName, i, pubSubConfig.MessageSize)
			if err != nil {
				logger.Warn("MQTT failed to create SenML message", "message_num", i, "error", err)
				continue
			}

			token := client.Publish(config.Topic, pubSubConfig.QoS, false, message)

			if token.Wait() && token.Error() != nil {
				logger.Warn("MQTT failed to publish message", "message_num", i, "error", token.Error())
				continue
			}

			stats.IncrementPublished(i)
			logger.Debug("MQTT published message", "message_num", i, "size", len(message))

			if i < pubSubConfig.MessageCount {
				time.Sleep(pubSubConfig.Delay)
			}
		}
	}

	logger.Debug("MQTT publisher completed, all messages sent")
}

// subscribeMessages subscribes to messages from the MQTT broker
func subscribeMessages(ctx context.Context, client mqtt.Client, config ClientConfig, pubSubConfig PubSubConfig, stats *MessageStats, readyChan chan<- struct{}, wg *sync.WaitGroup, logger *slog.Logger) {
	defer wg.Done()

	receivedChan := make(chan struct{}, pubSubConfig.MessageCount)

	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		// Parse SenML message
		records, err := pkg.ParseSenMLMessage(string(msg.Payload()))
		if err != nil {
			logger.Warn("MQTT failed to parse SenML message", "error", err)
			return
		}

		// Extract message number from SenML
		msgID, err := pkg.ExtractMessageNumber(records)
		if err != nil {
			logger.Warn("MQTT failed to extract message number", "error", err)
			return
		}

		stats.IncrementReceived(msgID)

		_, received := stats.GetCounts()
		logger.Debug("MQTT received message", "message_num", msgID, "received_count", received, "expected", pubSubConfig.MessageCount)

		if received >= pubSubConfig.MessageCount {
			receivedChan <- struct{}{}
		}
	}

	token := client.Subscribe(config.Topic, pubSubConfig.QoS, messageHandler)
	if token.Wait() && token.Error() != nil {
		logger.Error("MQTT failed to subscribe", "topic", config.Topic, "error", token.Error())
		if readyChan != nil {
			close(readyChan)
		}
		return
	}

	logger.Debug("MQTT subscribed to topic", "topic", config.Topic)

	// Signal that subscriber is ready
	if readyChan != nil {
		close(readyChan)
	}

	select {
	case <-receivedChan:
		logger.Debug("MQTT subscriber completed, all messages received")
	case <-ctx.Done():
		logger.Debug("MQTT subscriber context cancelled")
	}
}

func RunPubSub(clientConfig ClientConfig, pubSubConfig PubSubConfig, logger *slog.Logger) error {
	stats := NewMessageStats()

	pubClientConfig := clientConfig
	pubClientConfig.ClientID = fmt.Sprintf("%s-pub", clientConfig.ClientID)
	pubClient, err := createMQTTClient(pubClientConfig, pubSubConfig, logger)
	if err != nil {
		return fmt.Errorf("failed to create publisher client: %w", err)
	}
	defer pubClient.Disconnect(250)

	subClientConfig := clientConfig
	subClientConfig.ClientID = fmt.Sprintf("%s-sub", clientConfig.ClientID)
	subClient, err := createMQTTClient(subClientConfig, pubSubConfig, logger)
	if err != nil {
		return fmt.Errorf("failed to create subscriber client: %w", err)
	}
	defer subClient.Disconnect(250)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	subReady := make(chan struct{})

	wg.Add(1)
	go subscribeMessages(ctx, subClient, subClientConfig, pubSubConfig, stats, subReady, &wg, logger)

	logger.Debug("Waiting for MQTT subscriber to be ready")
	select {
	case <-subReady:
		logger.Debug("MQTT subscriber ready, starting publisher")
	case <-time.After(5 * time.Second):
		logger.Warn("MQTT subscriber took too long to be ready, starting publisher anyway")
	case <-ctx.Done():
		return fmt.Errorf("context cancelled while waiting for subscriber")
	}

	wg.Add(1)
	go publishMessages(ctx, pubClient, pubClientConfig, pubSubConfig, stats, &wg, logger)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Debug("MQTT pub/sub operations completed")
	case <-ctx.Done():
		logger.Warn("MQTT pub/sub operations timed out")
	}

	stats.SetEndTime()

	// Print final statistics with separated publish/subscribe counts
	published, received := stats.GetCounts()
	duration := stats.GetDuration()

	var pubSuccessRate, subSuccessRate float64
	if pubSubConfig.MessageCount > 0 {
		pubSuccessRate = float64(published) / float64(pubSubConfig.MessageCount) * 100
		subSuccessRate = float64(received) / float64(pubSubConfig.MessageCount) * 100
	}
	logger.Info("=== MQTT Final Statistics ===")
	logger.Info("Publishing Statistics:",
		"published", fmt.Sprintf("%d/%d", published, pubSubConfig.MessageCount),
		"success_rate", fmt.Sprintf("%.2f%%", pubSuccessRate))
	logger.Info("Subscription Statistics:",
		"received", fmt.Sprintf("%d/%d", received, pubSubConfig.MessageCount),
		"success_rate", fmt.Sprintf("%.2f%%", subSuccessRate))
	logger.Info("Overall:", "duration", duration)

	if published == pubSubConfig.MessageCount && received == pubSubConfig.MessageCount {
		logger.Info("MQTT Status: SUCCESS - All messages published and received")
		return nil
	} else if received < published {
		return fmt.Errorf("incomplete: %d messages published but only %d received", published, received)
	} else if published < pubSubConfig.MessageCount {
		return fmt.Errorf("incomplete: only %d/%d messages published", published, pubSubConfig.MessageCount)
	}

	return nil
}
