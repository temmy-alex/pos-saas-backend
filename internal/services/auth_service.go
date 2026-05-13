package services

import (
	"context"
	"errors"
	"time"

	"pos-saas-backend/internal/models"
	"pos-saas-backend/internal/repositories"
	"pos-saas-backend/internal/requests"
	"pos-saas-backend/internal/responses"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserRepository               *repositories.UserRepository
	JWTSecret                    string
	JWTAccessTokenExpiresMinutes int
}

type AuthClaims struct {
	UserID   int64  `json:"user_id"`
	StoreID  *int64 `json:"store_id"`
	BranchID *int64 `json:"branch_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(
	userRepository *repositories.UserRepository,
	jwtSecret string,
	jwtAccessTokenExpiresMinutes int,
) *AuthService {
	return &AuthService{
		UserRepository:               userRepository,
		JWTSecret:                    jwtSecret,
		JWTAccessTokenExpiresMinutes: jwtAccessTokenExpiresMinutes,
	}
}

func (s *AuthService) Login(ctx context.Context, request requests.LoginRequest) (*responses.LoginResponse, error) {
	user, err := s.UserRepository.FindByEmail(ctx, request.Email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("invalid email or password")
	}

	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	accessToken, expiresAt, err := s.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	if err := s.UserRepository.UpdateLastLogin(ctx, user.ID); err != nil {
		return nil, err
	}

	expiresIn := int64(time.Until(expiresAt).Seconds())

	return &responses.LoginResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		User:        s.userToAuthUserResponse(user),
	}, nil
}

func (s *AuthService) GenerateAccessToken(user *models.User) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(s.JWTAccessTokenExpiresMinutes) * time.Minute)

	claims := AuthClaims{
		UserID:   user.ID,
		StoreID:  user.StoreID,
		BranchID: user.BranchID,
		Name:     user.Name,
		Email:    user.Email,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Email,
			Issuer:    "pos-saas-backend",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(s.JWTSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return signedToken, expiresAt, nil
}

func (s *AuthService) userToAuthUserResponse(user *models.User) responses.AuthUserResponse {
	return responses.AuthUserResponse{
		ID:       user.ID,
		StoreID:  user.StoreID,
		BranchID: user.BranchID,
		Name:     user.Name,
		Email:    user.Email,
		Role:     user.Role,
	}
}
