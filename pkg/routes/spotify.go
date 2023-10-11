package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"t-murch/top-25-api/pkg/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func addSpotifyRoutes(rg *gin.RouterGroup) {
	// rg.Use(CookieTool())
	spotify := rg.Group("/spotify", CookieTool())

	spotify.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, "Spotify Base Route")
	})
	spotify.GET("/topTracks", func(ctx *gin.Context) {

	})
	spotify.GET("/topItems", func(ctx *gin.Context) {
		itemType := ctx.DefaultQuery("type", string(models.Tracks))
		if len(itemType) == 0 {
			itemType = string(models.Tracks)
		}
		term := ctx.DefaultQuery("term", string(models.Now))
		if len(term) == 0 {
			term = string(models.Now)
		}
		_, topTracks, error := getTopTracks(ctx, models.TopItems(itemType), models.TopItemsTerm(term))
		if error != nil {
			log.Println(error)
			ctx.Error(error)
		}
		ctx.JSON(http.StatusOK, string(topTracks))
	})
	spotify.GET("/profile", func(ctx *gin.Context) {
		_, profile, error := getProfile(ctx)
		if error != nil {
			log.Println(error)
			ctx.Error(error)
		}
		ctx.JSON(http.StatusOK, string(profile))
	})
	spotify.GET("/newReleases", func(ctx *gin.Context) {
		_, releases, error := getNewReleases(ctx)
		if error != nil {
			log.Println(error)
			ctx.Error(error)
		}
		ctx.JSON(http.StatusOK, string(releases))
	})
}

func fetchSpotifyApi(ctx *gin.Context, endpoint string, method string) (status string, aBody []byte, err error) {
	sessionToken, _ := ctx.Cookie("sessionToken") // Errors for this handled in middleware.
	if len(sessionToken) == 0 {
		errorMessage := fmt.Sprintf("Forbidden without session established. Please log in. Error=%s. ", err)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": errorMessage})
		log.Println(errorMessage)
		return http.StatusText(500), []byte(errorMessage), nil
		// ctx.Abort()
	}

	session := sessions.Default(ctx)
	redisToken := session.Get(sessionToken)
	log.Printf("redisToken: %s", redisToken)
	if redisToken == nil {
		errorMessage := fmt.Sprintf("Cookie found without Session established. ")
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

	fmt.Printf("passed in endpoint: %s \n", endpoint)
	bearer := "Bearer " + token

	client := &http.Client{}

	req, err := http.NewRequest(method, models.SpotifyUrl+endpoint, nil)
	if err != nil {
		log.Println(err)
		return "", nil, err
	}

	for name, vals := range req.Header {
		for _, value := range vals {
			fmt.Printf("Headers: name=%s, val=%s", name, value)
		}
	}

	req.Header.Add("Authorization", bearer)
	resp, err := client.Do(req)
	if err != nil {
		log.Print(err)
		return http.StatusText(500), nil, err
	}

	if resp.StatusCode != 200 {
		log.Println("Failed to fetch data from Spotify.")
		log.Println(resp.Status)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return "", nil, err
	}

	return resp.Status, body, nil
}

func getTopTracks(ctx *gin.Context, itemType models.TopItems, term models.TopItemsTerm) (string, []byte, error) {
	return fetchSpotifyApi(ctx, fmt.Sprintf("v1/me/top/%s?time_range=%s&limit=%d", string(itemType), string(term), 25), http.MethodGet)
}

func getProfile(ctx *gin.Context) (string, []byte, error) {
	return fetchSpotifyApi(ctx, "v1/me", http.MethodGet)
}

func getNewReleases(ctx *gin.Context) (string, []byte, error) {
	return fetchSpotifyApi(ctx, "v1/browse/new-releases", http.MethodGet)
}

var topSongsUSA = "37i9dQZEVXbLp5XoPON0wI"

func getTopPlayedByCountry(ctx *gin.Context) (string, []byte, error) {
	return fetchSpotifyApi(ctx, "v1/browse/playlists/%s/tracks", topSongsUSA)
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
