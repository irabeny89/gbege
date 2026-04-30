package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/irabeny89/gbege/internal/api"
	"github.com/irabeny89/gbege/internal/logger"
	"github.com/irabeny89/gbege/internal/security"
	"github.com/irabeny89/gbege/internal/session"
	"github.com/irabeny89/gbege/internal/user"
	"github.com/irabeny89/gosqlitex"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type LoginResponse struct {
	Token string     `json:"token"`
	User  *user.User `json:"user"`
}
type SignUpRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Validate validates the signup request
func (l *SignUpRequest) validate() error {
	if l.Username == "" {
		return errors.New("username is required")
	}
	if l.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

// Validate validates the login request
func (l *LoginRequest) validate() error {
	if l.Username == "" {
		return errors.New("username is required")
	}
	if l.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

// MARK: - Login
// HandleLogin handles login requests
func HandleLogin(db *gosqlitex.DbClient, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.Log.Error("Method not allowed", "method", r.Method)
		api.Fail(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Error("Failed to decode request body", "err", err)
		api.Fail(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := req.validate(); err != nil {
		logger.Log.Error("Failed to validate request body", "err", err)
		api.Fail(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	u, err := user.GetUserByUsername(db, req.Username)
	if err != nil {
		logger.Log.Error("User not found", "username", req.Username, "err", err)
		api.Fail(w, http.StatusUnauthorized, "Invalid credentials", err)
		return
	}

	ok, err := security.VerifyPassword(req.Password, u.Password)
	if err != nil || !ok {
		logger.Log.Error("Invalid password", "username", req.Username, "err", err)
		api.Fail(w, http.StatusUnauthorized, "Invalid credentials", err)
		return
	}

	ses, err := session.SaveSession(db, int(u.Id))
	if err != nil {
		logger.Log.Error("Failed to create session", "user_id", u.Id, "err", err)
		api.Fail(w, http.StatusInternalServerError, "Failed to create session", err)
		return
	}

	api.Success(w, http.StatusOK, "Login successful", LoginResponse{
		Token: base64.StdEncoding.EncodeToString(ses.Id),
		User:  u,
	})
}

// HandleSignUp handles signup requests
func HandleSignUp(db *gosqlitex.DbClient, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.Log.Error("Method not allowed", "method", r.Method)
		api.Fail(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	var req SignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Error("Failed to decode request body", "err", err)
		api.Fail(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := req.validate(); err != nil {
		logger.Log.Error("Failed to validate request body", "err", err)
		api.Fail(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	u, err := user.GetUserByUsername(db, req.Username)
	if err != nil {
		logger.Log.Error("User not found", "username", req.Username, "err", err)
		api.Fail(w, http.StatusUnauthorized, "Invalid credentials", err)
		return
	}

	if u != nil {
		logger.Log.Error("User already exists", "username", req.Username)
		api.Fail(w, http.StatusUnauthorized, "User already exists", nil)
		return
	}

	u, err = user.SaveUser(db, req.Username, req.Password)
	if err != nil {
		logger.Log.Error("Failed to create user", "username", req.Username, "err", err)
		api.Fail(w, http.StatusInternalServerError, "Failed to create user", err)
		return
	}

	api.Success(w, http.StatusOK, "User created successfully", u)
}

// HandleLogout handles logout requests
func HandleLogout(db *gosqlitex.DbClient, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.Log.Error("Method not allowed", "method", r.Method)
		api.Fail(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}
	_, token, ok := strings.Cut(r.Header.Get("authorization"), " ")
	if !ok {
		logger.Log.Error("Invalid authorization header")
		api.Fail(w, http.StatusUnauthorized, "Unauthorized", nil)
	}
	err := session.DeleteSession(db, []byte(token))
	if err != nil {
		logger.Log.Error("Failed to delete session")
		api.Fail(w, http.StatusInternalServerError, "Failed to delete session", err)
	}
	api.Success(w, http.StatusOK, "Logout successful", nil)
}

// HandleMe handles requests for the current user
func HandleMe(db *gosqlitex.DbClient, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		logger.Log.Error("Method not allowed", "method", r.Method)
		api.Fail(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}
	_, token, ok := strings.Cut(r.Header.Get("authorization"), " ")
	if !ok {
		logger.Log.Error("Invalid authorization header")
		api.Fail(w, http.StatusUnauthorized, "Unauthorized", nil)
	}
	ses, err := session.GetSession(db, []byte(token))
	if err != nil {
		logger.Log.Error("Session not found", "id", token)
		api.Fail(w, http.StatusNotFound, "User not found", err)
	}
	u, err := user.GetUser(db, int(ses.UserId))
	if err != nil {
		logger.Log.Error("Session not found", "id", token)
		api.Fail(w, http.StatusNotFound, "User not found", err)
	}
	api.Success(w, http.StatusOK, "User found", u)
}
