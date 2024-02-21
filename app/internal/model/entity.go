package model

type CallbackGoogleRequest struct {
	AuthCode string `json:"authCode"`
}

type CallbackGoogleResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type LoginGoogleCredentialRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type GetNewAccessTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}
