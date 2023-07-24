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

func addSpotifyRoutes(rg *gin.RouterGroup) {
	spotify := rg.Group("/spotify")

	spotify.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, "Spotify Base Route")
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
	/**
	1. Incoming request
	2. redirect to Spotc with a redirect param back to my BE
	3. Spotify returns token
	4. redirect to homepage with token
	**/
	spotify.GET("/login", func(ctx *gin.Context) {
		state, _ := GenerateRandomString(32)
		ctx.SetCookie("stateKey", state, 60*60*2, "/top_25", "localhost", false, true)
		ctx.AddParam("redirect_uri", os.Getenv("SPOTIFY_REDIRECT"))
		ctx.AddParam("show_dialog", "true")
		ctx.AddParam("state", state)
		ctx.AddParam("scope", models.SCOPES)
		ctx.Redirect(http.StatusFound, models.LOGIN_REDIRECT_URL)
	})

	spotify.GET("/callback", func(ctx *gin.Context) {
		ctx.AddParam("code", ctx.Query("code"))
		ctx.AddParam("redirect_uri", os.Getenv("SPOTIIFY_REDIRECT"))
		ctx.AddParam("grant_type", "authorization_code")

		state := ctx.Query("state")
		storedState, error := ctx.Cookie("stateKey")
		if error != nil || storedState != state {
			ctx.AddParam("error", "state_mismatch")
			log.Fatalf("Appears to have interference between /login & /callback. Error: %v", error)
			ctx.Redirect(http.StatusBadRequest, "http://localhost:8080")
		}

		error = getAccessToken(ctx)
		if error != nil {
			ctx.AddParam("error", "Failed to gain Access Token")
			log.Fatalln("Failed to gain Access Token")
			ctx.Redirect(http.StatusBadRequest, "http://localhost:8080")
		}
		ctx.Redirect(http.StatusAccepted, "http://localhost:8080")
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

func fetchSpotifyApi(ctx *gin.Context, endpoint string, method string) (status string, aBody []byte, err error) {
	token, _ := ctx.Cookie("sessionToken") // Errors for this handled in middleware.

	fmt.Printf("passed in endpoint: %s \n", endpoint)
	// fmt.Printf("updated token: %s \n", token)
	bearer := "Bearer " + token

	client := &http.Client{}

	req, error := http.NewRequest(method, models.SpotifyUrl+endpoint, nil)
	if error != nil {
		log.Println(error)
		return "", nil, error
	}

	req.Header.Add("Authorization", bearer)
	resp, error := client.Do(req)
	if error != nil {
		log.Print(error)
		return "", nil, error
	}

	defer resp.Body.Close()
	body, error := io.ReadAll(resp.Body)
	if error != nil {
		log.Println(error)
		return "", nil, error
	}

	return resp.Status, body, nil
}

func getAccessToken(ctx *gin.Context) error {
	fmt.Println("Getting new Access Token")

	session := sessions.Default(ctx)
	sessionToken := uuid.NewString()

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", os.Getenv("SPOT_CLIENT_ID"))
	data.Set("client_secret", os.Getenv("SPOT_CLIENT_SECRET"))

	client := &http.Client{}
	req, error := http.NewRequest(http.MethodPost, "https://accounts.spotify.com/api/token", bytes.NewBuffer([]byte(data.Encode())))
	if error != nil {
		log.Printf("Error building Request Struct for Spotify. Error: %s \n", error)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, error := client.Do(req)
	if error != nil {
		log.Printf("Error getting new Access Token from Spotify. Error: %s \n", error)
	}
	defer resp.Body.Close()

	body, error := io.ReadAll(resp.Body)
	if error != nil {
		log.Printf("Error getting reading body of Response from Acccess Token. Error: %s \n", error)
	}

	var tokenResponse models.SpotifyTokenResponse
	error = json.Unmarshal(body, &tokenResponse)
	if error != nil {
		log.Printf("Error parsing response from getting new Spot Token. Error: %s \n", error)
	}

	ctx.SetCookie("sessionToken", sessionToken, 60*60*2, "/top_25", "localhost", false, true)
	session.Set(sessionToken, tokenResponse)
	log.Printf("Sucessfully set Cookie and established session with sessionToken: %s", sessionToken)
	// fmt.Println("tokenResponse: " + tokenResponse.AccessToken + "\n")
	return error
}

// /**
//  * After gaining user permission, we
//  * ping /token to gain Access and Refresh Tokens.
//  */
// server.get('/callback', async (request: UserRequest, reply: FastifyReply): Promise<void> => {
//   if (client_id === undefined || redirect_uri === undefined) {
//     server.log.error('You forgot to add your client credentials!! ');
//     reply.redirect(FE_REDIRECT);
//   } else {
//     const requestDataProperty = new URLSearchParams({ 'code': request.query.code, 'redirect_uri': redirect_uri, 'grant_type': 'authorization_code' });
//     const state: string = request.query.state;
//     const storedState = request.cookies ? request.cookies['stateKey'] : null;

//     /**
//      * If state mismatch, redirect back
//      * Otherwise, create the auth object needed
//      * for Access & Refresh Tokens
//      */
//     if (state === null || state !== storedState) {
//       reply.redirect(FE_REDIRECT + '#' + new URLSearchParams({ 'error': 'state_mismatch' }));
//     } else {
//       reply.clearCookie('stateKey');

//       try {
//         const { data, status, statusText } = await getTokens(axiosInstance, requestDataProperty);

//         if (statusText === 'OK') {
//           const tokens = new URLSearchParams({
//             'access_token': data?.access_token,
//             'refresh_token': data?.refresh_token,
//           });
//           reply.redirect(FE_REDIRECT + '#' + tokens);
//         } else {
//           server.log.error(new Error('Server related error retrieving login tokens.' + { cause: data }));
//           reply.redirect(FE_REDIRECT);
//         }
//       } catch (error: any) {
//         server.log.error('Failed to login and retrieve tokens. Error: %o' + error);
//         reply.redirect(FE_REDIRECT);
//       }
//     }
//   }
// });
// };

func getTopTracks(ctx *gin.Context, itemType models.TopItems, term models.TopItemsTerm) (string, []byte, error) {
	return fetchSpotifyApi(ctx, fmt.Sprintf("v1/me/top/%s?time_range=%s&limit=%d", string(itemType), string(term), 5), http.MethodGet)
}

func getProfile(ctx *gin.Context) (string, []byte, error) {
	return fetchSpotifyApi(ctx, "v1/me", http.MethodGet)
}

func getNewReleases(ctx *gin.Context) (string, []byte, error) {
	return fetchSpotifyApi(ctx, "v1/browse/new-releases", http.MethodGet)
}
