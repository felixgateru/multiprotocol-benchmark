package pkg

import (
	"fmt"
	"log/slog"
	"time"
)

const (
	maxRetries        = 3
	initialBackoff    = 1 * time.Second
	backoffMultiplier = 2
)

type RetryConfig struct {
	Timeout time.Duration
	Logger  *slog.Logger
}

func RetryWithExponentialBackoff(config RetryConfig, operation func() error, operationName string) error {
	var lastErr error
	backoff := initialBackoff

	for attempt := 1; attempt <= maxRetries; attempt++ {
		config.Logger.Debug("Attempting operation", "operation", operationName, "attempt", attempt, "max_retries", maxRetries)

		err := operation()
		if err == nil {
			if attempt > 1 {
				config.Logger.Info("Operation succeeded after retry", "operation", operationName, "attempt", attempt)
			}
			return nil
		}

		lastErr = err
		config.Logger.Warn("Operation failed", "operation", operationName, "attempt", attempt, "error", err)

		if attempt < maxRetries {
			if config.Timeout > 0 && backoff > config.Timeout {
				backoff = config.Timeout
			}

			config.Logger.Debug("Retrying after backoff", "operation", operationName, "backoff", backoff)
			time.Sleep(backoff)

			backoff *= backoffMultiplier
		}
	}

	return fmt.Errorf("%s failed after %d attempts: %w", operationName, maxRetries, lastErr)
}
