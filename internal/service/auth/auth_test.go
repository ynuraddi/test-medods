package auth

import (
	"context"
	"database/sql"
	"fmt"
	"medods/internal/model"
	mock_jwt "medods/internal/service/jwt/mock"
	mock_session "medods/internal/service/session/mock"
	mock_user "medods/internal/service/user/mock"
	"medods/pkg/logger"
	mock_smtp "medods/pkg/smtp/mock"
	"reflect"
	"time"

	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type sessionMatcher struct {
	session         model.Session
	compareHashFunc func(hash, plain string) bool
}

func (m sessionMatcher) Matches(x interface{}) bool {
	input, ok := x.(model.Session)
	if !ok {
		return false
	}

	// if session field is not empty that is required

	if m.session.ID != 0 && m.session.ID != input.ID {
		return false
	}
	if m.session.UserID != 0 && m.session.UserID != input.UserID {
		return false
	}
	if m.session.ATokenID != "" && m.session.ATokenID != input.ATokenID {
		return false
	}
	// use for compare non hashed token
	if m.session.RTokenHash != "" && !m.compareHashFunc(input.RTokenHash, m.session.RTokenHash) {
		return false
	}
	if m.session.IP != "" && m.session.IP != input.IP {
		return false
	}
	if m.session.CreatedAt != 0 && m.session.CreatedAt != input.CreatedAt {
		return false
	}
	if m.session.Version != 0 && m.session.Version != input.Version {
		return false
	}

	return true
}

func (m sessionMatcher) String() string {
	return fmt.Sprintf("%v (%v)", m.session, reflect.TypeOf(m.session))
}

type payloadMatcher struct {
	payload model.Payload
}

func (m payloadMatcher) Matches(x interface{}) bool {
	input, ok := x.(model.Payload)
	if !ok {
		return false
	}

	if m.payload.UserID > 0 && m.payload.UserID != input.UserID {
		return false
	}
	if m.payload.IP != "" && m.payload.IP != input.IP {
		return false
	}
	if m.payload.RegisteredClaims.ID != "" && m.payload.RegisteredClaims.ID != input.RegisteredClaims.ID {
		return false
	}

	return true
}

func (m payloadMatcher) String() string {
	return fmt.Sprintf("%v (%v)", m.payload, reflect.TypeOf(m.payload))
}

func TestCreateSession(t *testing.T) {
	ctrl := gomock.NewController(t)

	sessionService := mock_session.NewMockInterface(ctrl)
	user := mock_user.NewMockInterface(ctrl)
	jwtMaker := mock_jwt.NewMockInterface(ctrl)
	logger := logger.New("debug", true)
	smtp := mock_smtp.NewMockInterface(ctrl)

	auth := New(sessionService, user, jwtMaker, smtp, logger, true)

	defaultATokenID := auth.generateUUID()            // means than uuid always generate than string when testMode is truw
	defaultRTokenRandString, err := auth.randString() // like uuid
	assert.NoError(t, err)

	defaultAToken := "access_token"
	defaultRToken := "rand_string"

	iat := time.Now().Add(-1 * time.Minute)

	unexpectedError := fmt.Errorf("unexpected error")

	type args struct {
		uid int
		ip  string
	}

	defaultIP := "2"

	defaultInput := args{
		uid: 1,
		ip:  defaultIP,
	}

	tc := []struct {
		name        string
		input       args
		buildStubs  func()
		checkResult func(t *testing.T, aToken, rToken string, err error)
	}{
		{
			name:  "OK with existing",
			input: defaultInput,
			buildStubs: func() {
				jwtMaker.EXPECT().CreateToken(payloadMatcher{model.Payload{
					UserID:           defaultInput.uid,
					IP:               defaultInput.ip,
					RegisteredClaims: jwt.RegisteredClaims{ID: defaultATokenID},
				}}).Times(1).Return(defaultAToken, nil)

				dbSession := model.Session{
					ID:         1,
					UserID:     1,
					ATokenID:   "other",
					RTokenHash: "other",
					IP:         "other",
					CreatedAt:  iat.Unix(),
					Version:    1,
				}

				sessionService.EXPECT().GetByUserID(gomock.Any(), gomock.Eq(defaultInput.uid)).Times(1).
					Return(dbSession, nil)

				dbSession.ATokenID = defaultATokenID
				dbSession.RTokenHash = defaultRTokenRandString // use rand_string for compareHash
				dbSession.IP = defaultIP                       // check that ip is from input
				dbSession.CreatedAt = 0                        // don't check time

				sessionService.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)

				sessionService.EXPECT().Update(gomock.Any(), sessionMatcher{dbSession, CompareHash}).Times(1).
					Return(model.Session{}, nil) // first field is not matter cause we use '_, err'
			},
			checkResult: func(t *testing.T, aToken, rToken string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, defaultAToken, aToken)
				assert.Equal(t, defaultRToken, rToken)
			},
		},
		{
			name:  "update session error",
			input: defaultInput,
			buildStubs: func() {
				jwtMaker.EXPECT().CreateToken(payloadMatcher{model.Payload{
					UserID:           defaultInput.uid,
					IP:               defaultInput.ip,
					RegisteredClaims: jwt.RegisteredClaims{ID: defaultATokenID},
				}}).Times(1).Return(defaultAToken, nil)

				dbSession := model.Session{
					ID:         1,
					UserID:     1,
					ATokenID:   "other",
					RTokenHash: "other",
					IP:         "other",
					CreatedAt:  iat.Unix(),
					Version:    1,
				}

				sessionService.EXPECT().GetByUserID(gomock.Any(), gomock.Eq(defaultInput.uid)).Times(1).
					Return(dbSession, nil)

				dbSession.ATokenID = defaultATokenID
				dbSession.RTokenHash = defaultRTokenRandString // use rand_string for compareHash
				dbSession.IP = defaultIP                       // check that ip is from input
				dbSession.CreatedAt = 0                        // don't check time

				sessionService.EXPECT().Update(gomock.Any(), sessionMatcher{dbSession, CompareHash}).Times(1).
					Return(model.Session{}, unexpectedError) // return unexpected error
			},
			checkResult: func(t *testing.T, aToken, rToken string, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, unexpectedError)
				assert.Empty(t, aToken)
				assert.Empty(t, rToken)
			},
		},
		{
			name:  "OK with new",
			input: defaultInput,
			buildStubs: func() {
				jwtMaker.EXPECT().CreateToken(payloadMatcher{model.Payload{
					UserID:           defaultInput.uid,
					IP:               defaultInput.ip,
					RegisteredClaims: jwt.RegisteredClaims{ID: defaultATokenID},
				}}).Times(1).Return(defaultAToken, nil)

				sessionService.EXPECT().GetByUserID(gomock.Any(), gomock.Eq(defaultInput.uid)).Times(1).
					Return(model.Session{}, sql.ErrNoRows)

				checkCreateInput := model.Session{
					UserID:     1,
					ATokenID:   defaultATokenID,
					RTokenHash: defaultRTokenRandString,
					IP:         defaultIP,
				}

				sessionService.EXPECT().Update(gomock.Any(), gomock.Any()).Times(0)
				sessionService.EXPECT().Create(gomock.Any(), sessionMatcher{checkCreateInput, CompareHash}).Times(1).Return(nil)
			},
			checkResult: func(t *testing.T, aToken, rToken string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, defaultAToken, aToken)
				assert.Equal(t, defaultRToken, rToken)
			},
		},
		{
			name:  "create session error",
			input: defaultInput,
			buildStubs: func() {
				jwtMaker.EXPECT().CreateToken(payloadMatcher{model.Payload{
					UserID:           defaultInput.uid,
					IP:               defaultInput.ip,
					RegisteredClaims: jwt.RegisteredClaims{ID: defaultATokenID},
				}}).Times(1).Return(defaultAToken, nil)

				sessionService.EXPECT().GetByUserID(gomock.Any(), gomock.Eq(defaultInput.uid)).Times(1).
					Return(model.Session{}, sql.ErrNoRows)

				checkCreateInput := model.Session{
					UserID:     1,
					ATokenID:   defaultATokenID,
					RTokenHash: defaultRTokenRandString,
					IP:         defaultIP,
				}

				sessionService.EXPECT().Update(gomock.Any(), gomock.Any()).Times(0)
				sessionService.EXPECT().Create(gomock.Any(), sessionMatcher{checkCreateInput, CompareHash}).Times(1).Return(unexpectedError)
			},
			checkResult: func(t *testing.T, aToken, rToken string, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, unexpectedError)
				assert.Empty(t, aToken)
				assert.Empty(t, rToken)
			},
		},
		{
			name:  "get session by user id unexpected error",
			input: defaultInput,
			buildStubs: func() {
				jwtMaker.EXPECT().CreateToken(payloadMatcher{model.Payload{
					UserID:           defaultInput.uid,
					IP:               defaultInput.ip,
					RegisteredClaims: jwt.RegisteredClaims{ID: defaultATokenID},
				}}).Times(1).Return(defaultAToken, nil)

				sessionService.EXPECT().GetByUserID(gomock.Any(), gomock.Eq(defaultInput.uid)).Times(1).
					Return(model.Session{}, unexpectedError)

				sessionService.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)
				sessionService.EXPECT().Update(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResult: func(t *testing.T, aToken, rToken string, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, unexpectedError)
				assert.Empty(t, aToken)
				assert.Empty(t, rToken)
			},
		},
		{
			name:  "jwt create token error",
			input: defaultInput,
			buildStubs: func() {
				jwtMaker.EXPECT().CreateToken(payloadMatcher{model.Payload{
					UserID:           defaultInput.uid,
					IP:               defaultInput.ip,
					RegisteredClaims: jwt.RegisteredClaims{ID: defaultATokenID},
				}}).Times(1).Return("", unexpectedError)

				sessionService.EXPECT().GetByUserID(gomock.Any(), gomock.Any()).Times(0)
				sessionService.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)
				sessionService.EXPECT().Update(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResult: func(t *testing.T, aToken, rToken string, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, unexpectedError)
				assert.Empty(t, aToken)
				assert.Empty(t, rToken)
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			test.buildStubs()
			aT, rT, err := auth.CreateSession(context.Background(), test.input.uid, test.input.ip)
			test.checkResult(t, aT, rT, err)
		})
	}
}

func TestRefreshSession(t *testing.T) {
	ctrl := gomock.NewController(t)

	sessionService := mock_session.NewMockInterface(ctrl)
	userService := mock_user.NewMockInterface(ctrl)
	jwtMaker := mock_jwt.NewMockInterface(ctrl)
	logger := logger.New("debug", true)
	smtpService := mock_smtp.NewMockInterface(ctrl)

	auth := New(sessionService, userService, jwtMaker, smtpService, logger, true)

	defaultATokenID := auth.generateUUID()
	defaultRTokenRandString, err := auth.randString()
	assert.NoError(t, err)

	defaultAToken := "access_token"
	defaultRToken := "rand_string"

	defaultMail := "mock@gmail.com"

	iat := time.Now().Add(-1 * time.Minute)
	exp := iat.Add(ATokenLifetime)

	unexpectedError := fmt.Errorf("unexpected error")

	defaultIP := "::1"

	defaultPayload := model.Payload{
		UserID: 1,
		IP:     defaultIP,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(iat),
			ID:        defaultATokenID,
		},
	}

	defaultRTokenHash, err := auth.hashString(defaultRTokenRandString)
	assert.NoError(t, err)

	defaultSession := model.Session{
		ID:         1,
		UserID:     1,
		ATokenID:   defaultATokenID,
		RTokenHash: defaultRTokenHash,
		IP:         defaultIP,
		CreatedAt:  iat.Unix(),
		Version:    1,
	}

	callCreateSession := func(uid int, ip string, aTID, rTHash string) {
		jwtMaker.EXPECT().CreateToken(payloadMatcher{model.Payload{
			UserID: uid,
			IP:     ip,
		}}).Times(1).Return(defaultAToken, nil)
		sessionService.EXPECT().GetByUserID(gomock.Any(), gomock.Eq(uid)).Times(1).
			Return(model.Session{}, sql.ErrNoRows)

		checkCreateInput := model.Session{
			UserID:     1,
			ATokenID:   aTID,
			RTokenHash: rTHash,
			IP:         ip,
		}

		sessionService.EXPECT().Update(gomock.Any(), gomock.Any()).Times(0)
		sessionService.EXPECT().Create(gomock.Any(), sessionMatcher{checkCreateInput, CompareHash}).Times(1).Return(nil)
	}

	type args struct {
		aToken string
		rToken string
		ip     string
	}

	defaultInput := args{
		aToken: defaultAToken,
		rToken: defaultRToken,
		ip:     defaultIP,
	}

	tc := []struct {
		name        string
		input       args
		buildStubs  func()
		checkResult func(t *testing.T, aToken, rToken string, err error)
	}{
		{
			name:  "OK",
			input: defaultInput,
			buildStubs: func() {
				jwtMaker.EXPECT().VerifyToken(gomock.Eq(defaultAToken)).Times(1).Return(nil, &defaultPayload, nil)
				sessionService.EXPECT().GetByUserID(gomock.Any(), gomock.Eq(defaultPayload.UserID)).Times(1).Return(defaultSession, nil)

				callCreateSession(defaultPayload.UserID, defaultPayload.IP, defaultATokenID, defaultRTokenRandString) // use rand_string cause it's a plain of compare func
			},
			checkResult: func(t *testing.T, aToken, rToken string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, defaultAToken, aToken)
				assert.Equal(t, defaultRToken, rToken)
			},
		},
		{
			name:  "OK with expired aToken",
			input: defaultInput,
			buildStubs: func() {
				jwtMaker.EXPECT().VerifyToken(gomock.Eq(defaultAToken)).Times(1).Return(nil, &defaultPayload, jwt.ErrTokenExpired) // note: expired token
				sessionService.EXPECT().GetByUserID(gomock.Any(), gomock.Eq(defaultPayload.UserID)).Times(1).Return(defaultSession, nil)

				callCreateSession(defaultPayload.UserID, defaultPayload.IP, defaultATokenID, defaultRTokenRandString) // use rand_string cause it's a plain of compare func
			},
			checkResult: func(t *testing.T, aToken, rToken string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, defaultAToken, aToken)
				assert.Equal(t, defaultRToken, rToken)
			},
		},
		{
			name:  "error verify token",
			input: defaultInput,
			buildStubs: func() {
				jwtMaker.EXPECT().VerifyToken(gomock.Eq(defaultAToken)).Times(1).Return(nil, &defaultPayload, unexpectedError)
				sessionService.EXPECT().GetByUserID(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResult: func(t *testing.T, aToken, rToken string, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, unexpectedError)
				assert.Empty(t, aToken)
				assert.Empty(t, rToken)
			},
		},
		{
			name:  "error unexpected get session by user id",
			input: defaultInput,
			buildStubs: func() {
				jwtMaker.EXPECT().VerifyToken(gomock.Eq(defaultAToken)).Times(1).Return(nil, &defaultPayload, nil)
				sessionService.EXPECT().GetByUserID(gomock.Any(), gomock.Eq(defaultPayload.UserID)).Times(1).Return(model.Session{}, unexpectedError)
			},
			checkResult: func(t *testing.T, aToken, rToken string, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, unexpectedError)
				assert.Empty(t, aToken)
				assert.Empty(t, rToken)
			},
		},
		{
			name:  "error no exists get session by user id",
			input: defaultInput,
			buildStubs: func() {
				jwtMaker.EXPECT().VerifyToken(gomock.Eq(defaultAToken)).Times(1).Return(nil, &defaultPayload, nil)
				sessionService.EXPECT().GetByUserID(gomock.Any(), gomock.Eq(defaultPayload.UserID)).Times(1).Return(model.Session{}, sql.ErrNoRows)
			},
			checkResult: func(t *testing.T, aToken, rToken string, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, sql.ErrNoRows)
				assert.Empty(t, aToken)
				assert.Empty(t, rToken)
			},
		},
		{
			name:  "error compare different access iat and db iat times",
			input: defaultInput,
			buildStubs: func() {
				jwtMaker.EXPECT().VerifyToken(gomock.Eq(defaultAToken)).Times(1).Return(nil, &defaultPayload, nil)

				copySession := defaultSession
				copySession.CreatedAt = iat.Add(1 * time.Minute).Unix()

				sessionService.EXPECT().GetByUserID(gomock.Any(), gomock.Eq(defaultPayload.UserID)).Times(1).Return(copySession, nil)
			},
			checkResult: func(t *testing.T, aToken, rToken string, err error) {
				assert.Error(t, err)
				assert.Empty(t, aToken)
				assert.Empty(t, rToken)
			},
		},
		{
			name:  "error compare refresh token expired",
			input: defaultInput,
			buildStubs: func() {
				iat := time.Now().Add(-RTokenLifeTime).Add(-1 * time.Minute) // iat more than refresh token life
				accessExp := iat.Add(ATokenLifetime)

				cpPayload := defaultPayload
				cpPayload.IssuedAt = jwt.NewNumericDate(iat)
				cpPayload.ExpiresAt = jwt.NewNumericDate(accessExp)

				jwtMaker.EXPECT().VerifyToken(gomock.Eq(defaultAToken)).Times(1).Return(nil, &cpPayload, nil)

				cpSession := defaultSession
				cpSession.CreatedAt = iat.Unix()

				sessionService.EXPECT().GetByUserID(gomock.Any(), gomock.Eq(defaultPayload.UserID)).Times(1).Return(cpSession, nil)
			},
			checkResult: func(t *testing.T, aToken, rToken string, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, jwt.ErrTokenExpired)
				assert.Empty(t, aToken)
				assert.Empty(t, rToken)
			},
		},
		{
			name: "error compare refresh token hash",
			input: args{
				aToken: defaultAToken,
				rToken: "other",
				ip:     defaultIP,
			},
			buildStubs: func() {
				jwtMaker.EXPECT().VerifyToken(gomock.Eq(defaultAToken)).Times(1).Return(nil, &defaultPayload, nil)
				sessionService.EXPECT().GetByUserID(gomock.Any(), gomock.Eq(defaultPayload.UserID)).Times(1).Return(defaultSession, nil)
			},
			checkResult: func(t *testing.T, aToken, rToken string, err error) {
				assert.Error(t, err)
				assert.Empty(t, aToken)
				assert.Empty(t, rToken)
			},
		},
		{
			name:  "error compare payload and db jti",
			input: defaultInput,
			buildStubs: func() {
				cpPayload := defaultPayload
				cpPayload.ID = "other"

				jwtMaker.EXPECT().VerifyToken(gomock.Eq(defaultAToken)).Times(1).Return(nil, &cpPayload, nil)
				sessionService.EXPECT().GetByUserID(gomock.Any(), gomock.Eq(defaultPayload.UserID)).Times(1).Return(defaultSession, nil)
			},
			checkResult: func(t *testing.T, aToken, rToken string, err error) {
				assert.Error(t, err)
				assert.Empty(t, aToken)
				assert.Empty(t, rToken)
			},
		},
		{
			name:  "OK login from new ip",
			input: defaultInput,
			buildStubs: func() {
				cpPayload := defaultPayload
				cpPayload.IP = "::2"

				jwtMaker.EXPECT().VerifyToken(gomock.Eq(defaultAToken)).Times(1).Return(nil, &cpPayload, nil)                       // note return ip 'other'
				sessionService.EXPECT().GetByUserID(gomock.Any(), gomock.Eq(cpPayload.UserID)).Times(1).Return(defaultSession, nil) // note return default ip

				userService.EXPECT().GetByID(gomock.Any(), gomock.Eq(cpPayload.UserID)).Times(1).
					Return(model.User{ID: cpPayload.UserID, Email: defaultMail}, nil)
				smtpService.EXPECT().SendLoginFromNewIP(gomock.Eq(cpPayload.IP), gomock.Eq(defaultMail)).Times(1).Return(nil)

				callCreateSession(cpPayload.UserID, cpPayload.IP, defaultATokenID, defaultRTokenRandString)
			},
			checkResult: func(t *testing.T, aToken, rToken string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, defaultAToken, aToken)
				assert.Equal(t, defaultRToken, rToken)
			},
		},
		{
			name:  "error unexpected get user by id",
			input: defaultInput,
			buildStubs: func() {
				cpPayload := defaultPayload
				cpPayload.IP = "::2"

				jwtMaker.EXPECT().VerifyToken(gomock.Eq(defaultAToken)).Times(1).Return(nil, &cpPayload, nil) // note return ip 'other'
				userService.EXPECT().GetByID(gomock.Any(), gomock.Eq(cpPayload.UserID)).Times(1).Return(model.User{}, unexpectedError)
			},
			checkResult: func(t *testing.T, aToken, rToken string, err error) {
				assert.Error(t, err)
				assert.Empty(t, aToken)
				assert.Empty(t, rToken)
			},
		},
		{
			name:  "error unexpected send email",
			input: defaultInput,
			buildStubs: func() {
				cpPayload := defaultPayload
				cpPayload.IP = "::2"

				jwtMaker.EXPECT().VerifyToken(gomock.Eq(defaultAToken)).Times(1).Return(nil, &cpPayload, nil) // note return ip 'other'
				userService.EXPECT().GetByID(gomock.Any(), gomock.Eq(cpPayload.UserID)).Times(1).
					Return(model.User{ID: cpPayload.UserID, Email: defaultMail}, nil)
				smtpService.EXPECT().SendLoginFromNewIP(gomock.Eq(cpPayload.IP), gomock.Eq(defaultMail)).Times(1).Return(unexpectedError)
			},
			checkResult: func(t *testing.T, aToken, rToken string, err error) {
				assert.Error(t, err)
				assert.Empty(t, aToken)
				assert.Empty(t, rToken)
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			test.buildStubs()
			aT, rT, err := auth.RefreshSession(context.Background(), test.input.aToken, test.input.rToken, test.input.ip)
			test.checkResult(t, aT, rT, err)
		})
	}
}
