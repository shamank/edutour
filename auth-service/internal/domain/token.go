package domain

type RefreshToken struct {
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type Token struct {
}
