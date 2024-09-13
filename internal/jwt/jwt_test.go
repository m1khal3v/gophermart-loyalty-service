package jwt

import (
	crypto "crypto/rand"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecode(t *testing.T) {
	secret := make([]byte, 0, 32)
	_, err := crypto.Read(secret)
	require.NoError(t, err)

	jwt := New(fmt.Sprintf("%x", secret))

	id := rand.Uint32N(1000000) + 100
	subject := make([]byte, 0, 32)
	_, err = crypto.Read(subject)
	require.NoError(t, err)
	subjectString := fmt.Sprintf("%x", subject)

	token, err := jwt.Encode(id, subjectString)
	require.NoError(t, err)

	claims, err := jwt.Decode(token)
	require.NoError(t, err)

	assert.Equal(t, id, claims.SubjectID)
	assert.Equal(t, subjectString, claims.Subject)
}
