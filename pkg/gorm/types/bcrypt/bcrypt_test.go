package bcrypt

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand/v2"
	"testing"
)

func TestOK(t *testing.T) {
	// cost := rand.IntN(MaxCost-MinCost) + MinCost - values near max cost is very slow
	cost := rand.IntN(MaxCost-MinCost-20) + MinCost
	hash, err := NewHash("test_password", cost)
	require.NoError(t, err)
	hashCost, err := hash.Cost()
	require.NoError(t, err)
	assert.Equal(t, cost, hashCost)
	assert.NoError(t, hash.CompareWithPassword("test_password"))
}

func TestInvalidCost(t *testing.T) {
	_, err := NewHash("test_password", MaxCost+1)
	require.Error(t, err)
}
