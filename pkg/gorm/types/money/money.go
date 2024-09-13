package money

import (
	"context"
	"fmt"
	"math"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const precision = 2

var factor = math.Pow10(precision)

type ErrUnsupportedDBValue struct {
	Value any
}

func (err ErrUnsupportedDBValue) Error() string {
	return fmt.Sprintf("unsupported db value: %v", err.Value)
}

type Amount uint64

func New(value float64) Amount {
	return Amount(math.Round(value * factor))
}

func (Amount) GormDataType() string {
	return "uint"
}

func (amount *Amount) Scan(value any) error {
	count, ok := value.(int64) // Go sql doesnt support uint
	if !ok {
		return ErrUnsupportedDBValue{Value: value}
	}

	*amount = Amount(count)

	return nil
}

func (amount Amount) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	return clause.Expr{
		SQL: "?",
		Vars: []any{
			uint64(amount),
		},
	}
}

func (amount *Amount) AsFloat() float64 {
	return float64(*amount) / factor
}
