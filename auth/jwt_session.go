package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/scriptmaster/openagent/common"
)

var (
	sessions     = make(map[string]Session) // Uses Session from types.go
	sessionMutex = &sync.RWMutex{}
)

var jwtSecret []byte

// JWT custom claims structure
type UserClaims struct {
	UserID  int    `json:"user_id"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

// InitializeJWTSecret loads the secret key
func InitializeJWTSecret(host string) error {
	if jwtSecret != nil {
		return nil
	}

	// Initialize JWT secret if not already done (important for middleware)
	secretKey := common.GetEnv("JWT_SECRET_KEY")
	if secretKey == "" {
		log.Println("** WARNING: JWT_SECRET_KEY environment variable not set. Generating a default key.")
		// Generate a default key using SESSION_SALT, host, APP_NAME, and APP_VERSION
		sessionSalt := common.GetEnv("SESSION_SALT")
		appName := common.GetEnv("APP_NAME")
		appVersion := common.GetEnv("APP_VERSION")

		if sessionSalt == "" {
			return errors.New("SESSION_SALT environment variable is required")
		}

		// Create a composite key from multiple sources
		compositeKey := fmt.Sprintf("%s:%s:%s:%s", sessionSalt, host, appName, appVersion)
		hash := sha256.Sum256([]byte(compositeKey))
		secretKey = hex.EncodeToString(hash[:])

		log.Printf("** Generated default JWT secret from SESSION_SALT, host, APP_NAME, and APP_VERSION")
	}

	jwtSecret = []byte(secretKey)
	return nil
}

// GenerateSessionToken generates a secure session token
func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateSession creates a new session for a user
func CreateSession(user *User, host string) (string, error) {
	// Initialize JWT secret if not already done
	if err := InitializeJWTSecret(host); err != nil {
		return "", err
	}

	// Create JWT claims
	claims := &UserClaims{
		UserID:  user.ID,
		Email:   user.Email,
		IsAdmin: user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(168 * time.Hour)), // 7 days
		},
	}

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	// Store session
	sessionMutex.Lock()
	sessions[jwtString] = Session{
		Token:     jwtString,
		User:      user,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(168 * time.Hour),
	}
	sessionMutex.Unlock()

	return jwtString, nil
}

// ValidateJWT validates a JWT token and returns the claims
func ValidateJWT(tokenString string) (*UserClaims, error) {
	if jwtSecret == nil {
		return nil, errors.New("JWT secret not initialized")
	}

	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GetSession retrieves a session by token
func GetSession(token string) (Session, bool) {
	sessionMutex.RLock()
	defer sessionMutex.RUnlock()
	session, exists := sessions[token]
	return session, exists
}

// ClearSession removes a session
func ClearSession(token string) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	delete(sessions, token)
}

// SetUserContext adds user information to the request context
func SetUserContext(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userCtxKey{}, user)
}

// GetUserFromContext retrieves user information from the request context
func GetUserFromContext(ctx context.Context) *User {
	if user, ok := ctx.Value(userCtxKey{}).(*User); ok {
		return user
	}
	return nil
}

// CleanupSessions removes expired sessions
func CleanupSessions() {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	now := time.Now()
	for token, session := range sessions {
		if session.ExpiresAt.Before(now) {
			delete(sessions, token)
		}
	}
}
