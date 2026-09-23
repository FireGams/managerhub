// Package auth provides password hashing, JWT sessions and RBAC helpers.
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Claims is the JWT payload for web users.
type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// HashPassword returns a bcrypt hash of the password.
func HashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}

// CheckPassword compares a plaintext password with a stored hash.
func CheckPassword(hash, pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

// Issuer used in every JWT.
const Issuer = "managerhub"

// IssueToken signs a session token for a user.
func IssueToken(secret, username, role string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    Issuer,
			Subject:   username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

// ParseToken validates a session token.
func ParseToken(secret, token string) (Claims, error) {
	var claims Claims
	_, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("auth: unexpected signing method")
		}
		return []byte(secret), nil
	})
	return claims, err
}

// RoleAtLeast reports whether role has at least the level of min.
func RoleAtLeast(role, min string) bool {
	rank := map[string]int{"viewer": 1, "operator": 2, "admin": 3}
	return rank[role] >= rank[min]
}
