package spotify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"playlist-maker/global"
	"playlist-maker/structs"
	"strings"

	"golang.org/x/oauth2"
)

func Authorize(ctx context.Context, clientID string, clientSecret string) (string, error) {
	authURL := "https://accounts.spotify.com/authorize"

	config := &oauth2.Config{
		ClientID: clientID,
		ClientSecret: clientSecret,
		Scopes: []string{"user-read-private",
			"user-read-email",
			"playlist-modify-public",
			"playlist-modify-private",
		},
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

func makeRequest(method string, url string, body io.Reader, errorMessage string, client *http.Client, headers []structs.Header, goodResponse int) ([]byte, error) {
	req, errReq := http.NewRequest(method, url ,body)
	if errReq != nil {
		return nil, fmt.Errorf("Request error on %s: %w", errorMessage, errReq)
	}
	for _, header := range headers {
		req.Header.Add(header.Key, header.Value)
	}

	resp, errResp := client.Do(req)
	if errResp != nil {
		return nil, fmt.Errorf("Responce error on %s: %w", errorMessage, errResp)
	}

	defer resp.Body.Close()
	if resp.StatusCode == goodResponse {
		body, errBody := io.ReadAll(resp.Body)
		if errBody != nil {
			return nil, fmt.Errorf("Body error on %s, %w", errorMessage, errBody)
		}
		return body, nil
	}
	return nil, fmt.Errorf("Status code on %s: %d", errorMessage, resp.StatusCode)
}

func SearchAlbum(album structs.Album, bearer string) (string, error) {
	client := &http.Client{}

	params := url.Values{}
	params.Add("q", album.Title)
	encodedParams := params.Encode()

	var headers []structs.Header
	headers = append(headers, structs.Header{Key: "Authorization", Value: "Bearer " + bearer})
	body, errBody := makeRequest("GET", 
		fmt.Sprintf("https://api.spotify.com/v1/search?%s&type=album", encodedParams), 
		nil, 
		"search album", 
		client,
		headers,
		http.StatusOK,
	)
	if errBody != nil {
		return "", fmt.Errorf("Error: %w", errBody)
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
					return item.ID, nil
				}
			}
		}
	}

	return "", fmt.Errorf("Album is not available on Spotify")
}

func GetUserId(accessToken string) (string, error) {
	client := &http.Client{}
	var headers []structs.Header
	headers = append(headers, structs.Header{Key: "Authorization", Value: "Bearer " + accessToken})
	body, errBody := makeRequest("GET",
		"https://api.spotify.com/v1/me", 
		nil,
		"get user id",
		client,
		headers,
		http.StatusOK,
	)

	if errBody != nil {
		return "", fmt.Errorf("Error: %w", errBody)
	}

	var user structs.GetUserId
	errJson := json.Unmarshal(body, &user)
	if errJson != nil {
		return "", fmt.Errorf("Error on reading json: %w", errJson)
	}
	return user.ID, nil
}

func CreatePlaylist(accessToken string, name string, userId string) (string, error) {
	postBody := []byte(fmt.Sprintf(`{
		"name": "%s"
	}`, name))

	client := &http.Client{}
	var headers []structs.Header
	headers = append(headers, structs.Header{Key: "Authorization", Value: "Bearer " + accessToken})
	headers = append(headers, structs.Header{Key: "Content-type", Value: "application/json"})
	
	body, errBody := makeRequest("POST",
		fmt.Sprintf("https://api.spotify.com/v1/users/%s/playlists", userId),
		bytes.NewBuffer(postBody),
		"create playlist",
		client,
		headers,
		http.StatusCreated,
	)

	if errBody != nil {
		return "", fmt.Errorf("Error: %w", errBody)
	}

	var playlist structs.CreatePlaylist
	errJson := json.Unmarshal(body, &playlist)
	if errJson != nil {
		return "", fmt.Errorf("Error on reading json: %w", errJson)
	}
	return playlist.ID, nil
}