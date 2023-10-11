package routes

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"t-murch/top-25-api/pkg/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func addUserRoutes(rg *gin.RouterGroup) {
	client := rg.Group("/user")

	client.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, "User Base Route")
	})
	/**
	1. Incoming request
	2. redirect to Spotc with a redirect param back to my BE
	3. Spotify returns token
	4. redirect to homepage with token
	**/
	client.GET("/spotify/login", func(ctx *gin.Context) {
		request := map[string]interface{}{
			"url":    models.LOGIN_REDIRECT_URL,
			"params": nil,
		}

		queryParams := url.Values{}
		queryParams.Set("client_id", os.Getenv("SPOT_CLIENT_ID"))
		queryParams.Set("response_type", "code")
		queryParams.Set("scope", models.SCOPES)
		queryParams.Set("show_dialog", "true")

		var state string
		stateCookie, _ := ctx.Cookie("stateKey")
		if len(stateCookie) > 0 {
			fmt.Println("User already has established stateKey. Returning existing stateKey")
			// ctx.JSON(http.StatusAccepted, sessionCookie)
			queryParams.Set("state", stateCookie)
			request["params"] = queryParams
			ctx.JSON(http.StatusOK, request)
		}

		fmt.Printf("todd - Cookie incoming: %s \n", stateCookie)

		state, _ = GenerateRandomString(32)
		queryParams.Set("state", state)
		ctx.SetCookie("stateKey", state, 60*60*2, "/", "localhost", false, true)

		request["params"] = queryParams
		ctx.JSON(http.StatusOK, request)
	})

	client.GET("/spotify/callback", func(ctx *gin.Context) {
		// session := sessions.Default(ctx)
		// sessionCookie, err := ctx.Cookie("sessionToken")
		// if err != nil {
		// 	sessionToken := session.Get(sessionCookie)
		// 	if sessionToken != nil {
		// 		ctx.JSON(http.StatusAccepted, sessionCookie)
		// 	}
		// }
		// fmt.Println("We shouldn't hit this line multiple times. ")

		ctx.AddParam("client_id", os.Getenv("SPOT_CLIENT_ID"))
		ctx.AddParam("client_secret", os.Getenv("SPOT_CLIENT_SECRET"))
		ctx.AddParam("response_type", "code")

		requestState := ctx.Query("state")

		code := ctx.Query("code")
		redirectUri := ctx.Query("redirectUri")
		if len(requestState) == 0 {
			ctx.AddParam("error", "state_mismatch")
			log.Fatalf("Appears to have interference between /login & /callback. Error: mismatch in Header values")
			ctx.Redirect(http.StatusBadRequest, ctx.Request.Referer())
		}

		authRequestInfo := buildAccessTokenRequest(code, redirectUri, requestState)

		sessionToken := getAccessToken(ctx, authRequestInfo)
		if len(sessionToken) == 0 {
			ctx.AddParam("error", "Failed to gain Access Token")
			log.Fatalln("Failed to gain Access Token")
			ctx.Redirect(http.StatusBadRequest, ctx.Request.Referer())
		}
		ctx.JSON(http.StatusAccepted, sessionToken)
	})
}

func GenerateRandomBytes(num int) ([]byte, error) {
	bytes := make([]byte, num)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func GenerateRandomString(num int) (string, error) {
	bytes, err := GenerateRandomBytes(num)
	return base64.URLEncoding.EncodeToString(bytes), err
}

type TokenRequest struct {
	Body string
	Url  string
}

func buildAccessTokenRequest(code string, redirectUri string, state string) TokenRequest {
	data := url.Values{}
	data.Set("client_id", os.Getenv("SPOT_CLIENT_ID"))
	data.Set("client_secret", os.Getenv("SPOT_CLIENT_SECRET"))
	data.Set("code", code)
	data.Set("redirect_uri", redirectUri)
	data.Set("grant_type", "authorization_code")
	data.Set("state", state)

	tokenUrl := "https://accounts.spotify.com/api/token"

	request := TokenRequest{
		Body: data.Encode(),
		Url:  tokenUrl,
	}

	return request
}

func getAccessToken(ctx *gin.Context, authRequestInfo TokenRequest) string {
	fmt.Println("Getting new Access Token")
	// fmt.Printf("ctx= %s\n", ctx.Request.Cookies())

	session := sessions.Default(ctx)
	sessionToken := uuid.NewString()

	log.Printf("authRequest from client=%s \n", authRequestInfo.Body)

	client := &http.Client{}
	req, error := http.NewRequest(http.MethodPost, "https://accounts.spotify.com/api/token", bytes.NewBuffer([]byte(authRequestInfo.Body)))
	if error != nil {
		log.Printf("Error building Request Struct for Spotify. Error: %s \n", error)
		sessionToken = ""
		return sessionToken
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, error := client.Do(req)
	if error != nil {
		log.Printf("Error getting new Access Token from Spotify. Error: %s \n", error)
		sessionToken = ""
		return sessionToken
	}
	defer resp.Body.Close()

	body, error := io.ReadAll(resp.Body)
	if error != nil || resp.StatusCode != 200 {
		log.Printf("Error getting reading body of Response from Acccess Token. Error: %s \n", error)
		log.Printf("Body without Access Token: %s \n", body)
		sessionToken = ""
		return sessionToken
	}

	var tokenResponse models.SpotifyTokenResponse
	error = json.Unmarshal(body, &tokenResponse)
	if error != nil {
		log.Printf("Error parsing response from getting new Spot Token. Error: %s \n", error)
		sessionToken = ""
		return sessionToken
	}

	tokenJSON, _ := json.Marshal(tokenResponse.AccessToken)
	ctx.SetCookie("sessionToken", sessionToken, 60*60, "/", "10.0.0.5", false, true)
	session.Set(sessionToken, tokenJSON)
	err := session.Save()
	if err != nil {
		log.Printf("Failed to save our session. Error=%s\n", err)
		log.Printf("Failed tokenJSON Value=%s\n", tokenJSON)
		log.Printf("Failed sessionToken Value=%s\n", sessionToken)
	}

	return sessionToken
}
