package retry

import (
	"time"
)

func Retry(
	baseDelay,
	maxDelay time.Duration,
	retries,
	multiplier uint64,
	function func() error,
	filter func(err error) bool,
) error {
	var err error
	retry := uint64(0)

	for {
		if retry > retries {
			return err
		}

		if err = function(); err == nil {
			return nil
		}

		if filter != nil && !filter(err) {
			return err
		}

		time.Sleep(calculateDelay(baseDelay, maxDelay, retry, multiplier))
		retry++
	}
}

func pow(x, y uint64) uint64 {
	if y == 0 {
		return 1
	}

	if y == 1 {
		return x
	}

	result := x
	for i := uint64(2); i <= y; i++ {
		result *= x
	}
	return result
}

func calculateDelay(baseDelay time.Duration, maxDelay time.Duration, attempt uint64, multiplier uint64) time.Duration {
	if attempt == 0 {
		return min(baseDelay, maxDelay)
	} else {
		return min(baseDelay*time.Duration(pow(multiplier, attempt)), maxDelay)
	}
}
