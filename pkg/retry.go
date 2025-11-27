package pkg

import (
	"fmt"
	"log/slog"
	"time"
)

const (
	MaxRetries        = 3
	InitialBackoff    = 1 * time.Second
	BackoffMultiplier = 2
)

type RetryConfig struct {
	MaxRetries int
	Timeout    time.Duration
	Logger     slog.Logger
}

func RetryWithExponentialBackoff(config RetryConfig, operation func() error, operationName string) error {
	var lastErr error
	backoff := InitialBackoff

	for attempt := 1; attempt <= config.MaxRetries; attempt++ {
		config.Logger.Debug("Attempting operation", "operation", operationName, "attempt", attempt, "max_retries", config.MaxRetries)

		err := operation()
		if err == nil {
			if attempt > 1 {
				config.Logger.Info("Operation succeeded after retry", "operation", operationName, "attempt", attempt)
			}
			return nil
		}

		lastErr = err
		config.Logger.Warn("Operation failed", "operation", operationName, "attempt", attempt, "error", err)

		if attempt < config.MaxRetries {
			if config.Timeout > 0 && backoff > config.Timeout {
				backoff = config.Timeout
			}

			config.Logger.Debug("Retrying after backoff", "operation", operationName, "backoff", backoff)
			time.Sleep(backoff)

			backoff *= BackoffMultiplier
		}
	}

	return fmt.Errorf("%s failed after %d attempts: %w", operationName, config.MaxRetries, lastErr)
}
