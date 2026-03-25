package service

import (
	"context"
	"fmt"
	"net/http"

	"bidxi/internal/config"
	"bidxi/internal/dto"
	"bidxi/internal/models"
	"bidxi/internal/repository"
	"bidxi/internal/utils"
)

// AuthService defines the interface for authentication business logic.
type AuthService interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, int, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, int, error)
	GetCurrentUser(ctx context.Context, userID string) (*dto.UserDTO, int, error)
}

// authService is the concrete implementation of AuthService.
type authService struct {
	userRepo repository.UserRepository
}

// NewAuthService creates a new AuthService with the given user repository.
func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{userRepo: userRepo}
}

// Register handles user registration: validates uniqueness, hashes password, creates user, returns JWT.
func (s *authService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, int, error) {
	// Check if email already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("internal server error")
	}
	if existingUser != nil {
		return nil, http.StatusConflict, fmt.Errorf("email already registered")
	}

	// Check if username already exists
	existingUser, err = s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("internal server error")
	}
	if existingUser != nil {
		return nil, http.StatusConflict, fmt.Errorf("username already taken")
	}

	// Hash the password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("internal server error")
	}

	// Create user model
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	// Save to database
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to create user: %v", err)
	}

	// Generate JWT token
	token, err := utils.GenerateToken(
		user.ID.Hex(),
		user.Email,
		user.Username,
		config.AppConfig.JWTSecret,
		config.AppConfig.JWTExpiry,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to generate token")
	}

	return &dto.AuthResponse{
		Token: token,
		User:  toUserDTO(user),
	}, http.StatusCreated, nil
}

// Login authenticates a user by email/password and returns a JWT token.
func (s *authService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, int, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("internal server error")
	}
	if user == nil {
		return nil, http.StatusUnauthorized, fmt.Errorf("invalid email or password")
	}

	// Compare password
	if err := utils.ComparePassword(user.PasswordHash, req.Password); err != nil {
		return nil, http.StatusUnauthorized, fmt.Errorf("invalid email or password")
	}

	// Generate JWT token
	token, err := utils.GenerateToken(
		user.ID.Hex(),
		user.Email,
		user.Username,
		config.AppConfig.JWTSecret,
		config.AppConfig.JWTExpiry,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to generate token")
	}

	return &dto.AuthResponse{
		Token: token,
		User:  toUserDTO(user),
	}, http.StatusOK, nil
}

// GetCurrentUser retrieves the currently authenticated user's data.
func (s *authService) GetCurrentUser(ctx context.Context, userID string) (*dto.UserDTO, int, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("internal server error")
	}
	if user == nil {
		return nil, http.StatusNotFound, fmt.Errorf("user not found")
	}

	userDTO := toUserDTO(user)
	return &userDTO, http.StatusOK, nil
}

// toUserDTO converts a User model to a safe UserDTO (no password hash).
func toUserDTO(user *models.User) dto.UserDTO {
	return dto.UserDTO{
		ID:        user.ID.Hex(),
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
