# Playlist maker
## Program that fetches the newest reviews from [Pitchfork](https://pitchfork.com/) and creates playlist on [Spotify](https://open.spotify.com/)

User will put genres and rates range which it is interested in. Program will webscrap the [Pitchfork](https://pitchfork.com/reviews/albums) website and create a playlist on [Spotify](https://open.spotify.com/) which will contain 10 newest albums which will follow user's requirenments. It will pick 3 random songs from each album.

--------
## Register app
To run the program you need to create a Spotify app. Follow instructions under [this link](https://developer.spotify.com/documentation/web-api/tutorials/getting-started#create-an-app) (`Only Create an app` step). While creation add `http://localhost:8080` to `Redirect URIs`. This is used to run temporary server while authentication.

After app is created copy-paste `Cliend ID` and `Client secret` and paste it to `config.json` file.

--------
## Run
To run app run
```
go run . -minRate {value} -maxRate {value} -playlistName="{value}" -albumsNumber {value} {list of genres}
```
e.g.
```
go run . -minRate 7.0 -maxRate 9.5 -playlistName="My playlist" -albumsNumber 10 Rock Experimental Jazz
```
All flags can be ommited - default values will be used.

--------
## Valid genres
Valid genres are: Rock, Pop/R&B, Folk/Country, Experimental, Jazz, Rap, Electronic. Be careful and set proper genres!

--------
## Authorization
After app running you will be asked to click the link and give permisions to the Spotify app. When all is set up, temporary server will inform you that you could close the window.

--------
## Final comment
Due to go routines there is a bit of randomness on app creation. Running app with the same flags and genres list can generate completely different playlists. Be aware of that!!!
All in all have fun and listen to good music!