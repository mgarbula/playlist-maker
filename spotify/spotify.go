package spotify

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"encoding/json"
)

func GetBearer(clientId string, clientSecret string) (string, error) {
	resp, err := http.PostForm("https://accounts.spotify.com/api/token",
		url.Values{"grant_type": {"client_credentials"}, "client_id": {clientId}, "client_secret": {clientSecret}})
	if err != nil {
		return "", fmt.Errorf("Error on getBearer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error on getBearer: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading response body: %w", err)
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(body, &m)
	if err != nil {
		return "", fmt.Errorf("Error: %w", err)
	}
	
	return m["access_token"].(string), nil
}
