// libs/go-common/types/jwt/claims.go
package jwt

import "github.com/golang-jwt/jwt/v5"

// JwtCustomClaims определяет стандартную структуру данных,
// которую мы помещаем в JWT. Все сервисы будут использовать ее.
type JwtCustomClaims struct {
	UserID      int64                  `json:"user_id"`
	Role        string                 `json:"role,omitempty"`
	Permissions map[string]interface{} `json:"perms,omitempty"`
	jwt.RegisteredClaims
}
