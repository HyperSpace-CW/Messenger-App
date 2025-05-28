package utils

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"strconv"
	"strings"
)

func ExtractUserIDFromHeader(header string, secretKey string) (int64, error) {
	if !strings.HasPrefix(header, "Bearer ") {
		return 0, errors.New("invalid authorization header format")
	}

	tokenString := strings.TrimPrefix(header, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Проверяем, является ли id числом
		if userID, ok := claims["id"].(float64); ok {
			return int64(userID), nil
		}
		// Проверяем, является ли id строкой
		if userIDStr, ok := claims["id"].(string); ok {
			userID, err := strconv.ParseInt(userIDStr, 10, 64)
			if err != nil {
				return 0, errors.New("invalid user id format in token")
			}
			return userID, nil
		}
		return 0, errors.New("user id not found in token")
	}

	return 0, errors.New("invalid token")
}
