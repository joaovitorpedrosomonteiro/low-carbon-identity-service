package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/joaovitorpedrosomonteiro/low-carbon-identity-service/internal/application/command"
)

type JWTTokenService struct {
	accessSecret  []byte
	refreshTokens sync.Map
}

func NewJWTTokenService() *JWTTokenService {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-in-production"
	}
	return &JWTTokenService{accessSecret: []byte(secret)}
}

type accessClaims struct {
	UserID             string  `json:"sub"`
	Role               string  `json:"role"`
	CompanyID          *string `json:"company_id,omitempty"`
	BranchID           *string `json:"branch_id,omitempty"`
	MustChangePassword bool    `json:"must_change_password"`
	jwt.RegisteredClaims
}

func (s *JWTTokenService) GenerateTokenPair(userID, role string, companyID, branchID *string, mustChangePassword bool) (command.TokenPair, error) {
	accessExp := time.Now().Add(15 * time.Minute)
	claims := accessClaims{
		UserID:             userID,
		Role:               role,
		CompanyID:          companyID,
		BranchID:           branchID,
		MustChangePassword: mustChangePassword,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "low-carbon-identity-service",
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessStr, err := accessToken.SignedString(s.accessSecret)
	if err != nil {
		return command.TokenPair{}, fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshBytes := make([]byte, 32)
	rand.Read(refreshBytes)
	refreshStr := hex.EncodeToString(refreshBytes)

	s.refreshTokens.Store(refreshStr, userID)

	return command.TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
		ExpiresIn:    900,
	}, nil
}

func (s *JWTTokenService) ValidateAccessToken(tokenStr string) (map[string]any, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &accessClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.accessSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*accessClaims); ok && token.Valid {
		result := map[string]any{
			"sub":                  claims.UserID,
			"role":                 claims.Role,
			"must_change_password": claims.MustChangePassword,
		}
		if claims.CompanyID != nil {
			result["company_id"] = *claims.CompanyID
		}
		if claims.BranchID != nil {
			result["branch_id"] = *claims.BranchID
		}
		return result, nil
	}

	return nil, errors.New("invalid token")
}

func (s *JWTTokenService) RefreshAccessToken(refreshToken string) (command.TokenPair, error) {
	val, ok := s.refreshTokens.LoadAndDelete(refreshToken)
	if !ok {
		return command.TokenPair{}, errors.New("invalid or expired refresh token")
	}

	userID := val.(string)
	return s.GenerateTokenPair(userID, "unknown", nil, nil, false)
}

func (s *JWTTokenService) RevokeAllUserTokens(ctx context.Context, userID string) error {
	s.refreshTokens.Range(func(key, value any) bool {
		if uid, ok := value.(string); ok && uid == userID {
			s.refreshTokens.Delete(key)
		}
		return true
	})
	return nil
}

func writeJSON(w interface{ WriteHeader(int); Write([]byte) (int, error) }, status int, v any) {
	data, _ := json.Marshal(v)
	if headerWriter, ok := w.(interface{ Header() map[string][]string }); ok {
		headerWriter.Header()["Content-Type"] = []string{"application/json"}
	}
	w.WriteHeader(status)
	w.Write(data)
}
