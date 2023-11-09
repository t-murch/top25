package models

import (
	"log"
	"os"

	"github.com/nedpals/supabase-go"
)

var supabaseClient *supabase.Client

const (
	userTable          = "users"
	playlistTable      = "playlists"
	userXplaylistTable = "users_x_playlists"
)

func InitializeSupabaseClient() {
	supabaseURL := os.Getenv("SUPA_API_URL")
	supabaseKey := os.Getenv("SUPA_API_KEY")

	// Create the Supabase client
	client := supabase.CreateClient(supabaseURL, supabaseKey)
	// if err != nil {
	// 	// Handle the error
	// 	fmt.Println("cannot initalize client", err)
	// 	panic(err)
	// }

	// _, count, err := client.From("users").Select("*", "exact", false).Execute()
	var result []UserDAO
	err := client.DB.From("users").Select("*").Execute(&result)
	if err != nil {
		log.Println("failed to query users table", err)
	}
	log.Printf("Lets how many users we have=%v \n", len(result))

	supabaseClient = client
}

type User struct {
	Email             string `json:"email"`
	Username          string `json:"username"`
	Avatar_url        string `json:"avatar_url"`
	Native_id_spotify string `json:"native_id_spotify"`
}

type UserDAO struct {
	Avatar_url        string `json:"avatar_url"`
	Email             string `json:"email"`
	ID                string `json:"id"`
	Native_id_spotify string `json:"native_id_spotify"`
	Native_id_apple   string `json:"native_id_apple"`
	UpdatedAt         string `json:"updated_at"`
	Username          string `json:"username"`
}

func AddUser(email string, username string, avatar_url string, nativeID string) (status string, user UserDAO, err error) {

	tableName := "users"

	userToInsert := User{
		Email:             email,
		Username:          username,
		Avatar_url:        avatar_url,
		Native_id_spotify: nativeID,
	}

	var myUser [1]UserDAO
	err = supabaseClient.DB.From(tableName).Insert(userToInsert).Execute(&myUser)
	if err != nil {
		// panic(err)
		log.Printf("Failed to insert new User. email=%s, Error=%s \n", email, err)
		return "fail", UserDAO{}, err
	}

	log.Printf("users post-INSERT. RESPONSE=%s \n", string(myUser[0].Email))

	return "success", myUser[0], nil
}

func GetSpotifyUser(nativeID string) (status string, user UserDAO, error error) {
	tableName := "users"

	var dbUsers [1]UserDAO
	err := supabaseClient.DB.From(tableName).Select("*").Limit(1).Execute(&dbUsers)
	if err != nil {
		log.Printf("Failed to query for Spotify UserID=%s. Error=%s", nativeID, err)
		return "fail", UserDAO{}, err
	}

	return "success", dbUsers[0], nil
}

const (
	ShortTerm  RangeType = "short_term"
	MediumTerm RangeType = "medium_term"
	LongTerm   RangeType = "long_term"
)

type RangeType string

type PlaylistSpotify struct {
	SourceID    string    `json:"source_id"`
	Source      string    `json:"source"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Term        RangeType `json:"term"`
}

func GetUserPlaylistID(userID string, term string) (status string, playlistID string, error error) {
	tableName := userXplaylistTable

	var userRow [1]map[string]interface{}
	err := supabaseClient.DB.From(tableName).Select("playlist_source_id").Eq("user_id", userID).Eq("playlist_term", term).Execute(&userRow)
	if err != nil {
		log.Printf("Failed query for db playlist. userID=%s, Error=%s", userID, err)
		return "fail", "", err
	}

	if userRow[0] != nil {
		log.Printf("userRow = %s \n", userRow[0]["playlist_source_id"])
		playlistID = userRow[0]["playlist_source_id"].(string)
	}

	return "success", playlistID, nil
}

func AddPlaylist(title string, source string, description string, sourceID string, term RangeType, userID string) (status string, playlistID int, error error) {

	paramMap := map[string]interface{}{
		"description": description,
		"source":      source,
		"source_id":   sourceID,
		"term":        term,
		"title":       title,
		"user_id":     userID,
	}

	var result int
	err := supabaseClient.DB.Rpc("insert_playlist", paramMap).Execute(&result)
	if err != nil {
		log.Printf("Failed insert new playlist. userID=%s, Error=%s \n", userID, err)
		return "fail", 0, err
	}

	return "success", result, nil
}
