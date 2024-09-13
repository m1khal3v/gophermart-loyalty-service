package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const ttl = time.Hour * 24

var ErrInvalidClaims = errors.New("invalid claims")

type Claims struct {
	jwt.RegisteredClaims

	SubjectID uint32 `json:"sub_id"`
}

func (claims Claims) GetSubjectID() uint32 {
	return claims.SubjectID
}

type Container struct {
	secret []byte
}

func New(secret string) *Container {
	return &Container{
		secret: []byte(secret),
	}
}

func (container *Container) Encode(subjectID uint32, subject string) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, Claims{
		SubjectID: subjectID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	})

	return token.SignedString(container.secret)
}

func (container *Container) Decode(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return container.secret, nil
	}, jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithValidMethods([]string{"HS512"}))

	if err != nil {
		return nil, err
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}
