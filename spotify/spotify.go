package spotify

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"playlist-maker/global"
	"playlist-maker/structs"
	"strings"
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

func SearchAlbum(album structs.Album, bearer string) (string, error) {
	client := &http.Client{}

	params := url.Values{}
	params.Add("q", album.Title)
	encodedParams := params.Encode()

	req, errReq := http.NewRequest("GET", fmt.Sprintf("https://api.spotify.com/v1/search?%s&type=album", encodedParams), nil)
	if errReq != nil {
		return "", fmt.Errorf("Error on SearchAlbum: %w", errReq)
	}
	req.Header.Add("Authorization", "Bearer "+bearer)

	resp, errResp := client.Do(req)
	if errResp != nil {
		return "", fmt.Errorf("Error sending HTTP request: %w", errResp)
	}

	body, errBody := io.ReadAll(resp.Body)
	if errBody != nil {
		return "", fmt.Errorf("Error on reading body (SearchAlbum): %w", errBody)
	}

	var searchAlbum structs.SearchAlbum
	errJson := json.Unmarshal(body, &searchAlbum)
	if errJson != nil {
		return "", fmt.Errorf("Error: %w", errJson)
	}

	for _, item := range searchAlbum.Albums.Items {
		if strings.EqualFold(item.Name, album.Title) {
			for _, artist := range item.Artists {
				if global.Contains(album.Artist, artist.Name) {
					return item.Href, nil
				}
			}
		}
	}

	return "", fmt.Errorf("Album is not available on Spotify")
}
