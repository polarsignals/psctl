package cliauth

import (
	"crypto/sha256"
	"encoding/base64"

	"golang.org/x/oauth2"
)

type PKCEParams struct {
	Verifier            string
	CodeChallenge       string
	CodeChallengeMethod string
}

func GeneratePKCE() PKCEParams {
	verifier := oauth2.GenerateVerifier()
	h := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(h[:])
	return PKCEParams{
		Verifier:            verifier,
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
	}
}
