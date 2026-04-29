package response

type LoginResponse struct {
	Token        string `json:"token" comment:"登录token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresAt    int64  `json:"expires_at"`
}
