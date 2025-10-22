package oauthcontroller

type authProvider uint8

const (
	googleProvider authProvider = 1 << iota
)

var authProviderMap = map[authProvider]string{
	googleProvider: "google",
}

func (p authProvider) String() string {
	return authProviderMap[p]
}
