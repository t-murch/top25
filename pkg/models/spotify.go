package models

var SpotifyUrl = "https://api.spotify.com/"
var LOGIN_REDIRECT_URL = "https://accounts.spotify.com/authorize"
var SCOPES = "playlist-modify-private playlist-modify-public playlist-read-private user-read-email user-read-private user-top-read"

type SpotifyTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type TopItems string

const (
	Artists TopItems = "artists"
	Tracks  TopItems = "tracks"
)

type TopItemsTerm string

const (
	Now     TopItemsTerm = "short_term"
	Recent  TopItemsTerm = "medium_term"
	Distant TopItemsTerm = "long_term"
)
