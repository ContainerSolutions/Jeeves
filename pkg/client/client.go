package client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ContainerSolutions/jeeves/pkg/config"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	githubBaseUrl = "https://api.github.com"
)

// JWTAuth token issued by Github in response to signed JWT Token
type JWTAuth struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type authClient struct {
	BaseUrl string
	Client  *http.Client
	Ctx     context.Context
}

//GetInstallationClient returns an authorized Github Client for an installation
func GetInstallationClient(installationID int64) (*github.Client, error) {
	ctx := context.Background()
	cfg := config.JeevesConfig{}
	err := cfg.GetConfig()
	if err != nil {
		return &github.Client{}, err
	}
	aClient := &authClient{
		BaseUrl: githubBaseUrl,
		Client:  http.DefaultClient,
	}
	token, err := getAccessToken(aClient, cfg, installationID)
	if err != nil {
		return &github.Client{}, err
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	return client, err
}

//GetAccessToken returns a Github OAuth Token
func getAccessToken(client *authClient, config config.JeevesConfig, installationID int64) (string, error) {
	installationToken, tokenErr := makeAccessTokenForInstallation(
		client,
		config.GithubApplicationID,
		installationID,
		config.GithubPrivateKey,
	)
	if tokenErr != nil {
		return "", tokenErr
	}
	return installationToken, nil
}

// MakeAccessTokenForInstallation makes an access token for an installation / private key
func makeAccessTokenForInstallation(
	c *authClient,
	appID string,
	installation int64,
	privateKey string,
) (string, error) {
	signed, err := getSignedJwtToken(appID, privateKey)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf(
		"%v/app/installations/%d/access_tokens",
		c.BaseUrl,
		installation,
	)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", signed))
	req.Header.Add("Accept", "application/vnd.github.machine-man-preview+json")

	res, err := c.Client.Do(req)

	if err != nil {
		return "", fmt.Errorf("error getting Access token %v", err)
	}

	defer res.Body.Close()
	bytesOut, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return "", readErr
	}

	jwtAuth := JWTAuth{}
	jsonErr := json.Unmarshal(bytesOut, &jwtAuth)
	return jwtAuth.Token, jsonErr
}

// GetSignedJwtToken get a tokens signed with private key
func getSignedJwtToken(appID string, privateKey string) (string, error) {

	keyBytes := []byte(privateKey)
	key, keyErr := jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	if keyErr != nil {
		return "", keyErr
	}

	now := time.Now()
	claims := jwt.StandardClaims{
		Issuer:    appID,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(time.Minute * 9).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	signedVal, signErr := token.SignedString(key)
	if signErr != nil {
		return "", signErr
	}
	return string(signedVal), nil
}
