package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"medods/internal/model"
	"medods/internal/service/jwt"
	"medods/internal/service/session"

	"medods/pkg/logger"
	"time"

	gjwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	_accessTokenLifeTime  = 30 * time.Minute
	_refreshTokenLifeTime = 30 * 24 * time.Hour
)

type Interface interface {
	CreateSession(ctx context.Context, uid int, ip string) (aToken string, rToken string, err error)
	RefreshSession(ctx context.Context, aT string, rT string) (aToken string, rToken string, err error)
}

type auth struct {
	session session.Interface
	jwt     jwt.Interface

	// TODO: add notification service

	logger   logger.Interface
	testMode bool
}

func New(
	sessionService session.Interface,
	jwtMaker jwt.Interface,
	logger logger.Interface,
	testMode bool,
) *auth {
	return &auth{
		session: sessionService,
		jwt:     jwtMaker,

		logger: logger,

		testMode: testMode,
	}
}

// Notify: do not check user with uid exists or not, pls use only correct input
func (s auth) CreateSession(ctx context.Context, uid int, ip string) (aToken, rToken string, err error) {
	iat := time.Now()
	jti := s.generateUUID()

	aToken, rToken, err = s.createTokens(uid, ip, iat, jti)
	if err != nil {
		return "", "", err
	}
	s.logger.Debug("tokens created")

	rTokenHash, err := s.hashString(rToken)
	if err != nil {
		s.logger.Error(err)
		return "", "", err
	}
	s.logger.Debug("refresh token hashed")

	session, err := s.session.GetByUserID(ctx, uid)
	if errors.Is(err, sql.ErrNoRows) { // create session if not exists
		if err := s.session.Create(ctx, model.Session{
			UserID:     uid,
			ATokenID:   jti,
			RTokenHash: rTokenHash,
			IP:         ip,
			CreatedAt:  iat.Unix(),
		}); err != nil {
			s.logger.Error("failed to create session: %s", err.Error())
			return "", "", err
		}
		s.logger.Debug("session success created")
		return aToken, rToken, nil
	} else if err == nil { // update session if exists

		// можно добавить проверку на логирование с нового ip, сравнив текущий ip с тем что в базе

		if _, err := s.session.Update(ctx, model.Session{
			ID:         session.ID,
			UserID:     uid,
			ATokenID:   jti,
			RTokenHash: rTokenHash,
			IP:         ip,
			CreatedAt:  iat.Unix(),
			Version:    session.Version,
		}); err != nil {
			s.logger.Error("failed to update session: %s", err.Error())
			return "", "", err
		}
		s.logger.Debug("session success created")
		return aToken, rToken, nil
	}

	return "", "", err
}

func (s auth) RefreshSession(ctx context.Context, aT, rT string) (aToken, rToken string, err error) {
	_, payload, err := s.jwt.VerifyToken(aT)
	if err != nil && !errors.Is(err, gjwt.ErrTokenExpired) {
		err := fmt.Errorf("failed to verify access token: %w", err)
		s.logger.Error(err)
		return "", "", err
	}
	s.logger.Debug("success verified token")

	db, err := s.session.GetByUserID(ctx, payload.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		err = fmt.Errorf("session not exists: %w", err)
		s.logger.Error(err)
		return "", "", err
	} else if err != nil {
		err = fmt.Errorf("failed to get user: %w", err)
		s.logger.Error(err)
		return "", "", err
	}
	s.logger.Debug("success got user")

	now := time.Now()
	iat := payload.IssuedAt.Time
	rExp := iat.Add(_refreshTokenLifeTime)
	iatU := iat.Unix()
	if iatU != db.CreatedAt {
		err := fmt.Errorf("failed to validate refresh token: invalid iat")
		s.logger.Error(err)
		return "", "", err
	} else if rExp.Before(now) {
		err := fmt.Errorf("failed to validate refresh token: %w", gjwt.ErrTokenExpired)
		s.logger.Error(err)
		return "", "", err
	}

	if !s.compareHash(db.RTokenHash, rT) {
		err := fmt.Errorf("failed to validate refresh token: invalid token")
		s.logger.Error(err)
		return "", "", err
	} else if payload.ID != db.ATokenID {
		err := fmt.Errorf("failed to validate access token: invalid jti")
		s.logger.Error(err)
		return "", "", err
	}

	if db.IP != payload.IP {
		// TODO: notificate user
		// fmt.Println("IP was hacked*. Home Address is Russia, Moscow, ")
		s.logger.Warn("login from new IP addess")
	}

	return s.CreateSession(ctx, payload.UserID, payload.IP)
}

func (s auth) createTokens(uid int, ip string, iat time.Time, jti string) (aToken, rToken string, err error) {
	aToken, err = s.jwt.CreateToken(model.Payload{
		UserID: uid,
		IP:     ip,

		RegisteredClaims: gjwt.RegisteredClaims{
			ID:        jti,
			IssuedAt:  gjwt.NewNumericDate(iat),
			ExpiresAt: gjwt.NewNumericDate(iat.Add(_accessTokenLifeTime)),
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to create access token: %w", err)
	}

	rToken, err = s.randString()
	if err != nil {
		return "", "", fmt.Errorf("failed to create refresh token: %w", err)
	}

	return aToken, rToken, nil
}

func (s auth) generateUUID() string {
	if s.testMode {
		return "uuid_string"
	}

	return uuid.NewString()
}

func (s auth) randString() (string, error) {
	if s.testMode {
		return "rand_string", nil
	}

	str := make([]byte, 52)
	if _, err := rand.Read(str); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(str), nil
}

func (s auth) hashString(str string) (string, error) {
	hashedtoken, err := bcrypt.GenerateFromPassword([]byte(str), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedtoken), nil
}

func (s auth) compareHash(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
