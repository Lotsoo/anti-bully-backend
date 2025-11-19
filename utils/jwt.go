package handlers

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// We'll keep simple HS256 tokens with user_id and role claims.
func generateJWT(userID uint, role, secret string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

func parseJWT(tokenStr, secret string) (jwt.MapClaims, error) {
	parser := jwt.NewParser()
	tok, err := parser.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := tok.Claims.(jwt.MapClaims); ok && tok.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenMalformed
}
