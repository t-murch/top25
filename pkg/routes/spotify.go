package routes

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"t-murch/top-25-api/pkg/models"
	"t-murch/top-25-api/pkg/services"

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
		itemType := ctx.DefaultQuery("type", string(models.TracksType))
		if len(itemType) == 0 {
			itemType = string(models.TracksType)
		}
		term := ctx.DefaultQuery("term", string(models.Now))
		if len(term) == 0 {
			term = string(models.Now)
		}
		_, topTracks, error := services.GetTopTracks(ctx, models.TopItems(itemType), models.TopItemsTerm(term))
		if error != nil {
			log.Println(error)
			ctx.Error(error)
		}
		ctx.JSON(http.StatusOK, string(topTracks))
	})

	spotify.GET("/profile", func(ctx *gin.Context) {
		_, profile, error := services.GetCurrentProfile(ctx)
		if error != nil {
			log.Println(error)
			ctx.Error(error)
		}
		ctx.JSON(http.StatusOK, string(profile))
	})

	spotify.GET("/newReleases", func(ctx *gin.Context) {
		_, releases, error := services.GetNewReleases(ctx)
		if error != nil {
			log.Println(error)
			ctx.Error(error)
		}
		ctx.JSON(http.StatusOK, string(releases))
	})

	spotify.POST("/playlist/createPlaylist", func(ctx *gin.Context) {

		// Read Request Body
		clientData, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			log.Printf("Failed to read body of request. Error=%s \n", err)
			ctx.Error(err)
		}

		// Deserialize the body into struct
		var clientCreatePayload models.ClientCreatePayload
		err = json.Unmarshal(clientData, &clientCreatePayload)
		if err != nil {
			log.Printf("Failed to deserialize body of request. Error=%s \n", err)
			ctx.Error(err)
		}

		snapshotID := createOrUpdatePlaylist(clientCreatePayload, ctx)

		ctx.JSON(http.StatusOK, string(snapshotID))
	})
}

func createOrUpdatePlaylist(clientCreatePayload models.ClientCreatePayload, ctx *gin.Context) []byte {
	// Check for existing user in db for future usage
	_, myUser, err := models.GetSpotifyUser(clientCreatePayload.SpotifyUserID)
	if err != nil {
		log.Printf("Failed to fetch Spotify User from DB. UserID=%s. Error=%s \n", clientCreatePayload.SpotifyUserID, err)
		ctx.AbortWithError(500, err)
		return nil
	}

	/*
		If no user, we need to create it now so we can add their playlist IDs later.
		For user experience, the server will gather the details needed on its own.
	*/
	if myUser.Email == "" {
		_, profile, err := services.GetCurrentProfile(ctx)
		if err != nil {
			log.Printf("Failed to fetch User Profile from Spotify. UserID=%s. Error=%s \n", clientCreatePayload.SpotifyUserID, err)
			ctx.AbortWithError(500, err)
			return nil
		}
		var spotifyUser models.UserSpotify
		err = json.Unmarshal(profile, &spotifyUser)
		if err != nil {
			log.Printf("Failed to deserialize Spotify User Post fetching. nativeUserId=%s Error=%s \n", clientCreatePayload.SpotifyUserID, err)
			ctx.AbortWithError(500, err)
			return nil
		}

		// Update the existing User so we can use the same variable
		// Regardless of create or not.
		myUser.Email = spotifyUser.Email
		myUser.Username = spotifyUser.DisplayName
		myUser.Avatar_url = spotifyUser.Images[0].URL
		myUser.Native_id_spotify = spotifyUser.ID

		// // Finally add user to db.
		_, userSnapshot, err := models.AddUser(myUser.Email, myUser.Username, myUser.Avatar_url, myUser.Native_id_spotify)
		if err != nil || userSnapshot.Email == "" {
			log.Printf("Failed to add new user to db. email=%s. Error%s \n", myUser.Email, err)
			ctx.AbortWithError(500, err)
			return nil
		} else {
			myUser.ID = userSnapshot.ID
			log.Printf("New user successfully added to db. email=%s \n", userSnapshot.Email)
		}
	}

	var newPlaylist models.Playlist
	clientCreatePayload.Description = models.PlaylistDescription
	// We need to check for existence of a playlist
	// Since at this point we have a user from db.
	_, myPlaylistID, err := models.GetUserPlaylistID(myUser.ID, clientCreatePayload.Term)
	if err != nil {
		log.Printf("Failed to fetch playlistID from db. UserID=%s, term=%s Error=%v \n", myUser.ID, clientCreatePayload.Term, err)
		ctx.AbortWithError(500, err)
		return nil
	}

	if len(myPlaylistID) == 0 {
		// Send Spotify Create Playlist Request
		_, createData, err := services.CreatePlaylist(ctx, clientCreatePayload.Title, clientCreatePayload.SpotifyUserID, clientCreatePayload.Description)
		if err != nil {
			log.Printf("Failed to Create Spotify playlist. Error=%s\n", err)
			ctx.AbortWithError(500, err)
			return nil
		}

		// Deserialize result
		if err = json.Unmarshal(createData, &newPlaylist); err != nil {
			log.Printf("Failed to deserialize create playlist call. Error=%s \n", err)
			ctx.AbortWithError(500, err)
			return nil
		} else {
			log.Printf("Successfully created Playlist. ID=%s, Source=%s, UserID=%s", newPlaylist.ID, "spotify", myUser.ID)
		}

		// Add Playlist to db at this point with User info
		_, response, err := models.AddPlaylist(clientCreatePayload.Title, "spotify", clientCreatePayload.Description, newPlaylist.ID, models.RangeType(clientCreatePayload.Term), myUser.ID)
		if err != nil {
			log.Printf("failed to add Playlist to db for userEmail=%s. Error=%v", myUser.Email, err)
			ctx.AbortWithError(500, err)
			return nil
		}

		myPlaylistID = newPlaylist.ID
		log.Printf("response from db create playlist: %v", response)
	}

	_, snapshotID, err := services.AddItemsToPlaylist(ctx, clientCreatePayload.SpotifyUserID, myPlaylistID, clientCreatePayload.Items, 0)
	if err != nil {
		log.Println(err)
		ctx.AbortWithError(500, err)
		return nil
		// If this fails we should **DELETE THE CREATED PLAYLIST AS WELL**
		// THIS IS STILL A PART OF THE ***TRANSACTION***
	} else {
		log.Printf("New Playlist created. SnapshotID=%s \n", string(snapshotID))
	}
	return snapshotID
}
