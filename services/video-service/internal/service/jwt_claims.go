// internal/service/jwt_claims.go
package service

import "github.com/golang-jwt/jwt/v5"

//
// JWT Custom Claims
// NOTE: This is a temporary copy from user-service.
// We will move this to a shared library later.
//

type JwtCustomClaims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}
