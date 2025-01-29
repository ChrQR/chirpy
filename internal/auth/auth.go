package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return err
	}
	return nil
}

func MakeJWT(userId uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
  if expiresIn <= 0 {
    return "", errors.New("Token expiration must be positive.")
  }
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userId.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(tokenSecret), nil
		})
	if err != nil {
		return uuid.Nil, err
	}

  claims, ok := token.Claims.(*jwt.RegisteredClaims)
  if !ok || !token.Valid {
    return uuid.Nil, fmt.Errorf("invalid token claims")
  }

  userId, err := uuid.Parse(claims.Subject)
  if err != nil {
    return uuid.Nil, fmt.Errorf("invalid user ID in token")
  }
  return userId, nil
}

func GetBearerToken(headers http.Header) (string, error) {
  authHeader := headers.Get("Authorization")
  if authHeader == "" {
    return "", errors.New("No authorization header in request")
  }
  if !strings.HasPrefix(authHeader, "Bearer"){
    return "", errors.New("No token found in headers")
  }
  tokenString := strings.Split(authHeader, " ")[1]
  if tokenString == "" {
    return "", errors.New("Token is empty")
  }
  return tokenString, nil
}

func MakeRefreshToken() (string, error) {
  c := 32
  b := make([]byte, c)
  _, err := rand.Read(b)
  if err != nil {
    return "", err
  }
  token := hex.EncodeToString(b)
  return token, nil
}
