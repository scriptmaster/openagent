package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/scriptmaster/openagent/common"
	"golang.org/x/crypto/bcrypt"
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
	user := &User{}
	// Ensure the query selects the password hash
	query := common.MustGetSQL("auth/get_user_by_email")
	err := s.db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.IsAdmin, &user.CreatedAt, &user.LastLoggedIn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error fetching user: %w", err)
	}
	return user, nil
}

// CreateUser creates a new user and hashes their password
// If no users exist yet, the first user will be an admin
func (s *UserService) CreateUser(email, password string) (*User, error) {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Check if this should be an admin (first user)
	var count int
	countQuery := common.MustGetSQL("auth/count_all")
	err = s.db.QueryRow(countQuery).Scan(&count)
	if err != nil {
		// Consider specific error handling if the table doesn't exist yet during first setup
		return nil, fmt.Errorf("failed to check user count: %w", err)
	}

	isFirstUser := count == 0

	// Create the user
	user := &User{
		Email:        email,
		IsAdmin:      isFirstUser, // First user is admin
		CreatedAt:    time.Now(),
		LastLoggedIn: time.Now(), // Set initial login time
	}

	createQuery := common.MustGetSQL("auth/create")
	err = s.db.QueryRow(createQuery, user.Email, string(hashedPassword), user.IsAdmin, user.CreatedAt, user.LastLoggedIn).Scan(&user.ID)

	if err != nil {
		// Handle potential duplicate email errors more gracefully if needed
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}

	return user, nil
}

// UpdateUserLastLogin updates the last login time for a user
func (s *UserService) UpdateUserLastLogin(userID int) error {
	updateQuery := common.MustGetSQL("auth/update_last_login")
	_, err := s.db.Exec(updateQuery, time.Now(), userID)

	return err
}

// CheckIfAdminExists checks if any admin user exists in the system
func (s *UserService) CheckIfAdminExists() (bool, error) {
	var count int
	countQuery := common.MustGetSQL("auth/count_admins")
	err := s.db.QueryRow(countQuery).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// MakeUserAdmin makes a user an admin
func (s *UserService) MakeUserAdmin(email string) error {
	user, err := s.GetUserByEmail(email)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}

	if user == nil {
		return fmt.Errorf("user not found")
	}

	makeAdminQuery := common.MustGetSQL("auth/make_admin")
	_, err = s.db.Exec(makeAdminQuery, user.ID)

	return err
}

// VerifyPassword compares a plaintext password with the stored hash for a user
func (s *UserService) VerifyPassword(email, password string) (*User, error) {
	user, err := s.GetUserByEmail(email) // Assuming GetUserByEmail fetches the hash
	if err != nil {
		return nil, err // Handles user not found
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		// Handle incorrect password specifically
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, errors.New("incorrect password")
		}
		// Handle other potential errors
		return nil, fmt.Errorf("password verification error: %w", err)
	}

	// Password is correct
	return user, nil
}
