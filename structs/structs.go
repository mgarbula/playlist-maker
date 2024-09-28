package structs

import (
	"fmt"
	"sync"
)

type Album struct {
	Title, RatingHref string
	Rating            float32
	Genre, Artist 	  []string
}

func (album Album) String() string {
	return fmt.Sprintf("Title: %s, artist: %s, genre: %s, rating: %f", album.Title, album.Artist, album.Genre, album.Rating)
}

type SafeCounter struct {
	Mu      sync.Mutex
	Counter, AllCalls int
}

type SafeReady struct {
	Mu sync.Mutex
	Ready bool
}

type SafeAlbumsCounter struct {
	Mu sync.Mutex
	Counter int
}

type Configuration struct {
	ClientId, ClientSecret string
}

type SearchAlbum struct {
	Albums struct {
		Href  string `json:"href"`
		Items []struct {
			AlbumType string `json:"album_type"`
			Artists   []struct {
				ExternalUrls struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
				Href string `json:"href"`
				ID   string `json:"id"`
				Name string `json:"name"`
				Type string `json:"type"`
				URI  string `json:"uri"`
			} `json:"artists"`
			AvailableMarkets []string `json:"available_markets"`
			ExternalUrls     struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href   string `json:"href"`
			ID     string `json:"id"`
			Images []struct {
				Height int    `json:"height"`
				URL    string `json:"url"`
				Width  int    `json:"width"`
			} `json:"images"`
			Name                 string `json:"name"`
			ReleaseDate          string `json:"release_date"`
			ReleaseDatePrecision string `json:"release_date_precision"`
			TotalTracks          int    `json:"total_tracks"`
			Type                 string `json:"type"`
			URI                  string `json:"uri"`
		} `json:"items"`
		Limit    int    `json:"limit"`
		Next     string `json:"next"`
		Offset   int    `json:"offset"`
		Previous any    `json:"previous"`
		Total    int    `json:"total"`
	} `json:"albums"`
}

const MaxCalls = 30
var AlbumsNumber int
var MinRate, MaxRate float64