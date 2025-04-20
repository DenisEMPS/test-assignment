package unit

import (
	"testing"
	"time"

	"github.com/DenisEMPS/test-assignment/internal/domain"
	"github.com/DenisEMPS/test-assignment/internal/service/auth"
	mocked "github.com/DenisEMPS/test-assignment/internal/service/auth/mocks"
	"github.com/DenisEMPS/test-assignment/tests/suite"
	"github.com/dgrijalva/jwt-go"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAccessToken(t *testing.T) {
	st := suite.New(t)

	c := gomock.NewController(t)
	defer c.Finish()

	mockedRepo := mocked.NewMockAuthRepository(c)
	service := auth.New(mockedRepo, st.Log, &st.Cfg.JWT)

	userID := uuid.New()
	accessID := uuid.New()
	userIP := "127.0.0.1:1234"

	testTable := []struct {
		name       string
		tokenFn    func() string
		wantErr    bool
		expectType error
	}{
		{
			name: "Happy path",
			tokenFn: func() string {
				token, err := service.GenerateAccessToken(userID, accessID, userIP, time.Minute*15)
				require.NoError(t, err)
				return token
			},
		},
		{
			name: "Expired token",
			tokenFn: func() string {
				token, err := service.GenerateAccessToken(userID, accessID, userIP, -time.Minute)
				require.NoError(t, err)
				return token
			},
			wantErr:    true,
			expectType: auth.ErrTokenExpired,
		},
		{
			name: "Invalid format",
			tokenFn: func() string {
				return "not.a.valid.jwt"
			},
			wantErr:    true,
			expectType: auth.ErrInvalidToken,
		},
		{
			name: "Invalid signing method",
			tokenFn: func() string {
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
					"user_id": userID.String(),
				})
				tokenString, _ := token.SignedString([]byte("notused"))
				return tokenString
			},
			wantErr:    true,
			expectType: auth.ErrInvalidToken,
		},
		{
			name: "Invalid secret-key",
			tokenFn: func() string {
				token := jwt.NewWithClaims(jwt.SigningMethodHS512, domain.UserClaims{
					UserID:     userID,
					UserIP:     userIP,
					AccessUUID: accessID,
					StandardClaims: jwt.StandardClaims{
						IssuedAt:  time.Now().Unix(),
						ExpiresAt: time.Now().Add(st.Cfg.JWT.AccessTokenTTL).Unix(),
					},
				})

				tokenString, err := token.SignedString([]byte("random2143edrfgdfhgfd"))
				if err != nil {
					require.NoError(t, err)
				}

				return tokenString
			},
			wantErr:    true,
			expectType: auth.ErrInvalidToken,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			token := testCase.tokenFn()

			claims, err := service.ParseAccessToken(token)

			if testCase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, testCase.expectType)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, userID, claims.UserID)
			assert.Equal(t, userIP, claims.UserIP)
			assert.Equal(t, accessID, claims.AccessUUID)
		})
	}
}
