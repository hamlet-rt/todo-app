package service

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"math/rand"
	"time"
	"todo"
	"todo/pkg/repository"

	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
)

const (
	salt       = "asda32fsafsaf"
	signingKey = "fwsf3fsfafddsf"
	tokenTTL   = 12 * time.Hour
)

type tokenClaims struct {
	jwt.StandardClaims
	UserId int `json:"user_id"`
}

type AuthService struct {
	repo repository.Authorization
	tokenRepo repository.Token
}

func NewAuthService(
	repo repository.Authorization,
	tokenRepo repository.Token,
	) *AuthService {
	return &AuthService{
		repo: repo,
		tokenRepo: tokenRepo,
	}
}

func (s *AuthService) CreateUser(user todo.User) (int, error) {
	user.Password = generatePasswordHash(user.Password)

	return s.repo.CreateUser(user)
}

func (s *AuthService) GetUser(username, password string) (todo.User, error) {
    user, err := s.repo.GetUser(username, generatePasswordHash(password))
    if err != nil {
        return todo.User{}, fmt.Errorf("user not found: %w", err)
    }

	return user, nil
}

func (s *AuthService) GenerateTokens(userId int) (string, string, error) {

	if err := s.tokenRepo.DeleteByUserId(userId); err != nil {
		return "", "", fmt.Errorf("failed to delete old refresh tokens: %w", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		userId,
	})

	accessToken, err := token.SignedString([]byte(signingKey))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshToken, err := newRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	if err := s.tokenRepo.Create(userId, todo.RefreshToken{
		UserID: int64(userId),
		Token: refreshToken,
		ExpiresDate: time.Now().Add(time.Hour * 24 * 30),
	}); err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return accessToken, refreshToken, err
}

func (s *AuthService) ParseToken(accessToken string) (int, error) {
	token, err := jwt.ParseWithClaims(accessToken, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}

		return []byte(signingKey), nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return 0, errors.New("token claims are not of type *tokenClaims")
	}

	return claims.UserId, nil
}

func generatePasswordHash(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))

	return fmt.Sprintf("%x", hash.Sum([]byte(salt)))
}

func newRefreshToken() (string, error) {
	tokenBytes := make([]byte, 32)

	randomSource := rand.NewSource(time.Now().Unix())
	randomGenerator := rand.New(randomSource)

	if _, err := randomGenerator.Read(tokenBytes); err != nil {
		return "", nil
	}

	return fmt.Sprintf("%x", tokenBytes), nil
}

func (s *AuthService) RefreshTokens(refreshToken todo.RefreshToken) (string, string, error) {
	refreshToken, err := s.tokenRepo.Get(refreshToken.Token)
	if err != nil {
		return "", "", err
	}

	if refreshToken.ExpiresDate.Unix() < time.Now().Unix() {
		return "", "", todo.ErrRefreshTokenExpired
	}

	return s.GenerateTokens(int(refreshToken.UserID))
}

func (s *AuthService) DeleteExpiredRefreshTokens() error {
	currentDate := time.Now()
	err := s.tokenRepo.DeleteExpired(currentDate)
	if err != nil {
		logrus.Errorf("Error deleting expired tokens: %v", err)
		return err
	}

	logrus.Infof("Expired tokens successfully deleted: %v", currentDate)
	return nil
}