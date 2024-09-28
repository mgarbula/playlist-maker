package spotify

import (
	"fmt"
	"net/http"
	"context"
	"golang.org/x/oauth2"
)

func Authorize(ctx context.Context, clientID string, clientSecret string) (string, error) {
	authURL := "https://accounts.spotify.com/authorize"

	config := &oauth2.Config{
		ClientID: clientID,
		ClientSecret: clientSecret,
		Scopes: []string{"user-read-private", "user-read-email"},
		Endpoint: oauth2.Endpoint{
			AuthURL: authURL,
			TokenURL: "https://accounts.spotify.com/api/token",
		},
		RedirectURL: "http://localhost:8080",
	}

	accessTokenChan := make(chan string, 1)

	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			code := r.URL.Query().Get("code")
			if code != "" {
				token, errToken := config.Exchange(ctx, code)
				if errToken != nil {
					fmt.Println("Error while token exchange: %w", errToken)
					return
				}
				accessTokenChan <- token.AccessToken
				w.Write([]byte("You can close that window!"))
			}
		})
		http.ListenAndServe(":8080", nil)
	}()

	fmt.Println("Please open the following URL in your web browser:")
	fmt.Println(config.AuthCodeURL(""))

	select {
	case accessToken := <- accessTokenChan:
		return accessToken, nil
	case <- ctx.Done():
		fmt.Println("Authorization timed out!")
		return "", ctx.Err()
	}	
}
