package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"github.com/SergeyRG/shortener/internal/config"
	"github.com/SergeyRG/shortener/internal/contextutils"
	"github.com/golang-jwt/jwt/v4"
)

var ErrUnexpectedSigningMethod = fmt.Errorf("unexpected signing method")
var ErrTokenIsNotValid = fmt.Errorf("token is not valid")

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

func generateUserID() (userID string, err error) {
	b := make([]byte, 24)
	_, err = rand.Read(b)
	if err != nil {
		return "", err
	}
	userID = base64.URLEncoding.EncodeToString(b)
	return
}

func generateJWTAuthToken(userID string, secretKey []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: userID,
	})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func validateAndParseJWTAuthToken(tokenString string, secretKey []byte) (string, error) {
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

func Auth(h http.HandlerFunc, cfg config.Config) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		cookieToken, err := r.Cookie("auth_token")
		var tokenString string
		var userID string
		newTokenRequired := false

		if err == nil {
			tokenString = cookieToken.Value
			userID, err = validateAndParseJWTAuthToken(tokenString, []byte(cfg.SecretKey))
		}

		if err != nil {
			newTokenRequired = errors.Is(err, http.ErrNoCookie) ||
				errors.Is(err, ErrUnexpectedSigningMethod) ||
				errors.Is(err, ErrTokenIsNotValid)

			if !newTokenRequired {
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		if newTokenRequired {
			userID, err = generateUserID()
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
			tokenString, err = generateJWTAuthToken(userID, []byte(cfg.SecretKey))
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
			http.SetCookie(rw, &http.Cookie{
				Name:     "auth_token",
				Value:    tokenString,
				Path:     "/",
				HttpOnly: true,
			})
		}

		if userID == "" {
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), contextutils.UserIDKey, userID)
		h.ServeHTTP(rw, r.WithContext(ctx))
	}
}
