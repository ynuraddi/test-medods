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
	"medods/internal/service/user"
	"net"

	"medods/pkg/logger"
	"medods/pkg/smtp"
	"time"

	gjwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrValidationFailed = fmt.Errorf("fail to validate token")
)

const (
	// life time of access token
	ATokenLifetime = 30 * time.Minute
	// life time of refresh token
	RTokenLifeTime = 30 * 24 * time.Hour
)

type Interface interface {
	CreateSession(ctx context.Context, uid int, ip string) (aToken string, rToken string, err error)
	RefreshSession(ctx context.Context, aT, rT, ip string) (aToken string, rToken string, err error)
}

var _ Interface = (*auth)(nil)

type auth struct {
	session session.Interface
	user    user.Interface
	jwt     jwt.Interface
	smtp    smtp.Interface

	logger   logger.Interface
	testMode bool
}

func New(
	sessionService session.Interface,
	userService user.Interface,
	jwtMaker jwt.Interface,
	smtp smtp.Interface,
	logger logger.Interface,
	testMode bool,
) *auth {
	return &auth{
		session: sessionService,
		user:    userService,
		jwt:     jwtMaker,
		smtp:    smtp,

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

func (s auth) RefreshSession(ctx context.Context, aT, rT, ip string) (aToken, rToken string, err error) {
	_, payload, err := s.jwt.VerifyToken(aT)
	if err != nil && !errors.Is(err, gjwt.ErrTokenExpired) {
		err := fmt.Errorf("failed to verify access token: %w", err)
		s.logger.Error(err)
		return "", "", err
	}
	s.logger.Debug("success verified token")

	// dbSession.IP != payload.IP проверял до этого так, задался вопросом что это не имеет смылса только на интеграционных тестах)
	// перепрочитал и понял что нужно ip непосредственно получать и просто сверять с payload
	clientIP := net.ParseIP(ip)
	payloadIP := net.ParseIP(payload.IP)
	if !payloadIP.Equal(clientIP) {
		s.logger.Warn("login from new IP addess: old[%s], new[%s]", payload.IP, ip)

		dbUser, err := s.user.GetByID(ctx, payload.UserID)
		if err != nil {
			return "", "", err
		}

		if err := s.smtp.SendLoginFromNewIP(payload.IP, dbUser.Email); err != nil {
			return "", "", err
		}
	}

	dbSession, err := s.session.GetByUserID(ctx, payload.UserID)
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

	if !CompareHash(dbSession.RTokenHash, rT) {
		s.logger.Error(err)
		return "", "", ErrValidationFailed
	} else if payload.ID != dbSession.ATokenID {
		err := fmt.Errorf("%w: invalid jti", ErrValidationFailed)
		s.logger.Error(err)
		return "", "", err
	}

	now := time.Now()
	iat := payload.IssuedAt.Time
	rExp := iat.Add(RTokenLifeTime)
	if iat.Unix() != dbSession.CreatedAt {
		err := fmt.Errorf("different creation time of access and refresh token: %w", gjwt.ErrTokenExpired)
		s.logger.Error(err)
		return "", "", err
	} else if rExp.Before(now) {
		err := fmt.Errorf("refresh token: %w", gjwt.ErrTokenExpired)
		s.logger.Error(err)
		return "", "", err
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
			ExpiresAt: gjwt.NewNumericDate(iat.Add(ATokenLifetime)),
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

func CompareHash(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
