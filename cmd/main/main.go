package main

import (
	"log"
	"net/http"
	"os"
	"t-murch/top-25-api/pkg/routes"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func CookieTool() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get cookie
		if cookie, err := ctx.Cookie("sessionToken"); err == nil {
			if cookie == "ok" {
				ctx.Next()
				return
			}
		}

		// Cookie verification failed
		// c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden with no cookie"})
		ctx.Redirect(http.StatusForbidden, "/v1/spotify/login")
		// ctx.Abort()
	}
}

func main() {
	if error := godotenv.Load(); error != nil {
		log.Fatal("Error loading .env file")
	}
	router := gin.Default()

	// route.GET("/login", func(c *gin.Context) {
	// 	// Set cookie {"label": "ok" }, maxAge 30 seconds.
	// 	c.SetCookie("label", "ok", 30, "/", "localhost", false, true)
	// 	c.String(200, "Login success!")
	// })

	// route.GET("/home", CookieTool(), func(c *gin.Context) {
	// 	c.JSON(200, gin.H{"data": "Your home page"})
	// })

	// ***** FIX ME LATER, I AM NOOT VERY SECURE *****************
	store := cookie.NewStore([]byte("laneelise2512"))
	router.Use(sessions.Sessions("top_25_session", store))

	// r.GET("/incr", func(c *gin.Context) {
	//   session := sessions.Default(c)
	//   var count int
	//   v := session.Get("count")
	//   if v == nil {
	//     count = 0
	//   } else {
	//     count = v.(int)
	//     count++
	//   }
	//   session.Set("count", count)
	//   session.Save()
	//   c.JSON(200, gin.H{"count": count})
	// })

	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, "Welcome to the Top 25 API")
	})

	// If/when we need to serve static assets.
	// fs := http.FileServer(http.Dir("static/"))
	// http.Handle("/static", http.StripPrefix("/static/", fs))

	routes.Run()

	// log.Printf("SPOT_CLIENT_ID: %s", os.)
	log.Printf("SPOT_CLIENT_ID: %s", os.Getenv("SPOT_CLIENT_ID"))
}
