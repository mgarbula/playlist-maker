package main

import (
	"flag"
	"fmt"
	"playlist-maker/scrapper"
	"playlist-maker/structs"
	"sync"
	"time"
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

func contains[K comparable](slice []K, el K) bool {
	if len(slice) == 0 {
		return true
	}
	for _, s := range slice {
		if el == s {
			return true
		}
	}
	return false
}

func containsGenre(genres []string, album structs.Album) bool {
	for _, genre := range album.Genre {
		if contains(genres, genre) {
			return true
		}
	}
	return false
}

func main() {
	flag.Float64Var(&structs.MinRate, "minRate", 7.0, "Minimum rating of an album (1-10)")
	flag.Float64Var(&structs.MaxRate, "maxRate", 10.0, "Maximum rating of an album (1-10)")
	flag.IntVar(&structs.AlbumsNumber, "albumsNumber", 10, "Number of albums to add to playlist")
	flag.Parse()
	genres := flag.Args()

	path := "https://www.pitchfork.com/reviews/albums"
	currentPage := 1

	sr := structs.SafeReady{Ready: false}
	go workInProgress(&sr)

	var wg sync.WaitGroup
	sc := structs.SafeCounter{Counter: 0, AllCalls: 0}
	sac := structs.SafeAlbumsCounter{Counter: 0}

	start := time.Now()

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
	sr.Ready = true

	var howManyAlbums int
	for album := range c {
		fmt.Println(album)
		howManyAlbums++
	}

	elapsed := time.Since(start)
	fmt.Println("\nExecution time: " + elapsed.String())
	fmt.Printf("howManyAlbums = %d\n", howManyAlbums)
	sac.Mu.Lock()
	fmt.Printf("sac.Counter = %d\n", sac.Counter)
	sac.Mu.Unlock()
}
