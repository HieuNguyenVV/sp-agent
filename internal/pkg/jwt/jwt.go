package jwt

import (
	"errors"
	"fmt"
	"time"

	appErrors "github.com/HieuNguyenVV/sp-agent/internal/errors"
	"github.com/HieuNguyenVV/sp-agent/internal/pkg/config"
	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type TokenClaims struct {
	Claims Claims `json:"claims"`
	jwt.RegisteredClaims
}

type jwtService struct {
	config *config.Config
}

type JWTService interface {
	GenerateAccessToken(claims *Claims) (string, error)
	ValidateAccessToken(token string) (*TokenClaims, error)
	GenerateRefreshToken(userID string) (string, error)
	ValidateRefreshToken(token string) (string, error)
	GetAccessTokenTTL() time.Duration
	GetRefreshTokenTTL() time.Duration
}

func NewJWTService(config *config.Config) JWTService {
	return &jwtService{config: config}
}

func (s *jwtService) GenerateAccessToken(claims *Claims) (string, error) {
	if claims == nil {
		return "", appErrors.ErrNilClaims
	}
	now := time.Now()
	jwtClaims := TokenClaims{
		Claims: *claims,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "sp-agent",
			Subject:   claims.ID,
			Audience:  jwt.ClaimStrings{"api"},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.config.JWT.AccessTokenTTL) * time.Second)),
		},
	}
	return s.sign(jwtClaims, s.config.JWT.AccessTokenSecret)
}

func (s *jwtService) sign(claims jwt.Claims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func (s *jwtService) ValidateAccessToken(token string) (*TokenClaims, error) {
	claims := &TokenClaims{}
	if err := s.parse(token, s.config.JWT.AccessTokenSecret, claims); err != nil {
		return nil, err
	}
	return claims, nil
}

func (s *jwtService) parse(tokenString string, secret string, claims jwt.Claims) error {
	key := []byte(secret)
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return appErrors.ErrTokenExpired.Wrap(err)
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return appErrors.ErrTokenNotValidYet.Wrap(err)
		}
		return err
	}
	if !token.Valid {
		return appErrors.ErrInvalidToken.Wrap(err)
	}
	return nil
}

func (s *jwtService) GenerateRefreshToken(userID string) (string, error) {
	if userID == "" {
		return "", appErrors.ErrInvalidArgument.Wrap(errors.New("user ID is required"))
	}
	now := time.Now()
	jwtClaims := TokenClaims{
		Claims: Claims{
			ID: userID,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "sp-agent",
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.config.JWT.RefreshTokenTTL) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	return s.sign(jwtClaims, s.config.JWT.AccessTokenSecret)
}

func (s *jwtService) ValidateRefreshToken(token string) (string, error) {
	claims := &TokenClaims{}
	if err := s.parse(token, s.config.JWT.AccessTokenSecret, claims); err != nil {
		return "", err
	}
	if claims.Claims.ID == "" {
		return "", appErrors.ErrInvalidToken
	}
	return claims.Claims.ID, nil
}

func (s *jwtService) GetAccessTokenTTL() time.Duration {
	return time.Duration(s.config.JWT.AccessTokenTTL) * time.Second
}

func (s *jwtService) GetRefreshTokenTTL() time.Duration {
	return time.Duration(s.config.JWT.RefreshTokenTTL) * time.Second
}
