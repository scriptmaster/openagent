package auth

import (
	"database/sql"
	"errors"
	"net/http"
	"time"
)

// UserService handles user-related operations
type UserService struct {
	db *sql.DB
}

// NewUserService creates a new user service
func NewUserService(db *sql.DB) *UserService {
	return &UserService{db: db}
}

// GetUserByEmail retrieves a user by their email
func (s *UserService) GetUserByEmail(email string) (*User, error) {
	var user User
	err := s.db.QueryRow(`
		SELECT id, email, is_admin, created_at, last_logged_in 
		FROM ai.users 
		WHERE email = $1
	`, email).Scan(&user.ID, &user.Email, &user.IsAdmin, &user.CreatedAt, &user.LastLoggedIn)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// CreateUser creates a new user
// If no users exist yet, the first user will be an admin
func (s *UserService) CreateUser(email string) (*User, error) {
	// Check if this should be an admin (first user)
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM ai.users").Scan(&count)
	if err != nil {
		return nil, err
	}

	isFirstUser := count == 0

	// Create the user
	user := &User{
		Email:        email,
		IsAdmin:      isFirstUser, // First user is admin
		CreatedAt:    time.Now(),
		LastLoggedIn: time.Now(),
	}

	err = s.db.QueryRow(`
		INSERT INTO ai.users (email, is_admin, created_at, last_logged_in)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, user.Email, user.IsAdmin, user.CreatedAt, user.LastLoggedIn).Scan(&user.ID)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUserLastLogin updates the last login time for a user
func (s *UserService) UpdateUserLastLogin(userID int) error {
	_, err := s.db.Exec(`
		UPDATE ai.users 
		SET last_logged_in = $1 
		WHERE id = $2
	`, time.Now(), userID)

	return err
}

// GetUserFromSession retrieves the user from the session cookie
func (s *UserService) GetUserFromSession(r *http.Request) (*User, error) {
	// Get session cookie with versioned name
	cookieName := GetSessionCookieName()
	cookie, err := r.Cookie(cookieName)

	// Fall back to the generic "session" cookie for backward compatibility
	if err != nil {
		cookie, err = r.Cookie("session")
		if err != nil {
			return nil, errors.New("no session cookie found")
		}
	}

	// Validate session
	session, valid := GetSession(cookie.Value)
	if !valid {
		return nil, errors.New("invalid or expired session")
	}

	return session.User, nil
}

// GetSession retrieves a session from a request
func (s *UserService) GetSession(r *http.Request) (*Session, error) {
	// Get session cookie with versioned name
	cookieName := GetSessionCookieName()
	cookie, err := r.Cookie(cookieName)

	// Fall back to the generic "session" cookie for backward compatibility
	if err != nil {
		cookie, err = r.Cookie("session")
		if err != nil {
			return nil, errors.New("no session cookie found")
		}
	}

	// Validate session
	session, valid := GetSession(cookie.Value)
	if !valid {
		return nil, errors.New("invalid or expired session")
	}

	return &session, nil
}

// CheckIfAdminExists checks if any admin user exists in the system
func (s *UserService) CheckIfAdminExists() (bool, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) 
		FROM ai.users 
		WHERE is_admin = true
	`).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
