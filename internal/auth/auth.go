// Package auth provides authentication and authorization functionalities.
// It includes handlers for user registration and login with JWT token generation.
package auth

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication-related operations such as registering and logging in users.
type AuthService struct {
	db     *sql.DB
	secret string
}

// NewAuthService creates a new AuthService instance with a database connection and JWT secret.
func NewAuthService(db *sql.DB, secret string) *AuthService {
	return &AuthService{db: db, secret: secret}
}

// RegisterHandler handles POST /v1/register requests to register a new user.
func (a *AuthService) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	type req struct{ Email, Password string }
	var rr req
	_ = json.NewDecoder(r.Body).Decode(&rr)
	if rr.Email == "" || rr.Password == "" {
		http.Error(w, "missing", http.StatusBadRequest)
		return
	}
	ph, _ := bcrypt.GenerateFromPassword([]byte(rr.Password), bcrypt.DefaultCost)
	_, err := a.db.Exec("INSERT INTO users(email,password_hash) VALUES($1,$2)", rr.Email, string(ph))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// LoginHandler handles POST /v1/login requests and returns a JWT token on successful authentication.
func (a *AuthService) LoginHandler(w http.ResponseWriter, r *http.Request) {
	type req struct{ Email, Password string }
	var rr req
	_ = json.NewDecoder(r.Body).Decode(&rr)
	if rr.Email == "" || rr.Password == "" {
		http.Error(w, "missing", http.StatusBadRequest)
		return
	}
	var id int
	var hash string
	row := a.db.QueryRow("SELECT id,password_hash FROM users WHERE email=$1", rr.Email)
	if err := row.Scan(&id, &hash); err != nil {
		http.Error(w, "invalid", http.StatusUnauthorized)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(rr.Password)) != nil {
		http.Error(w, "invalid", http.StatusUnauthorized)
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": id, "exp": time.Now().Add(24 * time.Hour).Unix()})
	s, _ := token.SignedString([]byte(a.secret))
	_ = json.NewEncoder(w).Encode(map[string]string{"token": s})
}
