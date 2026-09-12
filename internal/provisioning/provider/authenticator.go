package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type authToken struct {
	AccessToken string `json:"access_token,omitempty"`
	ExpiresIn   int64  `json:"expires_in,omitempty"`
	IssuedAt    int64  // not in the response
}

func (token *authToken) isExpired() bool {
	now := time.Now().Unix()
	expiresAt := token.IssuedAt + (token.ExpiresIn - 10) // with 10s grace
	return now >= expiresAt
}

type authenticator struct {
	next         http.RoundTripper
	client       *http.Client
	authUrl      string
	clientId     string
	clientSecret string
	token        *authToken
	mu           sync.Mutex
}

func newAuthenticator(authUrl, clientId, clientSecret string, next http.RoundTripper) (*authenticator, error) {
	if authUrl == "" {
		return nil, errors.New("authUrl is required for authenticator")
	}
	if clientSecret == "" {
		return nil, errors.New("clientSecret is required for authenticator")
	}
	if clientId == "" {
		return nil, errors.New("clientId is required for authenticator")
	}
	client := http.DefaultClient
	return &authenticator{
		client:       client,
		clientId:     clientId,
		clientSecret: clientSecret,
		next:         next,
		authUrl:      authUrl,
		token:        nil,
	}, nil
}

func (at *authenticator) fetchToken(ctx context.Context) error {
	form := url.Values{}
	form.Set("client_id", at.clientId)
	form.Set("client_secret", at.clientSecret)
	form.Set("grant_type", "client_credentials")
	log.Printf("Provider authentication status update")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, at.authUrl, strings.NewReader(form.Encode()))
	if err != nil {
		log.Printf("Provider authentication operation failed; details omitted")
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := at.client.Do(req)
	if err != nil {
		log.Printf("Provider authentication operation failed; details omitted")
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("unexpected response code from auth, expected 200 got %d", res.StatusCode)
	}
	var token authToken
	err = json.NewDecoder(res.Body).Decode(&token)
	if err != nil {
		log.Printf("Provider authentication operation failed; details omitted")
		return err
	}
	token.IssuedAt = time.Now().Unix()
	at.token = &token
	log.Printf("Provider authentication status update")
	return nil
}

func (at *authenticator) RoundTrip(original *http.Request) (*http.Response, error) {
	at.mu.Lock()
	defer at.mu.Unlock()
	req := original.Clone(original.Context())
	token, err := at.getToken(req.Context())
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	return at.next.RoundTrip(req)
}

func (at *authenticator) getToken(ctx context.Context) (*authToken, error) {
	if at.token == nil || at.token.isExpired() {
		if err := at.fetchToken(ctx); err != nil {
			return nil, err
		}
	}
	return at.token, nil
}
