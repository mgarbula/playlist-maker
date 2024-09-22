package structs

import (
	"fmt"
	"sync"
)

type Album struct {
	Title, Artist, RatingHref string
	Rating                    float32
	Genre                     []string
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

const MaxCalls = 30