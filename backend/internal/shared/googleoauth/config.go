package googleoauth

import (
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	DeveloperTok string
	LoginCID     string
}

func Load() Config {
	return Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		DeveloperTok: os.Getenv("GOOGLE_DEVELOPER_TOKEN"),
		LoginCID:     os.Getenv("GOOGLE_LOGIN_CUSTOMER_ID"),
	}
}

func OAuth2(cfg Config) *oauth2.Config {
	ep := google.Endpoint
	ep.AuthURL = "https://accounts.google.com/o/oauth2/v2/auth"
	ep.TokenURL = "https://oauth2.googleapis.com/token"

	return &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/adwords",
			"openid", "email", // можно убрать временно, если экран согласия не донастроен
		},
		Endpoint: ep,
	}
}
