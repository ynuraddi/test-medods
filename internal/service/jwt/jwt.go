package jwt

import (
	"fmt"
	"medods/internal/model"
	"medods/pkg/logger"

	"github.com/golang-jwt/jwt/v5"
)

type Interface interface {
	CreateToken(payload model.Payload) (string, error)
	VerifyToken(tokenString string) (token *jwt.Token, payload *model.Payload, err error)
}

type Maker struct {
	sercretKey    []byte
	signingMethod jwt.SigningMethod
}

func New(secretKey []byte, logger logger.Interface) *Maker {
	return &Maker{
		signingMethod: jwt.SigningMethodHS512,
		sercretKey:    secretKey,
	}
}

func (s Maker) CreateToken(payload model.Payload) (string, error) {
	token := jwt.NewWithClaims(s.signingMethod, payload)
	return token.SignedString(s.sercretKey)
}

func (s Maker) VerifyToken(tokenString string) (token *jwt.Token, payload *model.Payload, err error) {
	payload = &model.Payload{}
	token, err = jwt.ParseWithClaims(tokenString, payload, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return s.sercretKey, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS512.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
	// do not check error, should return err for outside logic

	if payload.UserID <= 0 {
		return nil, nil,
			fmt.Errorf("%w: user_id is requied", jwt.ErrTokenInvalidClaims)
	}
	if len(payload.IP) == 0 {
		return nil, nil,
			fmt.Errorf("%w: ip is required", jwt.ErrTokenInvalidClaims)
	}

	return token, payload, err
}
