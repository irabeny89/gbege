package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/irabeny89/gbege/internal/api"
	"github.com/irabeny89/gbege/internal/logger"
	"github.com/irabeny89/gbege/internal/middleware"
	"github.com/irabeny89/gbege/internal/security"
	"github.com/irabeny89/gbege/internal/session"
	"github.com/irabeny89/gbege/internal/user"
	"github.com/irabeny89/gosqlitex"
)

type Handler struct {
	db *gosqlitex.DbClient
}

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
func (s *SignUpRequest) validate() error {
	if s.Username == "" {
		return errors.New("username is required")
	}
	if s.Password == "" {
		return errors.New("password is required")
	}
	if len(s.Username) < 3 {
		return errors.New("username must be at least 3 characters long")
	}
	if len(s.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
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
func (a *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	reqId := r.Context().Value(middleware.REQUEST_ID_KEY).(string)
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Error("Failed to decode request body", "req_id", reqId, "err", err)
		api.Fail(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := req.validate(); err != nil {
		logger.Log.Error("Failed to validate request body", "req_id", reqId, "err", err)
		api.Fail(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	u, err := user.GetUserByUsername(a.db, req.Username)
	if err != nil {
		logger.Log.Error("Something went wrong while fetching the user", "req_id", reqId, "username", req.Username, "err", err)
		api.Fail(w, http.StatusInternalServerError, "Something went wrong while fetching the user", err)
		return
	}

	if u == nil {
		logger.Log.Error("Invalid credentials", "req_id", reqId, "username", req.Username)
		api.Fail(w, http.StatusUnauthorized, "Invalid credentials", nil)
		return
	}

	ok, err := security.VerifyPassword(req.Password, u.Password)
	if err != nil || !ok {
		logger.Log.Error("Invalid password", "req_id", reqId, "username", req.Username, "err", err)
		api.Fail(w, http.StatusUnauthorized, "Invalid credentials", err)
		return
	}

	// clear old session if there is one
	ses, err := session.GetSessionByUserId(a.db, int(u.Id))
	if err != nil {
		logger.Log.Error("Something went wrong while fetching the session", "req_id", reqId, "user_id", u.Id, "err", err)
		api.Fail(w, http.StatusInternalServerError, "Something went wrong while fetching the session", err)
		return
	}
	if ses != nil {
		if err := session.DeleteSession(a.db, ses.Id); err != nil {
			logger.Log.Error("Something went wrong while deleting the session", "req_id", reqId, "user_id", u.Id, "err", err)
			api.Fail(w, http.StatusInternalServerError, "Something went wrong while deleting the session", err)
			return
		}
	}

	ses, err = session.SaveSession(a.db, int(u.Id))
	if err != nil {
		logger.Log.Error("Failed to create session", "req_id", reqId, "user_id", u.Id, "err", err)
		api.Fail(w, http.StatusInternalServerError, "Failed to create session", err)
		return
	}
	logger.Log.Info("Login successful", "req_id", reqId, "user_id", u.Id)
	api.Success(w, http.StatusOK, "Login successful", LoginResponse{
		Token: base64.StdEncoding.EncodeToString(ses.Id),
		User:  u,
	})
}

// MARK: - Signup
// HandleSignUp handles signup requests
func (a *Handler) HandleSignUp(w http.ResponseWriter, r *http.Request) {
	reqId := r.Context().Value(middleware.REQUEST_ID_KEY).(string)
	var req SignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Error("Failed to decode request body", "req_id", reqId, "err", err)
		api.Fail(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := req.validate(); err != nil {
		logger.Log.Error("Failed to validate request body", "req_id", reqId, "err", err)
		api.Fail(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	u, err := user.GetUserByUsername(a.db, req.Username)
	if err != nil {
		logger.Log.Error("Something went wrong while fetching the user", "req_id", reqId, "username", req.Username, "err", err)
		api.Fail(w, http.StatusInternalServerError, "Something went wrong while fetching the user", err)
		return
	}

	if u != nil {
		logger.Log.Error("User already exists", "req_id", reqId, "username", req.Username)
		api.Fail(w, http.StatusUnauthorized, "User already exists", nil)
		return
	}

	u, err = user.SaveUser(a.db, req.Username, req.Password)
	if err != nil {
		logger.Log.Error("Failed to create user", "req_id", reqId, "username", req.Username, "err", err)
		api.Fail(w, http.StatusInternalServerError, "Failed to create user", err)
		return
	}

	ses, err := session.SaveSession(a.db, int(u.Id))
	if err != nil {
		logger.Log.Error("Failed to create session", "req_id", reqId, "user_id", u.Id, "err", err)
		api.Fail(w, http.StatusInternalServerError, "Failed to create session", err)
		return
	}
	logger.Log.Info("User created successfully", "req_id", reqId, "user_id", u.Id)
	api.Success(w, http.StatusOK, "User created successfully", map[string]any{"user_id": u.Id, "token": base64.StdEncoding.EncodeToString(ses.Id)})
}

// MARK: - Logout
// HandleLogout handles logout requests
func (a *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	reqId := r.Context().Value(middleware.REQUEST_ID_KEY).(string)

	_, token, ok := strings.Cut(r.Header.Get("authorization"), " ")
	if !ok {
		logger.Log.Error("Invalid authorization header", "req_id", reqId)
		api.Fail(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	err := session.DeleteSession(a.db, []byte(token))
	if err != nil {
		logger.Log.Error("Failed to delete session", "req_id", reqId)
		api.Fail(w, http.StatusInternalServerError, "Failed to delete session", err)
		return
	}
	api.Success(w, http.StatusOK, "Logout successful", nil)
}

// MARK: - Me
// HandleMe handles requests for the current user
func (a *Handler) HandleMe(w http.ResponseWriter, r *http.Request) {
	reqId := r.Context().Value(middleware.REQUEST_ID_KEY).(string)
	_, token, ok := strings.Cut(r.Header.Get("authorization"), " ")
	if !ok {
		logger.Log.Error("Invalid authorization header", "req_id", reqId)
		api.Fail(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	ses, err := session.GetSession(a.db, []byte(token))
	if err != nil {
		logger.Log.Error("Session not found", "req_id", reqId, "id", token)
		api.Fail(w, http.StatusNotFound, "User not found", err)
		return
	}
	u, err := user.GetUser(a.db, int(ses.UserId))
	if err != nil {
		logger.Log.Error("User not found", "req_id", reqId, "user_id", ses.UserId, "err", err)
		api.Fail(w, http.StatusNotFound, "User not found", err)
		return
	}
	api.Success(w, http.StatusOK, "User found", u)
}

// MARK: - New
// NewHandler creates a new auth handler
func NewHandler(db *gosqlitex.DbClient) *Handler {
	return &Handler{db}
}
