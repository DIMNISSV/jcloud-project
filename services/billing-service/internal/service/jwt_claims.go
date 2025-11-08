package service

import "github.com/golang-jwt/jwt/v5"

// JWT Custom Claims
type JwtCustomClaims struct {
	UserID      int64                  `json:"user_id"`
	Permissions map[string]interface{} `json:"perms,omitempty"`
	jwt.RegisteredClaims
}
