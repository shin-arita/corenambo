package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type App struct {
	DB                  *sql.DB
	CORSAllowOrigin     string
	JWTAccessSecret     []byte
	JWTRefreshSecret    []byte
	JWTAccessExpiresMin int
	JWTRefreshExpiresHr int
	CookieSecure        bool
}

type contextKey string

const userIDContextKey contextKey = "userID"

type Claims struct {
	UserID int64 `json:"uid"`
	jwt.RegisteredClaims
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Product struct {
	ID    int64   `json:"id"`
	SKU   string  `json:"sku"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func main() {
	port := getenv("PORT", "8080")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getenv("DB_HOST", "db"),
		getenv("DB_PORT", "5432"),
		getenv("DB_USER", "app_user"),
		getenv("DB_PASSWORD", "password"),
		getenv("DB_NAME", "app_db"),
		getenv("DB_SSLMODE", "disable"),
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	app := &App{
		DB:                  db,
		CORSAllowOrigin:     getenv("CORS_ALLOW_ORIGIN", "http://localhost:5173"),
		JWTAccessSecret:     []byte(getenv("JWT_ACCESS_SECRET", "replace-me-access-secret")),
		JWTRefreshSecret:    []byte(getenv("JWT_REFRESH_SECRET", "replace-me-refresh-secret")),
		JWTAccessExpiresMin: getenvInt("JWT_ACCESS_EXPIRES_MINUTES", 15),
		JWTRefreshExpiresHr: getenvInt("JWT_REFRESH_EXPIRES_HOURS", 168),
		CookieSecure:        getenvBool("COOKIE_SECURE", false),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", app.health)
	mux.HandleFunc("/api/v1/auth/login", app.login)
	mux.HandleFunc("/api/v1/auth/refresh", app.refresh)
	mux.HandleFunc("/api/v1/auth/logout", app.logout)
	mux.Handle("/api/v1/me", app.authMiddleware(http.HandlerFunc(app.me)))
	mux.HandleFunc("/api/v1/products/search", app.searchProducts)

	handler := app.corsMiddleware(mux)

	log.Printf("server started on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

func (a *App) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && origin == a.CORSAllowOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *App) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]any{
				"error": "authorization header is required",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			writeJSON(w, http.StatusUnauthorized, map[string]any{
				"error": "invalid authorization header",
			})
			return
		}

		claims, err := a.parseAccessToken(tokenString)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]any{
				"error": "invalid access token",
			})
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *App) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
	})
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}

	userID, email, err := a.authenticateUser(r.Context(), strings.TrimSpace(req.Email), req.Password)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid email or password"})
		return
	}

	accessToken, err := a.createAccessToken(userID, email)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to create access token"})
		return
	}

	refreshToken, refreshTokenHash, refreshExpiresAt, err := a.createRefreshToken(userID, email)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to create refresh token"})
		return
	}

	if err := a.saveRefreshToken(r.Context(), userID, refreshTokenHash, refreshExpiresAt); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to save refresh token"})
		return
	}

	a.setRefreshCookie(w, refreshToken, refreshExpiresAt)

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"accessToken": accessToken,
		},
	})
}

func (a *App) refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "refresh token cookie is required"})
		return
	}

	claims, err := a.parseRefreshToken(cookie.Value)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid refresh token"})
		return
	}

	refreshHash := hashToken(cookie.Value)
	if err := a.validateRefreshToken(r.Context(), claims.UserID, refreshHash); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "refresh token is revoked or expired"})
		return
	}

	if err := a.revokeRefreshToken(r.Context(), claims.UserID, refreshHash); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to rotate refresh token"})
		return
	}

	accessToken, err := a.createAccessToken(claims.UserID, claims.Subject)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to create access token"})
		return
	}

	newRefreshToken, newRefreshHash, newRefreshExpiresAt, err := a.createRefreshToken(claims.UserID, claims.Subject)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to create refresh token"})
		return
	}

	if err := a.saveRefreshToken(r.Context(), claims.UserID, newRefreshHash, newRefreshExpiresAt); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to save refresh token"})
		return
	}

	a.setRefreshCookie(w, newRefreshToken, newRefreshExpiresAt)

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"accessToken": accessToken,
		},
	})
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}

	cookie, err := r.Cookie("refresh_token")
	if err == nil {
		claims, parseErr := a.parseRefreshToken(cookie.Value)
		if parseErr == nil {
			_ = a.revokeRefreshToken(r.Context(), claims.UserID, hashToken(cookie.Value))
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   a.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"message": "logged out",
		},
	})
}

func (a *App) me(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDContextKey).(int64)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
		return
	}

	var email, name string
	err := a.DB.QueryRowContext(
		r.Context(),
		`SELECT email, name FROM users WHERE id = $1`,
		userID,
	).Scan(&email, &name)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to fetch user"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"id":    userID,
			"email": email,
			"name":  name,
		},
	})
}

func (a *App) searchProducts(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "q is required",
		})
		return
	}

	rows, err := a.DB.QueryContext(r.Context(), `
		SELECT id, sku, name, price
		  FROM products
		 WHERE is_active = true
		   AND search_text &@~ $1
		 ORDER BY id DESC
		 LIMIT 20
	`, q)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": err.Error(),
		})
		return
	}
	defer rows.Close()

	products := []Product{}

	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.SKU, &p.Name, &p.Price); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{
				"error": err.Error(),
			})
			return
		}
		products = append(products, p)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": products,
	})
}

func (a *App) authenticateUser(ctx context.Context, email, password string) (int64, string, error) {
	var userID int64
	var dbEmail string

	err := a.DB.QueryRowContext(ctx, `
		SELECT id, email
		  FROM users
		 WHERE email = $1
		   AND password_hash = crypt($2, password_hash)
	`, email, password).Scan(&userID, &dbEmail)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, "", err
		}
		return 0, "", err
	}

	return userID, dbEmail, nil
}

func (a *App) createAccessToken(userID int64, email string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(a.JWTAccessExpiresMin) * time.Minute)

	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   email,
			Issuer:    "corenambo-api",
			Audience:  []string{"corenambo-frontend"},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.JWTAccessSecret)
}

func (a *App) createRefreshToken(userID int64, email string) (string, string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(a.JWTRefreshExpiresHr) * time.Hour)

	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   email,
			Issuer:    "corenambo-api",
			Audience:  []string{"corenambo-frontend"},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(a.JWTRefreshSecret)
	if err != nil {
		return "", "", time.Time{}, err
	}

	return tokenString, hashToken(tokenString), expiresAt, nil
}

func (a *App) parseAccessToken(tokenString string) (*Claims, error) {
	return a.parseToken(tokenString, a.JWTAccessSecret)
}

func (a *App) parseRefreshToken(tokenString string) (*Claims, error) {
	return a.parseToken(tokenString, a.JWTRefreshSecret)
}

func (a *App) parseToken(tokenString string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func (a *App) saveRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	_, err := a.DB.ExecContext(ctx, `
		INSERT INTO user_refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, tokenHash, expiresAt)
	return err
}

func (a *App) validateRefreshToken(ctx context.Context, userID int64, tokenHash string) error {
	var exists bool
	err := a.DB.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			  FROM user_refresh_tokens
			 WHERE user_id = $1
			   AND token_hash = $2
			   AND revoked_at IS NULL
			   AND expires_at > now()
		)
	`, userID, tokenHash).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("refresh token not found")
	}
	return nil
}

func (a *App) revokeRefreshToken(ctx context.Context, userID int64, tokenHash string) error {
	_, err := a.DB.ExecContext(ctx, `
		UPDATE user_refresh_tokens
		   SET revoked_at = now()
		 WHERE user_id = $1
		   AND token_hash = $2
		   AND revoked_at IS NULL
	`, userID, tokenHash)
	return err
}

func (a *App) setRefreshCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   a.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
	})
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func getenv(key, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}

func getenvInt(key string, defaultValue int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return defaultValue
	}
	return i
}

func getenvBool(key string, defaultValue bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	switch v {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultValue
	}
}
