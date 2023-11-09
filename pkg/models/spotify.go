package models

import (
	"fmt"
	"time"
)

var SpotifyUrl = "https://api.spotify.com/"
var LOGIN_REDIRECT_URL = "https://accounts.spotify.com/authorize"
var SCOPES = "playlist-modify-private playlist-modify-public playlist-read-private user-read-email user-read-private user-top-read"
var PlaylistDescription = fmt.Sprintf("Built by I Miss My Top 25 by https://github.com/t-murch - %s", time.Now().Format("2006-Jan-02"))

type SpotifyTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type TopItems string

const (
	ArtistsType TopItems = "artists"
	TracksType  TopItems = "tracks"
)

type TopItemsTerm string

const (
	Now     TopItemsTerm = "short_term"
	Recent  TopItemsTerm = "medium_term"
	Distant TopItemsTerm = "long_term"
)

type ExplicitContent struct {
	FilterEnabled bool `json:"filter_enabled"`
	FilterLocked  bool `json:"filter_locked"`
}

type Image struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

type UserSpotify struct {
	DisplayName     string          `json:"display_name"`
	ExternalURLs    ExternalURLs    `json:"external_urls"`
	Href            string          `json:"href"`
	ID              string          `json:"id"`
	Images          []Image         `json:"images"`
	Type            string          `json:"type"`
	URI             string          `json:"uri"`
	Followers       Followers       `json:"followers"`
	Country         string          `json:"country"`
	Product         string          `json:"product"`
	ExplicitContent ExplicitContent `json:"explicit_content"`
	Email           string          `json:"email"`
}

type AddToPlaylistPayload struct {
	UserID     string   `json:"userID"`
	PlaylistID string   `json:"playlistID"`
	Position   int      `json:"position"`
	Items      []string `json:"items"`
}

type ClientCreatePayload struct {
	Title         string   `json:"title"`
	Items         []string `json:"items"`
	SpotifyUserID string   `json:"userID"`
	Description   string   `json:"description"`
	Term          string   `json:"term"`
}

type CreatePlaylistSpot struct {
	Uris     []string `json:"uris"`
	Position int      `json:"position"`
}

type ExternalURLs struct {
	Spotify string `json:"spotify"`
}

type Followers struct {
	Href  interface{} `json:"href"`
	Total int         `json:"total"`
}

type Owner struct {
	DisplayName  string       `json:"display_name"`
	ExternalURLs ExternalURLs `json:"external_urls"`
	Href         string       `json:"href"`
	ID           string       `json:"id"`
	Type         string       `json:"type"`
	URI          string       `json:"uri"`
}

type Tracks struct {
	Href     string        `json:"href"`
	Items    []interface{} `json:"items"`
	Limit    int           `json:"limit"`
	Next     interface{}   `json:"next"`
	Offset   int           `json:"offset"`
	Previous interface{}   `json:"previous"`
	Total    int           `json:"total"`
}

type Playlist struct {
	Collaborative bool          `json:"collaborative"`
	Description   string        `json:"description"`
	ExternalURLs  ExternalURLs  `json:"external_urls"`
	Followers     Followers     `json:"followers"`
	Href          string        `json:"href"`
	ID            string        `json:"id"`
	Images        []interface{} `json:"images"`
	Name          string        `json:"name"`
	Owner         Owner         `json:"owner"`
	PrimaryColor  interface{}   `json:"primary_color"`
	Public        bool          `json:"public"`
	SnapshotID    string        `json:"snapshot_id"`
	Tracks        Tracks        `json:"tracks"`
	Type          string        `json:"type"`
	URI           string        `json:"uri"`
}
