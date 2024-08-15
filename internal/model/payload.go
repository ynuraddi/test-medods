package model

import (
	"github.com/golang-jwt/jwt/v5"
)

type Payload struct {
	UserID int    `json:"user_id"`
	IP     string `json:"ip"`
	jwt.RegisteredClaims
}
