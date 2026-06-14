package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v4"
)

type ctxKey int

const (
	userIDKey ctxKey = iota
)

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}

func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

var ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
var ErrTokenIsNotValid = errors.New("token is not valid")

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

func GenerateUserID() (userID string, err error) {
	b := make([]byte, 24)
	_, err = rand.Read(b)
	if err != nil {
		return "", err
	}
	userID = base64.URLEncoding.EncodeToString(b)
	return
}

func GenerateJWTAuthToken(userID string, secretKey []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: userID,
	})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ValidateAndParseJWTAuthToken(tokenString string, secretKey []byte) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("%w: %v", ErrUnexpectedSigningMethod, t.Header["alg"])
			}
			return []byte(secretKey), nil
		})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", ErrTokenIsNotValid
	}

	return claims.UserID, nil
}
