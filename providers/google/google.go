package google

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "strings"
    "time"

    "github.com/ahaodev/authx"
)

// Config contains the Google OAuth configuration.
type Config struct {
    ClientID     string
    ClientSecret string
    RedirectURL  string
    Scopes       []string
    AuthURL      string
    TokenURL     string
    UserInfoURL  string
}

// Provider implements the Google OAuth provider.
type Provider struct {
    config Config
    client *http.Client
}

// New creates a Google provider from a config map.
func New(cfg map[string]string) (authx.Provider, error) {
    provider := &Provider{
        config: Config{
            ClientID:     cfg["client_id"],
            ClientSecret: cfg["client_secret"],
            RedirectURL:  cfg["redirect_url"],
            Scopes:       []string{"openid", "profile", "email"},
            AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
            TokenURL:     "https://oauth2.googleapis.com/token",
            UserInfoURL:  "https://www.googleapis.com/oauth2/v3/userinfo",
        },
        client: &http.Client{Timeout: 10 * time.Second},
    }

    if provider.config.ClientID == "" || provider.config.ClientSecret == "" || provider.config.RedirectURL == "" {
        return nil, fmt.Errorf("google provider requires client_id, client_secret, and redirect_url")
    }

    return provider, nil
}

func init() {
    authx.Register("google", New)
}

// Name returns the provider name.
func (p *Provider) Name() string {
    return "google"
}

// AuthURL builds the Google authorization URL.
func (p *Provider) AuthURL(state string) (string, error) {
    endpoint, err := url.Parse(p.config.AuthURL)
    if err != nil {
        return "", err
    }

    values := endpoint.Query()
    values.Set("client_id", p.config.ClientID)
    values.Set("redirect_uri", p.config.RedirectURL)
    values.Set("response_type", "code")
    values.Set("scope", strings.Join(p.config.Scopes, " "))
    values.Set("state", state)
    endpoint.RawQuery = values.Encode()

    return endpoint.String(), nil
}

// Exchange exchanges the authorization code for an access token.
func (p *Provider) Exchange(ctx context.Context, code string) (*authx.Token, error) {
    data := url.Values{}
    data.Set("client_id", p.config.ClientID)
    data.Set("client_secret", p.config.ClientSecret)
    data.Set("code", code)
    data.Set("grant_type", "authorization_code")
    data.Set("redirect_uri", p.config.RedirectURL)

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.config.TokenURL, strings.NewReader(data.Encode()))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    resp, err := p.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("token exchange failed: %s", resp.Status)
    }

    var payload struct {
        AccessToken  string `json:"access_token"`
        RefreshToken string `json:"refresh_token"`
        TokenType    string `json:"token_type"`
        ExpiresIn    int    `json:"expires_in"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
        return nil, err
    }

    token := &authx.Token{
        AccessToken:  payload.AccessToken,
        RefreshToken: payload.RefreshToken,
        TokenType:    payload.TokenType,
    }
    if payload.ExpiresIn > 0 {
        token.Expiry = time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second)
    }

    return token, nil
}

// UserInfo retrieves profile information from the Google userinfo endpoint.
func (p *Provider) UserInfo(ctx context.Context, token *authx.Token) (*authx.Profile, error) {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.config.UserInfoURL, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Authorization", "Bearer "+token.AccessToken)

    resp, err := p.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("userinfo request failed: %s", resp.Status)
    }

    var payload struct {
        Sub    string `json:"sub"`
        Email  string `json:"email"`
        Name   string `json:"name"`
        Picture string `json:"picture"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
        return nil, err
    }

    raw := map[string]interface{}{
        "sub":     payload.Sub,
        "email":   payload.Email,
        "name":    payload.Name,
        "picture": payload.Picture,
    }

    return &authx.Profile{
        Provider:  p.Name(),
        Subject:   payload.Sub,
        Email:     payload.Email,
        Name:      payload.Name,
        AvatarURL: payload.Picture,
        Raw:       raw,
    }, nil
}
