package main

import (
	"log"
	"os"
	"t-murch/top-25-api/pkg/routes"

	"github.com/joho/godotenv"
)

func main() {
	if error := godotenv.Load(); error != nil {
		log.Fatal("Error loading .env file")
	}

	// If/when we need to serve static assets.
	// fs := http.FileServer(http.Dir("static/"))
	// http.Handle("/static", http.StripPrefix("/static/", fs))

	routes.Run()

	// log.Printf("SPOT_CLIENT_ID: %s", os.)
	log.Printf("SPOT_CLIENT_ID: %s", os.Getenv("SPOT_CLIENT_ID"))
}
