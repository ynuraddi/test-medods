package integrational

import (
	"context"
	"encoding/json"
	"fmt"
	"medods/config"
	"medods/internal/repository"
	"medods/internal/service"
	"medods/internal/service/auth"
	"medods/pkg/logger"
	"medods/pkg/postgres"
	"medods/pkg/smtp"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var defaultConfig *config.Config = &config.Config{
	JWT: config.JWT{
		SecretKey: "123",
	},
}

func setupService(t *testing.T) (s *service.Manager, close func(), smtpEndpoint, apiEndpoint, psgEndpoint string) {
	smtpClose, smtpEndpoint, apiEndpoint, err := setupMailHog()
	assert.NoError(t, err)

	// log.Println("smtp container smtp endpoint: ", smtpEndpoint)
	// log.Println("smtp container api endpoint: ", apiEndpoint)

	_, stringPort, err := net.SplitHostPort(smtpEndpoint)
	assert.NoError(t, err)

	smtpPort, err := strconv.Atoi(strings.Split(stringPort, "/")[0])
	assert.NoError(t, err)

	smtp := smtp.New(&smtp.Config{
		Port: smtpPort,
		From: "medods@gmail.com",
	})

	psgClose, psgEndpoint, err := setupPostgres()
	assert.NoError(t, err)

	// log.Println("psg container endpoint: ", psgEndpoint)

	psg, err := postgres.New(&postgres.Config{
		DSN:          fmt.Sprintf("postgresql://user:1234@%s/medods?sslmode=disable", psgEndpoint),
		MigrationURL: "file://../migrations",
	})
	assert.NoError(t, err)

	err = psg.MigrateUP()
	assert.NoError(t, err)

	repo := repository.New(psg.Conn)
	service := service.New(defaultConfig, repo, smtp, logger.New("debug", true))

	return service, func() {
		smtpClose()
		psgClose()
	}, smtpEndpoint, apiEndpoint, psgEndpoint
}

func TestAuth_CreateSession(t *testing.T) {
	service, close, _, _, _ := setupService(t)
	defer close()

	ctx := context.Background()

	// Get users
	user, err := service.User.GetByID(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, "user@gmail.com", user.Email)

	// Create session
	IP1 := "::1"
	aT1, rT1, err := service.Auth.CreateSession(ctx, user.ID, IP1)
	assert.NoError(t, err)
	assert.NotEmpty(t, aT1)
	assert.NotEmpty(t, rT1)

	s1, err := service.Session.GetByUserID(ctx, user.ID)
	assert.NoError(t, err)

	_, p1, err := service.JWT.VerifyToken(aT1)
	assert.NoError(t, err)

	assert.Equal(t, s1.UserID, p1.UserID)
	assert.Equal(t, s1.IP, p1.IP)
	assert.Equal(t, s1.ATokenID, p1.ID)
	assert.Equal(t, s1.CreatedAt, p1.IssuedAt.Time.Unix())
	assert.True(t, auth.CompareHash(s1.RTokenHash, rT1))

	expTime := time.Now().Add(auth.ATokenLifetime) // if processing of container more than 5 minute mb flucky
	assert.True(t, expTime.Add(-5*time.Minute).Before(p1.ExpiresAt.Time))
	assert.True(t, expTime.Add(5*time.Minute).After(p1.ExpiresAt.Time))

	// Create new session
	IP2 := "::2"
	aT2, rT2, err := service.Auth.CreateSession(context.Background(), user.ID, IP2)
	assert.NoError(t, err)
	assert.NotEmpty(t, aT2)
	assert.NotEmpty(t, rT2)
	assert.NotEqual(t, aT1, aT2)
	assert.NotEqual(t, rT1, rT2)

	s2, err := service.Session.GetByUserID(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, s1.ID, s2.ID)
	assert.Equal(t, s1.UserID, s2.UserID)
	assert.Equal(t, s1.Version+1, s2.Version)
	assert.NotEqual(t, s1.ATokenID, s2.ATokenID)
	assert.NotEqual(t, s1.RTokenHash, s2.RTokenHash)
	assert.NotEqual(t, s1.IP, s2.IP)
}

func TestAuth_RefreshSession(t *testing.T) {
	service, close, _, apiEndpoint, _ := setupService(t)
	defer close()

	ctx := context.Background()

	// Get users
	user, err := service.User.GetByID(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, "user@gmail.com", user.Email)

	// Create session
	IP1 := "::1"
	aT1, rT1, err := service.Auth.CreateSession(ctx, user.ID, IP1)
	assert.NoError(t, err)
	assert.NotEmpty(t, aT1)
	assert.NotEmpty(t, rT1)

	s1, err := service.Session.GetByUserID(ctx, user.ID)
	assert.NoError(t, err)

	_, p1, err := service.JWT.VerifyToken(aT1)
	assert.NoError(t, err)

	// Refresh session
	aT2, rT2, err := service.Auth.RefreshSession(ctx, aT1, rT1, IP1)
	assert.NoError(t, err)
	assert.NotEmpty(t, aT2)
	assert.NotEmpty(t, rT2)
	assert.NotEqual(t, aT1, aT2)
	assert.NotEqual(t, rT1, rT2)

	s2, err := service.Session.GetByUserID(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, s1.ID, s2.ID)
	assert.Equal(t, s1.IP, s2.IP)
	assert.Equal(t, s1.UserID, s2.UserID)
	assert.Equal(t, s1.Version+1, s2.Version)
	assert.NotEqual(t, s1.ATokenID, s2.ATokenID)
	assert.NotEqual(t, s1.RTokenHash, s2.RTokenHash)

	_, p2, err := service.JWT.VerifyToken(aT2)
	assert.NoError(t, err)
	assert.Equal(t, p1.IP, p2.IP)
	assert.Equal(t, p1.UserID, p2.UserID)
	assert.NotEqual(t, p1.ID, p2.ID)

	// Try refrsh with old aToken and old rToken
	aT3, rT3, err := service.Auth.RefreshSession(ctx, aT1, rT1, IP1)
	assert.Error(t, err)
	assert.Empty(t, aT3)
	assert.Empty(t, rT3)

	// Try refrsh with new aToken and old rToken
	// check that i can't refresh session even i have new aT
	aT3, rT3, err = service.Auth.RefreshSession(ctx, aT1, rT2, IP1)
	assert.Error(t, err)
	assert.Empty(t, aT3)
	assert.Empty(t, rT3)

	// Try refresh with old Token and new rToken
	// aToken also valid, so we need check strong link between at and rt
	// we check that i can't use valid aT1 with valid rT2
	aT3, rT3, err = service.Auth.RefreshSession(ctx, aT1, rT2, IP1)
	assert.Error(t, err)
	assert.Empty(t, aT3)
	assert.Empty(t, rT3)

	// check count of messages of smtp
	// ref: https://github.com/mailhog/MailHog/blob/master/docs/APIv2/swagger-2.0.yaml#L8
	getLenSmtpMessages := func(t *testing.T, apiEndpoint string) int {
		smtpMessagesPath := fmt.Sprintf("http://%s/api/v1/messages", apiEndpoint)
		res, err := http.Get(smtpMessagesPath)
		assert.NoError(t, err)
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)

		var msgs []interface{}
		err = json.NewDecoder(res.Body).Decode(&msgs)
		assert.NoError(t, err)

		return len(msgs)
	}

	// Refresh from new ip
	counOfMsgsBefore := getLenSmtpMessages(t, apiEndpoint)

	IP2 := "::2"
	aT3, rT3, err = service.Auth.RefreshSession(ctx, aT2, rT2, IP2)
	assert.NoError(t, err) // in my service we just notify user about login from new ip
	assert.NotEmpty(t, aT3)
	assert.NotEmpty(t, rT3)
	assert.True(t, aT3 != aT1 && aT3 != aT2)
	assert.True(t, rT3 != rT1 && rT3 != rT2)

	countOfMsgsAfter := getLenSmtpMessages(t, apiEndpoint)
	assert.Equal(t, counOfMsgsBefore+1, countOfMsgsAfter)
}
