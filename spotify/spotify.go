package spotify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"playlist-maker/structs"
	"strings"

	"golang.org/x/oauth2"
)

func containsString(slice []string, el string) bool {
	for _, s := range slice {
		if strings.Contains(strings.ToLower(s), strings.ToLower(el)) {
			return true
		}
	}
	return false
}

func Authorize(ctx context.Context, clientID, clientSecret string) (string, error) {
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

func makeRequest(method, url string, body io.Reader, errorMessage string, headers []structs.Header, goodResponse int) ([]byte, error, int) {
	client := &http.Client{}
	req, errReq := http.NewRequest(method, url ,body)
	if errReq != nil {
		return nil, fmt.Errorf("Request error on %s: %w", errorMessage, errReq), 0
	}
	for _, header := range headers {
		req.Header.Add(header.Key, header.Value)
	}

	resp, errResp := client.Do(req)
	if errResp != nil {
		return nil, fmt.Errorf("Responce error on %s: %w", errorMessage, errResp), 0
	}

	defer resp.Body.Close()
	if resp.StatusCode == goodResponse {
		body, errBody := io.ReadAll(resp.Body)
		if errBody != nil {
			return nil, fmt.Errorf("Body error on %s, %w", errorMessage, errBody), resp.StatusCode
		}
		return body, nil, resp.StatusCode
	}
	return nil, fmt.Errorf("Status code on %s: %d", errorMessage, resp.StatusCode), resp.StatusCode
}

func SearchAlbum(accessToken string, album structs.Album) (string, error) {
	params := url.Values{}
	params.Add("q", album.Title)
	encodedParams := params.Encode()

	url := fmt.Sprintf("https://api.spotify.com/v1/search?%s&type=album", encodedParams)

	var headers []structs.Header
	headers = append(headers, structs.Header{Key: "Authorization", Value: "Bearer " + accessToken})
	
	for i := 0; i < 5; i++ { 
		body, errBody, _ := makeRequest("GET", 
			url, 
			nil, 
			"search album", 
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
			if strings.Contains(strings.ToLower(item.Name), strings.ToLower(album.Title)) {
				for _, artist := range item.Artists {
					if containsString(album.Artist, artist.Name) {
						return item.ID, nil
					}
				}
			}
		}
		url = searchAlbum.Albums.Next
	}
	return "", fmt.Errorf("Album %s is not available on Spotify", album.Title)
}

func GetUserId(accessToken string) (string, error) {
	var headers []structs.Header
	headers = append(headers, structs.Header{Key: "Authorization", Value: "Bearer " + accessToken})
	body, errBody, _ := makeRequest("GET",
		"https://api.spotify.com/v1/me", 
		nil,
		"get user id",
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

func CreatePlaylist(accessToken, name, userId string) (string, error) {
	postBody := []byte(fmt.Sprintf(`{
		"name": "%s"
	}`, name))

	var headers []structs.Header
	headers = append(headers, structs.Header{Key: "Authorization", Value: "Bearer " + accessToken})
	headers = append(headers, structs.Header{Key: "Content-type", Value: "application/json"})
	
	body, errBody, _ := makeRequest("POST",
		fmt.Sprintf("https://api.spotify.com/v1/users/%s/playlists", userId),
		bytes.NewBuffer(postBody),
		"create playlist",
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

func GetAlbumTracksUris(accessToken, albumID string) ([]string, error) {
	var headers []structs.Header
	headers = append(headers, structs.Header{Key: "Authorization", Value: "Bearer " + accessToken})

	body, errBody, _ := makeRequest("GET",
		fmt.Sprintf("https://api.spotify.com/v1/albums/%s/tracks", albumID),
		nil,
		"get album tracks ids",
		headers,
		http.StatusOK,
	)

	if errBody != nil {
		return nil, fmt.Errorf("Error: %w", errBody)
	}

	var albumTracks structs.GetAlbumTracksIDs
	errJson := json.Unmarshal(body, &albumTracks)
	if errJson != nil {
		return nil, fmt.Errorf("Error on reading json: %w", errJson)
	}

	var tracksUris []string
	for _, track := range albumTracks.Items {
		tracksUris = append(tracksUris, track.URI)
	}
	return tracksUris, nil
}

func AddTracksToPlaylist(accessToken string, trackUris []string, playlistID string) error {
	postBody := structs.AddTracksRequest{Uris: trackUris, Position: 0}
	jsonReq, errJsonReq := json.Marshal(postBody)
	if errJsonReq != nil {
		return fmt.Errorf("Error on making json request: %w", errJsonReq)
	}

	var headers []structs.Header
	headers = append(headers, structs.Header{Key: "Authorization", Value: "Bearer " + accessToken})
	headers = append(headers, structs.Header{Key: "Content-Type", Value: "application/json"})

	for attempt := 1; attempt < 5; attempt++ {
		_, errReq, status := makeRequest("POST",
			fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", playlistID),
			bytes.NewBuffer(jsonReq),
			"add tracks to playlist",
			headers,
			http.StatusCreated,
		)
		if status == http.StatusBadGateway {
			if attempt != 4 {
				continue
			}
		} else {
			break
		}
		if errReq != nil {
			return fmt.Errorf("Error: %w", errReq)
	
		}
	}
	
	return nil
}