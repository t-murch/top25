package routes

import (
	"fmt"
	"log"
	"net/http"

	"t-murch/top-25-api/pkg/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
)

var router = gin.Default()

func Run() {

	/**
	Database for nows
	*/
	// models.SupabaseClient.InitializeSupabaseClient()
	models.InitializeSupabaseClient()

	// ***** FIX ME LATER, I AM NOOT VERY SECURE *****************
	// store := cookie.NewStore([]byte("laneelise2512"))
	// // store.MaxLength(8192)

	// store, err := redis.NewStore(10, "tcp", "redis:6379", "", []byte("laneelise2512"))
	store, err := redis.NewStore(10, "tcp", "localhost:6379", "", []byte("laneelise2512"))
	if err != nil {
		log.Fatal("Error creating Redis store: ", err)
	}

	store.Options(sessions.Options{
		MaxAge:   60 * 60, // * 2,
		Path:     "/",
		Domain:   "10.0.0.5",
		Secure:   false,
		HttpOnly: true,
	})

	router.Use(sessions.Sessions("top_25_session", store))

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowCredentials = true
	router.Use(cors.New(config))

	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, "Welcome to the Top 25 API")
	})

	getRoutes()
	router.Run(":8080")
}

func getRoutes() {
	v1 := router.Group("/v1")
	v1.GET("/status", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, "Top 25 API up and running. ")
	})
	addUserRoutes(v1)
	addSpotifyRoutes(v1)
}

func CookieTool() gin.HandlerFunc {
	fmt.Println("todd -  in CookieTool")
	return func(ctx *gin.Context) {
		cookie, err := ctx.Cookie("sessionToken")
		// Get cookie
		if err == nil {
			if len(cookie) > 0 {
				ctx.Next()
				return
			}
		}

		// Cookie verification failed
		errorMessage := fmt.Sprintf("Forbidden without session established. Please log in. Error=%s. ", err)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": errorMessage})
		log.Println(errorMessage)
		ctx.Abort()
	}
}
