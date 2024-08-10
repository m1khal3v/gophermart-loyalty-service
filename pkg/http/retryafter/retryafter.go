package retryafter

import (
	"strconv"
	"time"
)

func Parse(retryAfter string, defaultValue time.Duration) time.Duration {
	if duration, err := parseSeconds(retryAfter); err == nil {
		return duration
	}

	if dateTime, err := parseHTTPDate(retryAfter); err == nil {
		duration := time.Since(dateTime)

		if duration < 0 {
			return 0
		}

		return duration
	}

	return defaultValue
}

func parseSeconds(retryAfter string) (time.Duration, error) {
	seconds, err := strconv.ParseUint(retryAfter, 10, 64)
	if err != nil {
		return time.Duration(0), err
	}

	return time.Second * time.Duration(seconds), nil
}

func parseHTTPDate(retryAfter string) (time.Time, error) {
	dateTime, err := time.Parse(time.RFC1123, retryAfter)
	if err != nil {
		return time.Time{}, err
	}

	return dateTime, nil
}
