package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"log/slog"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/0x6flab/namegenerator"
	smqlog "github.com/absmach/supermq/logger"
	"github.com/absmach/supermq/pkg/sdk"
	"github.com/felixgateru/multiprotocol-benchmark/pkg"
	pkgcoap "github.com/felixgateru/multiprotocol-benchmark/pkg/coap"
	pkghttp "github.com/felixgateru/multiprotocol-benchmark/pkg/http"
	pkgmqtt "github.com/felixgateru/multiprotocol-benchmark/pkg/mqtt"
	pkgws "github.com/felixgateru/multiprotocol-benchmark/pkg/websocket"
	"github.com/plgd-dev/go-coap/v3/message"
)

var nameGen = namegenerator.NewGenerator()

func main() {
	cfg := NewConfig()

	logger, err := smqlog.New(os.Stdout, cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to init logger: %s", err.Error())
	}

	var exitCode int
	defer smqlog.ExitWithError(&exitCode)

	// Load TLS configuration
	var tlsConfig *tls.Config
	if cfg.TLSVerify {
		tlsConfig, err = LoadSystemTLSConfig(cfg.CACertPath)
		if err != nil {
			logger.Error("Failed to load TLS configuration", "error", err)
			exitCode = 1
			return
		}
		if cfg.CACertPath != "" {
			logger.Info("Loaded custom CA certificate", "path", cfg.CACertPath)
		} else {
			logger.Info("Using system CA certificates")
		}
	}

	aggregator := NewResultsAggregator()

	var wg sync.WaitGroup
	protocolCount := 0

	if cfg.RunMQTT {
		protocolCount++
	}
	if cfg.RunCOAP {
		protocolCount++
	}
	if cfg.RunHTTP {
		protocolCount++
	}
	if cfg.RunWS {
		protocolCount++
	}

	if protocolCount == 0 {
		logger.Warn("No protocols enabled. Please enable at least one protocol (RUN_MQTT, RUN_COAP, RUN_HTTP, RUN_WS)")
		return
	}

	wg.Add(protocolCount)

	if cfg.RunMQTT {
		go func() {
			defer wg.Done()
			testMQTTPubSub(cfg, clientIDs, channelIDs, clientSecretMap, aggregator, logger, tlsConfig)
		}()
	}

	if cfg.RunCOAP {
		go func() {
			defer wg.Done()
			testCOAPPubSub(cfg, clientIDs, channelIDs, clientSecretMap, aggregator, logger, tlsConfig)
		}()
	}

	if cfg.RunHTTP {
		go func() {
			defer wg.Done()
			testHTTPPublish(cfg, clientIDs, channelIDs, clientSecretMap, aggregator, logger, tlsConfig)
		}()
	}

	if cfg.RunWS {
		go func() {
			defer wg.Done()
			testWSPubSub(cfg, clientIDs, channelIDs, clientSecretMap, aggregator, logger, tlsConfig)
		}()
	}

	wg.Wait()

	aggregator.PrintSummary(cfg.MQTTMessageSize, cfg.MQTTDelay)

	if cfg.SaveToFile {
		// Save results to file
		resultsFile := fmt.Sprintf("test_results_%s.json", time.Now().Format("20060102_150405"))
		err = aggregator.SaveToFile(resultsFile, cfg.MQTTMessageSize, cfg.MQTTDelay)
		if err != nil {
			logger.Error("Failed to save results to file", "error", err)
			exitCode = 1
			return
		}
		logger.Info("Results saved to file", "file", resultsFile)
	}

	// if createdResources && len(channelIDs) > 0 && len(clientIDs) > 0 {
	// 	logger.Info("Cleaning up provisioned resources")
	// 	if err := cleanUpProvision(context.Background(), channelIDs, clientIDs, cfg.DomainID, token.AccessToken, channelsSDK, clientsSDK, logger); err == nil {
	// 		logger.Info("Cleanup completed successfully")
	// 		return
	// 	}
	// 	logger.Error("Cleanup failed", "error", err)
	// }
}

func getIDS(objects any) []string {
	v := reflect.ValueOf(objects)
	if v.Kind() != reflect.Slice {
		panic("objects argument must be a slice")
	}
	ids := make([]string, v.Len())
	for i := 0; i < v.Len(); i++ {
		id := v.Index(i).FieldByName("ID").String()
		ids[i] = id
	}

	return ids
}

func testMQTTPubSub(cfg Config, clientIDs, channelIDs []string, clientSecretMap map[string]string, aggregator *ResultsAggregator, logger *slog.Logger, tlsConfig *tls.Config) {
	pubSubCfg := pkgmqtt.PubSubConfig{
		MessageCount: cfg.MQTTMessageCount,
		Delay:        cfg.MQTTDelay,
		Timeout:      cfg.MQTTTimeout,
		MessageSize:  cfg.MQTTMessageSize,
		QoS:          cfg.MQTTQoS,
		TLSConfig:    tlsConfig,
	}

	var wg sync.WaitGroup

	for _, channelID := range channelIDs {
		for _, clientID := range clientIDs {
			wg.Add(1)
			go func(chID, clID string) {
				defer wg.Done()

				topic := constructTopic(cfg.DomainID, chID, clID, "mqtt")
				clientCfg := pkgmqtt.ClientConfig{
					Broker:   cfg.MQTTBroker,
					ClientID: nameGen.Generate(),
					Username: clID,
					Password: clientSecretMap[clID],
					Topic:    topic,
				}

				logger.Debug("Starting MQTT Pub/Sub", "ClientID", clID, "Topic", topic)

				startTime := time.Now()
				published, received, err := pkgmqtt.RunPubSub(clientCfg, pubSubCfg, logger)
				duration := time.Since(startTime)

				result := TestResult{
					Protocol:          "MQTT",
					ClientID:          clID,
					ChannelID:         chID,
					Topic:             topic,
					Success:           err == nil,
					Error:             err,
					Duration:          duration,
					Messages:          published + received,
					PublishedMessages: published,
					ReceivedMessages:  received,
				}
				aggregator.AddResult(result)

				if err != nil {
					logger.Error("MQTT Pub/Sub failed", "ClientID", clID, "Topic", topic, "error", err)
				} else {
					logger.Debug("MQTT Pub/Sub completed successfully", "ClientID", clID, "Topic", topic, "duration", duration)
				}
			}(channelID, clientID)
		}
	}

	wg.Wait()
}

func testCOAPPubSub(cfg Config, clientIDs, channelIDs []string, clientSecretMap map[string]string, aggregator *ResultsAggregator, logger *slog.Logger, tlsConfig *tls.Config) {
	pubSubCfg := pkgcoap.PubSubConfig{
		MessageCount: cfg.COAPMessageCount,
		Delay:        cfg.COAPDelay,
		Timeout:      cfg.COAPTimeout,
		MessageSize:  cfg.COAPMessageSize,
		TLSConfig:    tlsConfig,
	}

	var wg sync.WaitGroup

	for _, channelID := range channelIDs {
		for _, clientID := range clientIDs {
			wg.Add(1)
			go func(chID, clID string) {
				defer wg.Done()

				topic := constructTopic(cfg.DomainID, chID, clID, "coap")
				clientCfg := pkgcoap.ClientConfig{
					Host: cfg.COAPHost,
					Port: cfg.COAPPort,
					Path: topic,
					Opts: []message.Option{{Value: fmt.Appendf(nil, "auth=%s", clientSecretMap[clientID]), ID: message.URIQuery}},
				}

				logger.Debug("Starting COAP Pub/Sub", "ClientID", clID, "Topic", topic)

				startTime := time.Now()
				published, received, err := pkgcoap.RunPubSub(clientCfg, pubSubCfg, *logger)
				duration := time.Since(startTime)

				result := TestResult{
					Protocol:          "CoAP",
					ClientID:          clID,
					ChannelID:         chID,
					Topic:             topic,
					Success:           err == nil,
					Error:             err,
					Duration:          duration,
					Messages:          published + received,
					PublishedMessages: published,
					ReceivedMessages:  received,
				}
				aggregator.AddResult(result)

				if err != nil {
					logger.Error("COAP Pub/Sub failed", "ClientID", clID, "Topic", topic, "error", err)
				} else {
					logger.Debug("COAP Pub/Sub completed successfully", "ClientID", clID, "Topic", topic, "duration", duration)
				}
			}(channelID, clientID)
		}
	}

	wg.Wait()
}

func testHTTPPublish(cfg Config, clientIDs, channelIDs []string, clientSecretMap map[string]string, aggregator *ResultsAggregator, logger *slog.Logger, tlsConfig *tls.Config) {
	pubCfg := pkghttp.PublishConfig{
		MessageCount: cfg.HTTPMessageCount,
		Delay:        cfg.HTTPDelay,
		Timeout:      cfg.HTTPTimeout,
		Method:       "POST",
		MessageSize:  cfg.HTTPMessageSize,
		TLSConfig:    tlsConfig,
	}

	var wg sync.WaitGroup

	for _, channelID := range channelIDs {
		for _, clientID := range clientIDs {
			wg.Add(1)
			go func(chID, clID string) {
				defer wg.Done()

				endpoint := constructTopic(cfg.DomainID, chID, clID, "http")
				clientCfg := pkghttp.ClientConfig{
					BaseURL:  cfg.HTTPBaseUrl,
					Endpoint: endpoint,
					Headers: map[string]string{
						"Authorization": fmt.Sprintf("Client %s", clientSecretMap[clID]),
						"Content-Type":  "application/senml+json",
					},
				}

				logger.Debug("Starting HTTP Publish", "ClientID", clID, "Endpoint", endpoint)

				startTime := time.Now()
				published, _, _, err := pkghttp.RunPublish(clientCfg, pubCfg, logger)
				duration := time.Since(startTime)

				result := TestResult{
					Protocol:          "HTTP",
					ClientID:          clID,
					ChannelID:         chID,
					Topic:             endpoint,
					Success:           err == nil,
					Error:             err,
					Duration:          duration,
					Messages:          published,
					PublishedMessages: published,
					ReceivedMessages:  0, // HTTP is publish-only
				}
				aggregator.AddResult(result)

				if err != nil {
					logger.Error("HTTP Publish failed", "ClientID", clID, "Endpoint", endpoint, "error", err)
				} else {
					logger.Debug("HTTP Publish completed successfully", "ClientID", clID, "Endpoint", endpoint, "duration", duration)
				}
			}(channelID, clientID)
		}
	}

	wg.Wait()
}

func testWSPubSub(cfg Config, clientIDs, channelIDs []string, clientSecretMap map[string]string, aggregator *ResultsAggregator, logger *slog.Logger, tlsConfig *tls.Config) {
	pubCfg := pkgws.PubSubConfig{
		MessageCount: cfg.WSMessageCount,
		Delay:        cfg.WSDelay,
		Timeout:      cfg.WSTimeout,
		MessageSize:  cfg.WSMessageSize,
		TLSConfig:    tlsConfig,
	}

	var wg sync.WaitGroup
	for _, channelID := range channelIDs {
		for _, clientID := range clientIDs {
			wg.Add(1)
			go func(chID, clID string) {
				defer wg.Done()

				topic := constructTopic(cfg.DomainID, chID, clID, "ws")
				clientCfg := pkgws.ClientConfig{
					ServerURL: cfg.WSBaseUrl,
					Path:      topic,
					Headers: map[string]string{
						"Authorization": fmt.Sprintf("%s", clientSecretMap[clID]),
						"Content-Type":  "application/senml+json",
					},
				}

				logger.Debug("Starting WebSocket Pub/Sub", "ClientID", clID, "Topic", topic)

				startTime := time.Now()
				published, received, err := pkgws.RunPubSub(clientCfg, pubCfg, logger)
				duration := time.Since(startTime)

				result := TestResult{
					Protocol:          "WebSocket",
					ClientID:          clID,
					ChannelID:         chID,
					Topic:             topic,
					Success:           err == nil,
					Error:             err,
					Duration:          duration,
					Messages:          published + received,
					PublishedMessages: published,
					ReceivedMessages:  received,
				}
				aggregator.AddResult(result)

				if err != nil {
					logger.Error("WebSocket Pub/Sub failed", "ClientID", clID, "Topic", topic, "error", err)
				} else {
					logger.Debug("WebSocket Pub/Sub completed successfully", "ClientID", clID, "Topic", topic, "duration", duration)
				}
			}(channelID, clientID)
		}
	}

	wg.Wait()
}

func constructTopic(domainID, channelID, clientID, protocol string) string {
	return fmt.Sprintf("m/%s/c/%s/%s-%s", domainID, channelID, clientID, protocol)
}

func buildClientSecretMap(clients []sdk.Client) map[string]string {
	secretMap := make(map[string]string)
	for _, client := range clients {
		secretMap[client.ID] = client.Credentials.Secret
	}
	return secretMap
}

func cleanUpProvision(ctx context.Context, channelIDs, clientIDs []string, domainID, token string, channelsSDK *pkg.Channels, clientsSDK *pkg.Clients, logger *slog.Logger) error {
	var hasErrors bool
	for _, chID := range channelIDs {
		if err := channelsSDK.DeleteChannel(ctx, chID, domainID, token); err != nil {
			logger.Warn("Failed to delete channel", "channel_id", chID, "error", err)
			hasErrors = true
		}
	}
	for _, clID := range clientIDs {
		if err := clientsSDK.DeleteClient(ctx, clID, domainID, token); err != nil {
			logger.Warn("Failed to delete client", "client_id", clID, "error", err)
			hasErrors = true
		}
	}
	if hasErrors {
		return fmt.Errorf("some resources failed to delete")
	}

	return nil
}
