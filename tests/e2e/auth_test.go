package e2e

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"testing"

	httppkg "net/http"

	"github.com/DenisEMPS/test-assignment/internal/delivery/http"
	"github.com/DenisEMPS/test-assignment/internal/domain"
	"github.com/DenisEMPS/test-assignment/internal/repository/postgres"
	"github.com/DenisEMPS/test-assignment/internal/service/auth"
	"github.com/DenisEMPS/test-assignment/tests/suite"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ip = "127.0.0.1:1234"

func TestAuthHappyPathAndReuse(t *testing.T) {
	st := suite.New(t)

	db, err := postgres.NewDB(postgres.Config{
		Host:     st.Cfg.Postgres.Host,
		Port:     st.Cfg.Postgres.Port,
		User:     st.Cfg.Postgres.Username,
		Password: st.Cfg.Postgres.Password,
		DBName:   st.Cfg.Postgres.DBname,
		SSLMode:  st.Cfg.Postgres.SSLmode,
	})
	require.NoError(t, err)

	repo := postgres.NewAuth(db)
	service := auth.New(repo, st.Log, &st.Cfg.JWT)
	handler := http.NewHandler(service)

	r := setupRouter(handler)

	validTokenPair := domain.TokenPairResponse{}
	invalidTokenPair := domain.TokenPairResponse{}
	refresedTokenPair := domain.TokenPairResponse{}
	userID := uuid.New()

	t.Run("Generate tokens", func(t *testing.T) {
		req, err := newGenerateTokenHTTPRequest(userID.String())
		require.NoError(t, err)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &validTokenPair))
		assert.Equal(t, httppkg.StatusOK, w.Code)
		require.NotEmpty(t, validTokenPair.Access)
		require.NotEmpty(t, validTokenPair.Refresh)
	})

	t.Run("Refresh invalid tokens pair", func(t *testing.T) {
		invalidTokenPair.Access = validTokenPair.Access
		invalidTokenPair.Refresh = base64.URLEncoding.EncodeToString([]byte(uuid.NewString()))

		body, err := json.Marshal(invalidTokenPair)
		require.NoError(t, err)

		req, err := newRefreshTokenHTTPRequest(string(body))
		require.NoError(t, err)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, httppkg.StatusUnauthorized, w.Code)
		assert.Equal(t, `{"message":"unauthorized"}`, w.Body.String())
	})

	t.Run("Refresh valid tokens", func(t *testing.T) {
		body, err := json.Marshal(validTokenPair)
		require.NoError(t, err)

		req, err := newRefreshTokenHTTPRequest(string(body))
		require.NoError(t, err)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &refresedTokenPair))
		assert.Equal(t, httppkg.StatusOK, w.Code)
		require.NotEmpty(t, refresedTokenPair.Access)
		require.NotEmpty(t, refresedTokenPair.Refresh)
		require.NotEqual(t, refresedTokenPair.Access, validTokenPair.Access)
		require.NotEqual(t, refresedTokenPair.Refresh, validTokenPair.Refresh)
	})

	t.Run("Refresh tokens reuse", func(t *testing.T) {
		body, err := json.Marshal(validTokenPair)
		require.NoError(t, err)

		req, err := newRefreshTokenHTTPRequest(string(body))
		require.NoError(t, err)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, httppkg.StatusUnauthorized, w.Code)
		assert.Equal(t, `{"message":"unauthorized"}`, w.Body.String())
	})
}

func setupRouter(handler *http.Handler) *gin.Engine {
	r := gin.Default()
	auth := r.Group("/auth")
	{
		auth.POST("/generate", handler.GenerateTokens)
		auth.POST("/refresh", handler.RefreshTokens)
	}
	return r
}

func newGenerateTokenHTTPRequest(userID string) (*httppkg.Request, error) {
	req, err := httppkg.NewRequest("POST", "/auth/generate?user_id="+userID, nil)
	if err != nil {
		return nil, err
	}
	req.RemoteAddr = ip
	return req, nil
}

func newRefreshTokenHTTPRequest(body string) (*httppkg.Request, error) {
	req, err := httppkg.NewRequest("POST", "/auth/refresh", bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}
	req.RemoteAddr = ip
	return req, nil
}
