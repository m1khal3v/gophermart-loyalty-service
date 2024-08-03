package validator

import (
	"github.com/asaskevich/govalidator"
	"strconv"
)

func init() {
	govalidator.CustomTypeTagMap.Set("luhn", func(value any, context any) bool {
		return IsLuhn(value)
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

func toUint64(value any) (uint64, bool) {
	switch typed := value.(type) {
	case uint64:
		return typed, true
	case uint8, uint16, uint32, uint, int8, int16, int32, int64, int:
		converted, ok := typed.(uint64)
		if !ok {
			return 0, false
		}

		return converted, true
	case string:
		converted, err := strconv.ParseUint(typed, 10, 64)
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
