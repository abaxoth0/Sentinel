package responsebody

type Token struct {
	Message     string `json:"message"`
	AccessToken string `json:"accessToken"`
    ExpiresIn   int    `json:"expiresIn"`
}

type Message struct {
	Message string `json:"message"`
}

type IsLoginAvailable struct {
	Available bool `json:"available"`
}

type Roles struct {
	Roles []string `json:"roles"`
}

