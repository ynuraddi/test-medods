package jwt

import (
	"medods/internal/model"
	mock_logger "medods/pkg/logger/mock"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var (
	secretKey = []byte("secret")
)

func TestCreateToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	l := mock_logger.NewMockInterface(ctrl)

	now := time.Now()

	jwtMaker := New(secretKey, l)
	defaultPayload := model.Payload{
		UserID: 1,
		IP:     "2",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)),
		},
	}

	token, err := jwtMaker.CreateToken(defaultPayload)
	assert.NoError(t, err)

	jwtToken, payload, err := jwtMaker.VerifyToken(token)
	assert.NoError(t, err)
	assert.Equal(t, defaultPayload, *payload)
	assert.True(t, jwtToken.Valid)
}

func TestVerifyToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	l := mock_logger.NewMockInterface(ctrl)

	iat := time.Now().Add(-1 * time.Minute)
	exp := iat.Add(30 * time.Minute)

	jwtMaker := New(secretKey, l)
	defaultPayload := model.Payload{
		UserID: 1,
		IP:     "2",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(iat),
			ID:        "test",
		},
	}

	tc := []struct {
		name        string
		input       func(t *testing.T) (token string)
		checkResult func(t *testing.T, token *jwt.Token, payload *model.Payload, err error)
	}{
		{
			name: "OK",
			input: func(t *testing.T) (token string) {
				token, err := jwtMaker.CreateToken(defaultPayload)
				assert.NoError(t, err)
				return token
			},
			checkResult: func(t *testing.T, token *jwt.Token, payload *model.Payload, err error) {
				assert.NoError(t, err)
				assert.True(t, token.Valid)
				assert.Equal(t, defaultPayload, *payload)
			},
		},
		{
			name: "wrong signing method",
			input: func(t *testing.T) (token string) {
				token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, defaultPayload).
					SignedString(secretKey)
				assert.NoError(t, err)
				return token
			},
			checkResult: func(t *testing.T, token *jwt.Token, payload *model.Payload, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, jwt.ErrTokenSignatureInvalid)
			},
		},
		{
			name: "token is expired",
			input: func(t *testing.T) (token string) {
				df := defaultPayload

				// expired 30 min ago
				iat := time.Now().Add(-1 * time.Hour)
				exp := iat.Add(30 * time.Minute)

				df.IssuedAt = jwt.NewNumericDate(iat)
				df.ExpiresAt = jwt.NewNumericDate(exp)

				token, err := jwtMaker.CreateToken(df)
				assert.NoError(t, err)
				return token
			},
			checkResult: func(t *testing.T, token *jwt.Token, payload *model.Payload, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, jwt.ErrTokenExpired)
			},
		},
		{
			name: "field exp is required",
			input: func(t *testing.T) (token string) {
				df := defaultPayload

				df.ExpiresAt = nil

				token, err := jwtMaker.CreateToken(df)
				assert.NoError(t, err)
				return token
			},
			checkResult: func(t *testing.T, token *jwt.Token, payload *model.Payload, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, jwt.ErrTokenInvalidClaims)
			},
		},
		{
			name: "field user_id is required",
			input: func(t *testing.T) (token string) {
				df := defaultPayload

				df.UserID = 0

				token, err := jwtMaker.CreateToken(df)
				assert.NoError(t, err)
				return token
			},
			checkResult: func(t *testing.T, token *jwt.Token, payload *model.Payload, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, jwt.ErrTokenInvalidClaims)
				assert.Empty(t, token)
				assert.Empty(t, payload)
			},
		},
		{
			name: "field ip is required",
			input: func(t *testing.T) (token string) {
				df := defaultPayload

				df.IP = ""

				token, err := jwtMaker.CreateToken(df)
				assert.NoError(t, err)
				return token
			},
			checkResult: func(t *testing.T, token *jwt.Token, payload *model.Payload, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, jwt.ErrTokenInvalidClaims)
				assert.Empty(t, token)
				assert.Empty(t, payload)
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			token, payload, err := jwtMaker.VerifyToken(test.input(t))
			test.checkResult(t, token, payload, err)
		})
	}
}
