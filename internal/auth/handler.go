package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/irabeny89/gbege/internal/api"
	"github.com/irabeny89/gbege/internal/security"
	"github.com/irabeny89/gbege/internal/session"
	"github.com/irabeny89/gbege/internal/user"
	"github.com/irabeny89/gosqlitex"
)

type LoginRequest struct {
	Alias    string `json:"alias"`
	Password string `json:"password"`
}

func (l *LoginRequest) Validate() error {
	if l.Alias == "" {
		return errors.New("alias is required")
	}
	if l.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

// HandleLogin handles login requests
func HandleLogin(db *gosqlitex.DbClient, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.Fail(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Fail(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := req.Validate(); err != nil {
		api.Fail(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	user, err := user.GetUserByAlias(db, req.Alias)
	if err != nil {
		api.Fail(w, http.StatusUnauthorized, "Invalid credentials", nil)
		return
	}

	ok, err := security.VerifyPassword(req.Password, user.Password)
	if err != nil || !ok {
		api.Fail(w, http.StatusUnauthorized, "Invalid credentials", nil)
		return
	}

	sess, err := session.SaveSession(db, int(user.Id))
	if err != nil {
		api.Fail(w, http.StatusInternalServerError, "Failed to create session", err)
		return
	}

	api.Success(w, http.StatusOK, "Login successful", map[string]any{
		"token": base64.StdEncoding.EncodeToString(sess.Id),
		"user":  user,
	})
}

// HandleSignUp handles signup requests
func HandleSignUp(db *gosqlitex.DbClient, w http.ResponseWriter, r *http.Request) {

}

// HandleLogout handles logout requests
func HandleLogout(db *gosqlitex.DbClient, w http.ResponseWriter, r *http.Request) {

}

// HandleMe handles requests for the current user
func HandleMe(db *gosqlitex.DbClient, w http.ResponseWriter, r *http.Request) {

}
