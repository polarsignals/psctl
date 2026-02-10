package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alecthomas/kong"

	"github.com/polarsignals/psctl/cliauth"
)

type CLI struct {
	Auth AuthCmd `cmd:"" help:"Authentication commands"`
}

type AuthCmd struct {
	Login  LoginCmd  `cmd:"" help:"Login to Polar Signals"`
	Status StatusCmd `cmd:"" help:"Show authentication status"`
}

type LoginCmd struct {
	NoOpen   bool   `help:"Print URL instead of opening browser"`
	APIURL   string `help:"API base URL for discovery" default:"${api_url}"`
	ClientID string `help:"OAuth client ID" default:"${client_id}"`
}

type StatusCmd struct{}

func (c *LoginCmd) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalCh
		cancel()
	}()

	auth := cliauth.NewAuthenticator()
	auth.NoOpen = c.NoOpen
	auth.APIURL = c.APIURL
	auth.ClientID = c.ClientID

	token, err := auth.Login(ctx)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	if err := cliauth.SaveCredentials(token); err != nil {
		return fmt.Errorf("save credentials: %w", err)
	}

	path, _ := cliauth.CredentialsPath()
	fmt.Printf("Successfully logged in!\nCredentials saved to %s\n", path)
	return nil
}

func (c *StatusCmd) Run() error {
	token, err := cliauth.LoadCredentials()
	if err != nil {
		fmt.Println("Not logged in.")
		fmt.Println("Run 'psctl auth login' to authenticate.")
		return nil
	}

	path, _ := cliauth.CredentialsPath()
	fmt.Println("Logged in.")
	fmt.Printf("Credentials file: %s\n", path)
	fmt.Printf("Token type: %s\n", token.TokenType)

	if !token.Expiry.IsZero() {
		if token.Expiry.Before(time.Now()) {
			fmt.Println("Token status: Expired")
		} else {
			fmt.Printf("Token expires: %s\n", token.Expiry.Format(time.RFC3339))
		}
	}

	if token.RefreshToken != "" {
		fmt.Println("Refresh token: Present")
	}

	return nil
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli,
		kong.Name("psctl"),
		kong.Description("Polar Signals CLI"),
		kong.UsageOnError(),
		kong.Vars{
			"api_url":   cliauth.DefaultAPIURL,
			"client_id": cliauth.DefaultClientID,
		},
	)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
