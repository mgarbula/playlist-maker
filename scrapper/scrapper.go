package scrapper

import (
	"fmt"
	"playlist-maker/structs"
	"strconv"
	"time"
	"github.com/gocolly/colly"
)

var albums []structs.Album

func GetAlbums() []structs.Album {
	c := colly.NewCollector(
		colly.AllowedDomains("www.pitchfork.com", "pitchfork.com"),
	)

	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println("Something went wrong: ", err)
	})

	c.OnHTML("div.summary-item", func(albumElement *colly.HTMLElement) {
		album := structs.Album{}
		albumElement.ForEach("span.rubric__name", func(_ int, e *colly.HTMLElement) {
			album.Genre = append(album.Genre, e.Text)
		})
		album.Title = albumElement.ChildText("h3.summary-item__hed")
		album.Artist = albumElement.ChildText("div.summary-item__sub-hed")
		album.RatingHref = albumElement.ChildAttr("a", "href")
		albums = append(albums, album)
	})

	c.Visit("https://www.pitchfork.com/reviews/albums")
	return albums
}

func GetRating(album *structs.Album, channel chan<-*structs.Album, sc *structs.SafeCounter) {
	c := colly.NewCollector(
		colly.AllowedDomains("www.pitchfork.com", "pitchfork.com"),
	)

	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println("Something went wrong: ", err)
	})

	c.OnHTML("p.Rating-bkjebD", func(ratingElement *colly.HTMLElement) {
		ratingAsText := ratingElement.Text
		rating, _ := strconv.ParseFloat(ratingAsText, 32)
		album.Rating = float32(rating)
	})

	sc.Mu.Lock()
	sc.Counter++
	if sc.Counter == structs.MaxCalls {
		time.Sleep(time.Second)
		sc.Counter = 0
	}
	sc.Mu.Unlock()
	c.Visit("https://www.pitchfork.com" + album.RatingHref)

	channel <- album
}
