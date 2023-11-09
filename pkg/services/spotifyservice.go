package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"t-murch/top-25-api/pkg/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func fetchSpotifyApi(ctx *gin.Context, endpoint string, method string, body []byte) (status string, aBody []byte, err error) {
	sessionToken, _ := ctx.Cookie("sessionToken") // Errors for this handled in middleware.
	if len(sessionToken) == 0 {
		errorMessage := fmt.Sprintf("Forbidden without session established. Please log in. Error=%s. \n", err)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": errorMessage})
		log.Println(errorMessage)
		return http.StatusText(500), []byte(errorMessage), nil
		// ctx.Abort()
	}

	log.Printf("** sessionToken = %s \n", sessionToken)
	session := sessions.Default(ctx)
	redisToken := session.Get(sessionToken)
	log.Printf("redisToken: %s \n", redisToken)
	if redisToken == nil {
		errorMessage := fmt.Sprintln("Cookie found without Session established. ")
		log.Println(errorMessage)
		ctx.SetCookie("sessionToken", sessionToken, 0, "/", "10.0.0.5", false, true)
		// ctx.JSON(http.StatusUnauthorized, gin.H{"error": errorMessage})
		return http.StatusText(500), []byte(errorMessage), nil
		// ctx.Abort()
	}
	// var token models.SpotifyTokenResponse
	var token string
	error := json.Unmarshal(session.Get(sessionToken).([]byte), &token)
	if error != nil {
		log.Println(err)
	}

	if len(token) == 0 {
		fmt.Println("token is empty, return out.")
		return "", nil, nil
	}

	log.Printf("passed in endpoint: %s \n", endpoint)
	bearer := "Bearer " + token

	client := &http.Client{}

	req, err := http.NewRequest(method, models.SpotifyUrl+endpoint, bytes.NewBuffer(body))
	if err != nil {
		log.Println(err)
		return "", nil, err
	}
	//  else {
	// 	log.Printf("Built Request=%s \n", req.Body)
	// }

	for name, vals := range req.Header {
		for _, value := range vals {
			log.Printf("Headers: name=%s, val=%s", name, value)
		}
	}

	req.Header.Add("Authorization", bearer)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error with Spotify Auth. Error=%s \n", err)
		return http.StatusText(500), nil, err
	}

	defer resp.Body.Close()
	aBody, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error parsing Spotify response. Error=%s, Body=%s. \n", err, aBody)
		return "", nil, err
	}

	if resp.StatusCode > 399 {
		log.Println("Failed to fetch data from Spotify.")
		log.Printf("Response Status=%s \n", resp.Status)
		log.Printf("Response Body=%s \n", aBody)
		return "", nil, err
	}

	return resp.Status, aBody, nil
}

func GetTopTracks(ctx *gin.Context, itemType models.TopItems, term models.TopItemsTerm) (string, []byte, error) {
	return fetchSpotifyApi(ctx, fmt.Sprintf("v1/me/top/%s?time_range=%s&limit=%d", string(itemType), string(term), 25), http.MethodGet, nil)
}

func GetCurrentProfile(ctx *gin.Context) (string, []byte, error) {
	return fetchSpotifyApi(ctx, "v1/me", http.MethodGet, nil)
}

func GetUserProfile(ctx *gin.Context, userID string) (string, []byte, error) {
	return fetchSpotifyApi(ctx, fmt.Sprintf("v1/users/%s", string(userID)), http.MethodGet, nil)
}

func GetNewReleases(ctx *gin.Context) (string, []byte, error) {
	return fetchSpotifyApi(ctx, "v1/browse/new-releases", http.MethodGet, nil)
}

/*
We will create users in the db only after they've saved a playlist.
*/
func CreatePlaylist(ctx *gin.Context, title string, userID string, description string) (string, []byte, error) {
	createPlaylistPayload := map[string]interface{}{
		"name":        title,
		"description": description,
	}

	requestBody, err := json.Marshal(createPlaylistPayload)
	if err != nil {
		log.Printf("Failed to serialize createPlaylistPayload. UserID=%s, PlaylistTitle=%s. Error=%s. \n", userID, title, err)
		return "", nil, err
	}

	status, data, error := fetchSpotifyApi(ctx, fmt.Sprintf("v1/users/%s/playlists", string(userID)), http.MethodPost, requestBody)
	if error != nil {
		log.Printf("Failed to create playlist, will not continue to addTo if expected. spotUserID=%s. Error=%s \n", userID, error)
		ctx.Error(error)
		return status, nil, error
	}

	return status, data, nil
}

func AddItemsToPlaylist(ctx *gin.Context, userID string, playlistID string, tracks []string, position int) (string, []byte, error) {
	addToPlaylistPayload := map[string]interface{}{
		// "insert_before": 0,
		// "range_length":  25,
		// "range_start":   position,
		"uris": tracks,
		// "uris": []string{},
	}

	requestBody, err := json.Marshal(addToPlaylistPayload)
	if err != nil {
		log.Printf("Failed to serialize addToPlaylistPayload. UserID=%s, PlaylistID=%s. Error=%s. \n", userID, playlistID, err)
		return "", nil, err
	}

	status, response, error := fetchSpotifyApi(ctx, fmt.Sprintf("v1/playlists/%s/tracks", playlistID), http.MethodPut, requestBody)
	if error != nil {
		log.Printf("Failed to add items to a playlist. spotUserID=%s, playlistID=%s. Error=%s \n", userID, playlistID, error)
		ctx.Error(error)
		return status, nil, error
	}

	return status, response, nil
}

var topSongsUSA = "37i9dQZEVXbLp5XoPON0wI"

func getTopPlayedByCountry(ctx *gin.Context) (string, []byte, error) {
	return fetchSpotifyApi(ctx, "v1/browse/playlists/%s/tracks", topSongsUSA, nil)
}

func HeaderLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		headers := ctx.Request.Header
		fmt.Printf("Incoming headers = %s \n", headers)

		cookieExplicit, _ := ctx.Cookie("sessionToken")

		fmt.Printf("Incoming cookieExplicit = %s \n", cookieExplicit)
		ctx.Next()
	}
}
