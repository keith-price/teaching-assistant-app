package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GetHTTPClient returns an authenticated HTTP client for the given OAuth scopes.
// It reuses a cached token from tokenFile if available, otherwise prompts the
// user to authorize via browser and caches the new token.
func GetHTTPClient(ctx context.Context, credentialsFile, tokenFile string, scopes ...string) (*http.Client, error) {
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %w", err)
	}

	config, err := google.ConfigFromJSON(b, scopes...)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %w", err)
	}

	tok, err := tokenFromFile(tokenFile)
	if err != nil {
		return nil, fmt.Errorf("token missing or invalid, please run 'go run cmd/auth/main.go' to authorize")
	}

	return config.Client(ctx, tok), nil
}

// AuthorizeInteractively runs the full OAuth browser flow, prompting the
// user to visit a URL and paste the auth code. Only call this from a
// standalone CLI tool, NEVER from inside a TUI.
func AuthorizeInteractively(ctx context.Context, credentialsFile, tokenFile string, scopes ...string) error {
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return fmt.Errorf("unable to read client secret file: %w", err)
	}

	config, err := google.ConfigFromJSON(b, scopes...)
	if err != nil {
		return fmt.Errorf("unable to parse client secret file to config: %w", err)
	}

	tok, err := getTokenFromWeb(ctx, config)
	if err != nil {
		return err
	}
	if err := saveToken(tokenFile, tok); err != nil {
		return err
	}
	return nil
}

// getTokenFromWeb requests a token from the web, then returns the retrieved token.
// Note: Static state token is acceptable for local-only desktop OAuth flow.
func getTokenFromWeb(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the authorization code: \n%v\n> ", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %w", err)
	}

	tok, err := config.Exchange(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %w", err)
	}
	return tok, nil
}

// tokenFromFile retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// saveToken saves a token to a file path.
func saveToken(path string, token *oauth2.Token) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %w", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}
