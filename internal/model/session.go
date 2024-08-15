package model

type Session struct {
	ID         int    `json:"id"`
	UserID     int    `json:"user_id"`
	ATokenID   string `json:"access_token_id"`
	RTokenHash string `json:"refresh_token_hash"`
	IP         string `json:"ip"`
	CreatedAt  int64  `json:"created_at"`
	Version    int64  `json:"version"`
}
