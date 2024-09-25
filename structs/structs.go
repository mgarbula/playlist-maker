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

const MaxCalls = 30
var AlbumsNumber int
var MinRate, MaxRate float64