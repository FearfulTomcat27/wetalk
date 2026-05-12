//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wetalk/model"
)

func TestAPIAuth_Register_Success(t *testing.T) {
	resetDB(t)

	engine := newTestEngine()

	body := fmt.Sprintf(`{"username":"alice_%s","password":"pass123","nickname":"Alice"}`, t.Name())
	w := doRequest(t, engine, http.MethodPost, "/api/auth/register", strings.NewReader(body), nil)

	assert.Equal(t, http.StatusCreated, w.Code)
	resp := apiResult(t, w.Body.Bytes())
	require.NotNil(t, resp.Data)

	data, _ := json.Marshal(resp.Data)
	var authResp model.AuthResponse
	json.Unmarshal(data, &authResp)
	assert.NotEmpty(t, authResp.Token)
	assert.Equal(t, "alice_"+t.Name(), authResp.User.Username)
	assert.Equal(t, "Alice", authResp.User.Nickname)
}

func TestAPIAuth_Register_DuplicateUsername(t *testing.T) {
	resetDB(t)

	engine := newTestEngine()

	body := fmt.Sprintf(`{"username":"dup_%s","password":"pass123"}`, t.Name())
	w := doRequest(t, engine, http.MethodPost, "/api/auth/register", strings.NewReader(body), nil)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Register same username again
	w2 := doRequest(t, engine, http.MethodPost, "/api/auth/register", strings.NewReader(body), nil)
	assert.Equal(t, http.StatusConflict, w2.Code)
}

func TestAPIAuth_Register_InvalidParams(t *testing.T) {
	resetDB(t)

	engine := newTestEngine()

	// Empty username
	body := `{"username":"","password":"pass123"}`
	w := doRequest(t, engine, http.MethodPost, "/api/auth/register", strings.NewReader(body), nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Empty password
	body = fmt.Sprintf(`{"username":"valid_%s","password":""}`, t.Name())
	w = doRequest(t, engine, http.MethodPost, "/api/auth/register", strings.NewReader(body), nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAPIAuth_Login_Success(t *testing.T) {
	resetDB(t)

	engine := newTestEngine()

	// Register first
	regBody := fmt.Sprintf(`{"username":"login_%s","password":"pass123"}`, t.Name())
	w := doRequest(t, engine, http.MethodPost, "/api/auth/register", strings.NewReader(regBody), nil)
	require.Equal(t, http.StatusCreated, w.Code)

	// Login
	loginBody := fmt.Sprintf(`{"username":"login_%s","password":"pass123"}`, t.Name())
	w2 := doRequest(t, engine, http.MethodPost, "/api/auth/login", strings.NewReader(loginBody), nil)
	assert.Equal(t, http.StatusOK, w2.Code)

	resp := apiResult(t, w2.Body.Bytes())
	data, _ := json.Marshal(resp.Data)
	var authResp model.AuthResponse
	json.Unmarshal(data, &authResp)
	assert.NotEmpty(t, authResp.Token)
	assert.Equal(t, "login_"+t.Name(), authResp.User.Username)
}

func TestAPIAuth_Login_InvalidCredentials(t *testing.T) {
	resetDB(t)

	engine := newTestEngine()
	body := fmt.Sprintf(`{"username":"nobody_%s","password":"wrong"}`, t.Name())
	w := doRequest(t, engine, http.MethodPost, "/api/auth/login", strings.NewReader(body), nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAPIAuth_Me_Success(t *testing.T) {
	resetDB(t)

	engine := newTestEngine()
	authResp := registerUser(t, engine, "me_test_"+t.Name(), "pass123", "MeTest")

	w := doRequest(t, engine, http.MethodGet, "/api/me", nil, authHeaders(authResp.Token))
	assert.Equal(t, http.StatusOK, w.Code)

	resp := apiResult(t, w.Body.Bytes())
	data, _ := json.Marshal(resp.Data)
	var user model.User
	json.Unmarshal(data, &user)
	assert.Equal(t, authResp.User.ID, user.ID)
	assert.Equal(t, "me_test_"+t.Name(), user.Username)
}

func TestAPIAuth_Me_Unauthorized(t *testing.T) {
	resetDB(t)

	engine := newTestEngine()
	w := doRequest(t, engine, http.MethodGet, "/api/me", nil, nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAPIAuth_Ping(t *testing.T) {
	resetDB(t)

	engine := newTestEngine()
	w := doRequest(t, engine, http.MethodGet, "/ping", nil, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}
