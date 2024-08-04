package validator

import (
	"github.com/asaskevich/govalidator"
	"strconv"
)

func init() {
	govalidator.CustomTypeTagMap.Set("luhn", func(value any, context any) bool {
		return IsLuhn(value)
	})
	govalidator.CustomTypeTagMap.Set("positive", func(value any, context any) bool {
		return IsPositive(value)
	})
}

func IsLuhn(value any) bool {
	converted, ok := toUint64(value)
	if !ok {
		return false
	}

	number, controlDigit := converted/10, converted%10

	return (controlDigit+luhnChecksum(number))%10 == 0
}

func IsPositive(value any) bool {
	switch typed := value.(type) {
	case uint:
		return typed > 0
	case int:
		return typed > 0
	case uint8:
		return typed > 0
	case uint16:
		return typed > 0
	case uint32:
		return typed > 0
	case uint64:
		return typed > 0
	case int8:
		return typed > 0
	case int16:
		return typed > 0
	case int32:
		return typed > 0
	case int64:
		return typed > 0
	case string:
		float, err := strconv.ParseFloat(typed, 64)
		if err != nil {
			return false
		}

		return float > 0
	default:
		return false
	}
}

func toUint64(value any) (uint64, bool) {
	switch typed := value.(type) {
	case uint:
		return uint64(typed), true
	case uint8:
		return uint64(typed), true
	case uint16:
		return uint64(typed), true
	case uint32:
		return uint64(typed), true
	case uint64:
		return typed, true
	case int:
		if typed < 0 {
			return 0, false
		}
		return uint64(typed), true
	case int8:
		if typed < 0 {
			return 0, false
		}
		return uint64(typed), true
	case int16:
		if typed < 0 {
			return 0, false
		}
		return uint64(typed), true
	case int32:
		if typed < 0 {
			return 0, false
		}
		return uint64(typed), true
	case int64:
		if typed < 0 {
			return 0, false
		}
		return uint64(typed), true
	case string:
		converted, err := strconv.ParseUint(typed, 10, 64)
		if err != nil {
			return 0, false
		}

		return converted, true
	case []byte:
		converted, err := strconv.ParseUint(string(typed), 10, 64)
		if err != nil {
			return 0, false
		}

		return converted, true
	default:
		return 0, false
	}
}

func luhnChecksum(number uint64) uint64 {
	var luhn uint64

	for i := 0; number > 0; i++ {
		current := number % 10

		if i%2 == 0 {
			current = current * 2
			if current > 9 {
				current = current%10 + current/10
			}
		}

		luhn += current
		number = number / 10
	}
	return luhn % 10
}
