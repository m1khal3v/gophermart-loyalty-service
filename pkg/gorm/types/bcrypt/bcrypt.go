package bcrypt

import (
	"database/sql/driver"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// RecommendedCost must be >= 12
// See IETF article https://www.ietf.org/archive/id/draft-ietf-kitten-password-storage-07.html#name-bcrypt
const RecommendedCost = 12
const MinCost = bcrypt.MinCost
const MaxCost = bcrypt.MaxCost

type ErrUnsupportedDBValue struct {
	Value any
}

func (err ErrUnsupportedDBValue) Error() string {
	return fmt.Sprintf("unsupported db value: %v", err.Value)
}

type Hash []byte

func NewHash(password string, cost int) (Hash, error) {
	return bcrypt.GenerateFromPassword([]byte(password), cost)
}

func (Hash) GormDataType() string {
	return "bytes"
}

func (hash *Hash) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return ErrUnsupportedDBValue{Value: value}
	}

	*hash = bytes

	return nil
}

func (hash Hash) Value() (driver.Value, error) {
	return []byte(hash), nil
}

func (hash Hash) CompareWithPassword(password string) error {
	return bcrypt.CompareHashAndPassword(hash, []byte(password))
}

func (hash Hash) Cost() (int, error) {
	return bcrypt.Cost(hash)
}
