package main

import (
	"journal-backend/db"
	"journal-backend/logging"
	"journal-backend/models"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var globalClient *db.Client

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		logging.Log.Fatal("Error loading .env-File")
	}

	logging.Log.Info("Connecting to API...")
	router := gin.Default()
	router.GET("/profiles", getAllUsers)
	router.POST("/register", models.RegisterHandler)
	router.POST("/login", signInWithEmailPassword)
	router.POST("/logout", logoutUser)
	router.GET("/entries", getPersonalEntries)

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Next()
	})

	router.Run(":3000")
}

func signInWithEmailPassword(c *gin.Context) {
	var req LoginRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	url := os.Getenv("SUPABASE_URL")
	apiKey := os.Getenv("SUPABASE_KEY")

	dbClient, err := db.NewClient(url, apiKey, nil)
	if err != nil {
		c.JSON(500, gin.H{"error": "Error client initializing"})
		return
	}

	session, err := dbClient.SignInWithEmailPassword(req.Email, req.Password)
	dbClient.UserID = session.User.ID.String()
	logging.Log.Info("User ID (uuid):", dbClient.UserID)
	logging.Log.Info("User ID (uuid):", session.User.ID)

	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "Login successful",
		"session": session,
	})

	globalClient = dbClient

}

func logoutUser(c *gin.Context) {
	logging.Log.Info("Received POST-Request; requested to Logout User")

	err := globalClient.Auth.Logout()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	logging.Log.Info("User logged out")

	globalClient.UserID = ""

	c.JSON(200, "user logged out")

}

func getAllUsers(c *gin.Context) {
	logging.Log.Info("Received GET-Request for user entries")

	if globalClient == nil || globalClient.UserID == "" {
		c.JSON(400, gin.H{"error": "user not logged in"})
		return
	}

	profiles, err := models.GetAllUsers(*globalClient)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	logging.Log.Info("Selecting UserId + Name of all users stored in database was successful")
	logging.Log.Info("Sended response to client.")
	c.JSON(200, profiles)
}

func getPersonalEntries(c *gin.Context) {

	logging.Log.Info("Received GET-Request for user entries")

	if globalClient == nil || globalClient.UserID == "" {
		c.JSON(400, gin.H{"error": "user not logged in"})
		return
	}

	sSelectedIndex := c.Query("selected_index")
	selectedIndex, _ := strconv.Atoi(sSelectedIndex)

	personalEntries, err := models.FetchEntries(selectedIndex, *globalClient)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, personalEntries)
}
