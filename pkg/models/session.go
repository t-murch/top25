package models

import "time"

type session struct {
	appleToken   string
	spotifyToken SpotifyTokenResponse
	userId       string
	expiry       time.Time
}

func (s session) isExpired() bool {
	return s.expiry.Before(time.Now())
}
