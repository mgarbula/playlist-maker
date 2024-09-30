package main

import (
	"flag"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"playlist-maker/global"
	"playlist-maker/scrapper"
	"playlist-maker/spotify"
	"playlist-maker/structs"
	"sync"
	"time"
	"math/rand/v2"
)

func getRating(album structs.Album, sc *structs.SafeCounter, c *chan *structs.Album, wg *sync.WaitGroup, sac *structs.SafeAlbumsCounter) {
	defer wg.Done()
	scrapper.GetRating(&album, *c, sc, sac)
}

func workInProgress(sr *structs.SafeReady) {
	char := "-"
	for sr.Ready != true {
		fmt.Print("Work in progress" + char)
		time.Sleep(250 * time.Millisecond)
		if char == "-" {
			char = "\\"
		} else if char == "\\" {
			char = "|"
		} else if char == "|" {
			char = "/"
		} else if char == "/" {
			char = "-"
		}
		fmt.Print("\r")
	}
	fmt.Println()
}

func containsGenre(genres []string, album structs.Album) bool {
	for _, genre := range album.Genre {
		if global.Contains(genres, genre, true) {
			return true
		}
	}
	return false
}

func getRandomUris(tracksUris []string) []string {
	var randomUris []string
	var generatedRandoms []int
	for i := 0; i < 3; i++ {
		random := rand.IntN(len(tracksUris) - 1)
		for global.Contains(generatedRandoms, random, false) {
			random = rand.IntN(len(tracksUris) - 1)
		}
		generatedRandoms = append(generatedRandoms, random)
		randomUris = append(randomUris, tracksUris[random])
	}
	return randomUris
}

func addAlbumToPlaylist(bearer string, album structs.Album, playlistID string) error {
	albumID, errID := spotify.SearchAlbum(bearer, album)

	if errID != nil {
		return fmt.Errorf("\nError with %s album: %w", album.Title, errID)
	}

	tracksUris, errTracks := spotify.GetAlbumTracksUris(bearer, albumID)
	if errTracks != nil {
		return fmt.Errorf("\nError with %s album: %w", album.Title, errTracks)
	}

	threeTracksUris := getRandomUris(tracksUris)
	errAddTracks := spotify.AddTracksToPlaylist(bearer, threeTracksUris, playlistID)
	if errAddTracks != nil {
		return fmt.Errorf("\nError with %s album: %w", album.Title, errAddTracks)
	}
	return nil
}

func main() {
	flag.Float64Var(&structs.MinRate, "minRate", 7.0, "Minimum rating of an album (1-10)")
	flag.Float64Var(&structs.MaxRate, "maxRate", 10.0, "Maximum rating of an album (1-10)")
	flag.IntVar(&structs.AlbumsNumber, "albumsNumber", 10, "Number of albums to add to playlist")
	flag.StringVar(&structs.PlaylistName, "playlistName", "Playlist maker", "Playlist name")
	flag.Parse()
	genres := flag.Args()

	file, _ := os.Open("config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := structs.Configuration{}
	errDecode := decoder.Decode(&configuration)
	if errDecode != nil {
		fmt.Println("error: ", errDecode)
		return
	}

	ctx := context.Background()
	bearer, errBearer := spotify.Authorize(ctx, configuration.ClientId, configuration.ClientSecret)
	if errBearer != nil {
		fmt.Println(errBearer)
		return
	}

	path := "https://www.pitchfork.com/reviews/albums"
	currentPage := 1

	sr := structs.SafeReady{Ready: false}
	go workInProgress(&sr)

	var wg sync.WaitGroup
	sc := structs.SafeCounter{Counter: 0, AllCalls: 0}
	sac := structs.SafeAlbumsCounter{Counter: 0}

	c := make(chan *structs.Album, structs.AlbumsNumber)
	var albums []structs.Album

	sac.Mu.Lock()
	for sac.Counter < structs.AlbumsNumber {
		sac.Mu.Unlock()
		if currentPage == 1 {
			albums = scrapper.GetAlbums(path)
		} else {
			albums = scrapper.GetAlbums(fmt.Sprintf("%s%s%d", path, "?page=", currentPage))
		}
		currentPage++

		for _, album := range albums {
			sac.Mu.Lock()
			if sac.Counter == structs.AlbumsNumber {
				sac.Mu.Unlock()
				break
			}
			sac.Mu.Unlock()
			if containsGenre(genres, album) {
				wg.Add(1)
				go getRating(album, &sc, &c, &wg, &sac)
			}
		}
		wg.Wait()
		sac.Mu.Lock()
	}
	sac.Mu.Unlock()

	go func() {
		wg.Wait()
		close(c)
	}()

	userID, errUserID := spotify.GetUserId(bearer)
	if errUserID != nil {
		fmt.Println(errUserID)
		return
	}

	playlistID, errPlaylist := spotify.CreatePlaylist(bearer, structs.PlaylistName, userID)
	if errPlaylist != nil {
		fmt.Println(errPlaylist)
		return
	}

	errorChan := make(chan error)

	for album := range c {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := addAlbumToPlaylist(bearer, *album, playlistID)
			if err != nil {
				errorChan <- err
			}
		}()
	}
	
	wg.Wait()
	close(errorChan)
	sr.Ready = true

	if len(errorChan) > 0 {
		fmt.Println("\nPlaylist created with errors: ")
		for err := range errorChan {
			fmt.Println(err)
		}
	} else {
		fmt.Println("\nPlaylist created")
	}
}
