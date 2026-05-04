package entity

import "github.com/golang-jwt/jwt/v5"

// JWTClaims holds the JWT payload fields used by StreamPulse.
// Defined in entity so both usecase (sign) and middleware (verify) can import it
// without violating Clean Code dependency rules.
type JWTClaims struct {
	UserID string   `json:"sub"`
	Email  string   `json:"email"`
	Role   UserRole `json:"role"`
	jwt.RegisteredClaims
}
