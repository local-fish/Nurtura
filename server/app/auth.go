package app

import (
	"context"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost = 8 // Note: This is set as 8 but this is UNSAFE. This should be increased to at least 10, better 12 if this ever goes to prod.
	minPasswordLen = 10
	maxPasswordLen = 72 // bcrypt truncates beyond 72 bytes
	tokenTTL = 24 * time.Hour
)

type ctxKey int

const userIdKey ctxKey = 0

func userId(r *http.Request) string {
	uid, _ := r.Context().Value(userIdKey).(string)
	return uid
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func normalizeEmail(s string) (string, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	addr, err := mail.ParseAddress(s)
	if err != nil || addr.Address != s {
		return "", false
	}
	return s, true
}

func (a *App) register(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if err := decode(w, r, &c, false); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	email, ok := normalizeEmail(c.Email)
	if !ok {
		writeErr(w, http.StatusBadRequest, "invalid email")
		return
	}
	if len(c.Password) < minPasswordLen || len(c.Password) > maxPasswordLen {
		writeErr(w, http.StatusBadRequest, "password must be 10-72 characters")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(c.Password), bcryptCost)
	if err != nil {
		serverErr(w, r, err)
		return
	}
	var id string
	err = a.db.QueryRow(r.Context(),
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`,
		email, string(hash)).Scan(&id)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		writeErr(w, http.StatusConflict, "email already registered")
		return
	}
	if err != nil {
		serverErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id, "email": email})
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if err := decode(w, r, &c, false); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	email, _ := normalizeEmail(c.Email)

	var id, hash string
	err := a.db.QueryRow(r.Context(),
		`SELECT id, password_hash FROM users WHERE email = $1`, email).Scan(&id, &hash)
	if errors.Is(err, pgx.ErrNoRows) {
		// Burn the same bcrypt time so unknown emails aren't distinguishable by timing.
		_ = bcrypt.CompareHashAndPassword(a.dummyHash, []byte(c.Password))
		writeErr(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		serverErr(w, r, err)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(c.Password)) != nil {
		writeErr(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	now := time.Now()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   id,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
	}).SignedString(a.jwtSecret)
	if err != nil {
		serverErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"token": token, "token_type": "Bearer", "expires_in": int(tokenTTL.Seconds()),
	})
}

func (a *App) requireAuth(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || raw == "" {
			writeErr(w, http.StatusUnauthorized, "missing bearer token")
			return
		}
		claims := &jwt.RegisteredClaims{}
		tok, err := jwt.ParseWithClaims(raw, claims,
			func(*jwt.Token) (any, error) { return a.jwtSecret, nil },
			jwt.WithValidMethods([]string{"HS256"}),
			jwt.WithExpirationRequired(),
		)
		if err != nil || !tok.Valid || !uuidRe.MatchString(claims.Subject) {
			writeErr(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		ctx := context.WithValue(r.Context(), userIdKey, claims.Subject)
		next(w, r.WithContext(ctx))
	})
}
