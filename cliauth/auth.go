package cliauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

type Authenticator struct {
	APIURL   string
	ClientID string
	Scopes   []string
	NoOpen   bool
}

func NewAuthenticator() *Authenticator {
	return &Authenticator{
		APIURL:   DefaultAPIURL,
		ClientID: DefaultClientID,
		Scopes:   DefaultScopes,
	}
}

func (a *Authenticator) Login(ctx context.Context) (*TokenInfo, error) {
	metadata, err := DiscoverOAuthMetadata(ctx, a.APIURL)
	if err != nil {
		return nil, fmt.Errorf("discover oauth metadata: %w", err)
	}

	callbackServer, err := NewCallbackServer()
	if err != nil {
		return nil, fmt.Errorf("create callback server: %w", err)
	}
	defer callbackServer.Shutdown(ctx)

	pkce := GeneratePKCE()

	state, err := generateState()
	if err != nil {
		return nil, fmt.Errorf("generate state: %w", err)
	}

	// Parse the authorization endpoint which may already contain query
	// parameters returned by the server (e.g. auth_endpoint). We retain
	// those and add the standard OAuth parameters on top.
	authURL, err := url.Parse(metadata.AuthorizationEndpoint)
	if err != nil {
		return nil, fmt.Errorf("parse authorization endpoint: %w", err)
	}

	q := authURL.Query()
	q.Set("client_id", a.ClientID)
	q.Set("redirect_uri", callbackServer.RedirectURL())
	q.Set("response_type", "code")
	q.Set("scope", strings.Join(a.Scopes, " "))
	q.Set("state", state)
	q.Set("code_challenge", pkce.CodeChallenge)
	q.Set("code_challenge_method", pkce.CodeChallengeMethod)
	authURL.RawQuery = q.Encode()

	browserURL := authURL.String()

	callbackServer.Start()

	if a.NoOpen {
		fmt.Printf("Open the following URL in your browser:\n\n%s\n\n", browserURL)
	} else {
		fmt.Println("Opening browser for authentication...")
		if err := OpenBrowser(browserURL); err != nil {
			fmt.Printf("Could not open browser automatically.\nOpen the following URL in your browser:\n\n%s\n\n", browserURL)
		}
	}

	fmt.Println("Waiting for authentication...")

	callbackCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	result, err := callbackServer.WaitForCallback(callbackCtx)
	if err != nil {
		return nil, fmt.Errorf("wait for callback: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("authentication error: %s", result.Error)
	}

	if result.State != state {
		return nil, fmt.Errorf("state mismatch")
	}

	oauth2Config := oauth2.Config{
		ClientID: a.ClientID,
		Endpoint: oauth2.Endpoint{
			TokenURL: metadata.TokenEndpoint,
		},
		RedirectURL: callbackServer.RedirectURL(),
		Scopes:      a.Scopes,
	}

	token, err := oauth2Config.Exchange(ctx, result.Code, oauth2.VerifierOption(pkce.Verifier))
	if err != nil {
		return nil, fmt.Errorf("exchange code for token: %w", err)
	}

	return &TokenInfo{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
	}, nil
}

func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
