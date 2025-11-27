package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// ProtocolResult holds the aggregated results for a single protocol test
type ProtocolResult struct {
	Protocol           string        `json:"protocol"`
	TotalClients       int           `json:"total_clients"`
	TotalChannels      int           `json:"total_channels"`
	TotalTests         int           `json:"total_tests"`
	SuccessfulTests    int           `json:"successful_tests"`
	FailedTests        int           `json:"failed_tests"`
	SuccessRate        float64       `json:"success_rate_percent"`
	TotalMessages      int           `json:"total_messages"`
	PublishedMessages  int           `json:"published_messages"`
	ReceivedMessages   int           `json:"received_messages"`
	PublishSuccessRate float64       `json:"publish_success_rate_percent"`
	ReceiveSuccessRate float64       `json:"receive_success_rate_percent"`
	MessageSize        int           `json:"message_size_bytes"`
	TotalDataSent      int64         `json:"total_data_sent_bytes"`
	TotalDataSentMB    float64       `json:"total_data_sent_mb"`
	TotalDuration      time.Duration `json:"total_duration_ns"`
	AvgDuration        time.Duration `json:"avg_duration_per_test_ns"`
	TotalDurationSec   float64       `json:"total_duration_seconds"`
	Throughput         float64       `json:"throughput_msgs_per_sec"`
	Bandwidth          float64       `json:"bandwidth_mbps"`
	Errors             []string      `json:"errors,omitempty"`
}

// TestResult represents a single test execution
type TestResult struct {
	Protocol          string
	ClientID          string
	ChannelID         string
	Topic             string
	Success           bool
	Error             error
	Duration          time.Duration
	Messages          int // Total messages (published + received)
	PublishedMessages int // Messages published
	ReceivedMessages  int // Messages received/subscribed
}

// ResultsAggregator collects and aggregates test results
type ResultsAggregator struct {
	mu      sync.Mutex
	results map[string][]TestResult // protocol -> []TestResult
}

// NewResultsAggregator creates a new results aggregator
func NewResultsAggregator() *ResultsAggregator {
	return &ResultsAggregator{
		results: make(map[string][]TestResult),
	}
}

// AddResult adds a test result to the aggregator
func (ra *ResultsAggregator) AddResult(result TestResult) {
	ra.mu.Lock()
	defer ra.mu.Unlock()
	ra.results[result.Protocol] = append(ra.results[result.Protocol], result)
}

// GenerateReport generates an aggregated report for all protocols
func (ra *ResultsAggregator) GenerateReport(messageSize int, delay time.Duration) map[string]ProtocolResult {
	ra.mu.Lock()
	defer ra.mu.Unlock()

	report := make(map[string]ProtocolResult)

	for protocol, results := range ra.results {
		if len(results) == 0 {
			continue
		}

		pr := ProtocolResult{
			Protocol:    protocol,
			TotalTests:  len(results),
			MessageSize: messageSize,
			Errors:      []string{},
		}

		var totalDuration time.Duration
		clientMap := make(map[string]bool)
		channelMap := make(map[string]bool)
		totalMessages := 0
		totalPublished := 0
		totalReceived := 0

		for _, result := range results {
			clientMap[result.ClientID] = true
			channelMap[result.ChannelID] = true
			totalDuration += result.Duration
			totalMessages += result.Messages
			totalPublished += result.PublishedMessages
			totalReceived += result.ReceivedMessages

			if result.Success {
				pr.SuccessfulTests++
			} else {
				pr.FailedTests++
				if result.Error != nil {
					pr.Errors = append(pr.Errors, fmt.Sprintf("%s/%s: %v", result.ClientID, result.ChannelID, result.Error))
				}
			}
		}

		pr.TotalClients = len(clientMap)
		pr.TotalChannels = len(channelMap)
		pr.TotalMessages = totalMessages
		pr.PublishedMessages = totalPublished
		pr.ReceivedMessages = totalReceived

		// Calculate success rates for publish and receive based on actual counts
		if totalPublished > 0 {
			pr.PublishSuccessRate = 100.0 // If we got published count, it succeeded
		}
		if totalReceived > 0 && totalPublished > 0 {
			pr.ReceiveSuccessRate = float64(totalReceived) / float64(totalPublished) * 100
		}

		pr.TotalDataSent = int64(totalMessages * messageSize)
		pr.TotalDataSentMB = float64(pr.TotalDataSent) / (1024 * 1024)
		pr.TotalDuration = totalDuration
		pr.TotalDurationSec = totalDuration.Seconds()
		pr.AvgDuration = totalDuration / time.Duration(len(results))
		pr.SuccessRate = float64(pr.SuccessfulTests) / float64(pr.TotalTests) * 100

		// Calculate throughput (messages per second)
		if pr.TotalDurationSec > 0 {
			pr.Throughput = float64(pr.TotalMessages) / pr.TotalDurationSec
		}

		// Calculate bandwidth (Mbps) only if delay is 0
		if delay == 0 && pr.TotalDurationSec > 0 {
			// Bandwidth in bits per second
			bitsPerSecond := float64(pr.TotalDataSent*8) / pr.TotalDurationSec
			// Convert to Mbps
			pr.Bandwidth = bitsPerSecond / (1024 * 1024)
		}

		report[protocol] = pr
	}

	return report
}

// SaveToFile saves the aggregated results to a JSON file
func (ra *ResultsAggregator) SaveToFile(filename string, messageSize int, delay time.Duration) error {
	report := ra.GenerateReport(messageSize, delay)

	// Create a summary structure
	summary := struct {
		Timestamp      string                    `json:"timestamp"`
		TotalProtocols int                       `json:"total_protocols"`
		Results        map[string]ProtocolResult `json:"results"`
	}{
		Timestamp:      time.Now().Format(time.RFC3339),
		TotalProtocols: len(report),
		Results:        report,
	}

	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write results to file: %w", err)
	}

	return nil
}

// PrintSummary prints a formatted summary of all protocol results
func (ra *ResultsAggregator) PrintSummary(messageSize int, delay time.Duration) {
	report := ra.GenerateReport(messageSize, delay)

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                        PROTOCOL TEST RESULTS SUMMARY")
	fmt.Println(strings.Repeat("=", 80))

	if len(report) == 0 {
		fmt.Println("No test results available")
		return
	}

	for protocol, pr := range report {
		fmt.Printf("\n%s PROTOCOL\n", strings.ToUpper(protocol))
		fmt.Println(strings.Repeat("-", 80))
		fmt.Printf("  Clients:              %d\n", pr.TotalClients)
		fmt.Printf("  Channels:             %d\n", pr.TotalChannels)
		fmt.Printf("  Total Tests:          %d\n", pr.TotalTests)
		fmt.Printf("  Successful:           %d\n", pr.SuccessfulTests)
		fmt.Printf("  Failed:               %d\n", pr.FailedTests)
		fmt.Printf("  Success Rate:         %.2f%%\n", pr.SuccessRate)
		fmt.Println()
		fmt.Printf("  Publishing Metrics:\n")
		fmt.Printf("    Published Messages: %d\n", pr.PublishedMessages)
		if pr.PublishSuccessRate > 0 {
			fmt.Printf("    Publish Success:    %.2f%%\n", pr.PublishSuccessRate)
		}
		fmt.Println()
		fmt.Printf("  Subscription Metrics:\n")
		fmt.Printf("    Received Messages:  %d\n", pr.ReceivedMessages)
		if pr.ReceiveSuccessRate > 0 {
			fmt.Printf("    Receive Success:    %.2f%%\n", pr.ReceiveSuccessRate)
		}
		fmt.Println()
		fmt.Printf("  Overall Metrics:\n")
		fmt.Printf("    Total Messages:     %d\n", pr.TotalMessages)
		fmt.Printf("    Message Size:       %d bytes\n", pr.MessageSize)
		fmt.Printf("    Total Data Sent:    %.2f MB (%d bytes)\n", pr.TotalDataSentMB, pr.TotalDataSent)
		fmt.Println()
		fmt.Printf("  Performance:\n")
		fmt.Printf("    Total Duration:     %.2f seconds\n", pr.TotalDurationSec)
		fmt.Printf("    Avg Duration/Test:  %v\n", pr.AvgDuration)
		fmt.Printf("    Throughput:         %.2f msgs/sec\n", pr.Throughput)

		if delay == 0 && pr.Bandwidth > 0 {
			fmt.Printf("    Bandwidth:          %.2f Mbps\n", pr.Bandwidth)
		}

		if len(pr.Errors) > 0 {
			fmt.Println()
			fmt.Printf("  Errors (%d):\n", len(pr.Errors))
			for i, err := range pr.Errors {
				if i < 5 { // Show first 5 errors
					fmt.Printf("    - %s\n", err)
				}
			}
			if len(pr.Errors) > 5 {
				fmt.Printf("    ... and %d more errors\n", len(pr.Errors)-5)
			}
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println()
}
