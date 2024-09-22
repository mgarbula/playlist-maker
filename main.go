package main

import (
	"fmt"
	"playlist-maker/scrapper"
	"playlist-maker/structs"
	"sync"
	"time"
)

func getRatings(album structs.Album, sc *structs.SafeCounter, c *chan *structs.Album, wg *sync.WaitGroup) {
	defer wg.Done()
	//*readyAlbums = append(*readyAlbums, scrapper.GetRating(&album, *c, sc))
	scrapper.GetRating(&album, *c, sc)
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

// func getFromChan(c chan *structs.Album, size int) []*structs.Album {
// 	albums := make([]*structs.Album, size)
// 	for album := range c {
// 		albums = append(albums, album)
// 	}
// 	return albums
// }

func main() {
	sr := structs.SafeReady{Ready: false}
	go workInProgress(&sr)

	var wg sync.WaitGroup
	//var readyAlbums []*structs.Album
	sc := structs.SafeCounter{Counter: 0, AllCalls: 0}
	start := time.Now()

	albums := scrapper.GetAlbums()

	c := make(chan *structs.Album)
	for _, album := range albums {
		wg.Add(1)
		go getRatings(album, &sc, &c, &wg)
	}

	go func() {
		wg.Wait()
		close(c)
	}()
	sr.Ready = true

	//readyAlbums := getFromChan(c, len(albums))

	for album := range c {
		fmt.Println(album)
	}

	elapsed := time.Since(start)
	fmt.Println("\nExecution time: " + elapsed.String())
}
