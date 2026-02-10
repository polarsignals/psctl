package cliauth

const (
	DefaultAPIURL   = "https://api.polarsignals.com"
	DefaultClientID = "polarsignals-cli"
)

var DefaultScopes = []string{"openid", "profile", "email", "offline_access"}
