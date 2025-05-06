package main

import (
	"journal-backend/db"
	"journal-backend/logging"
	"journal-backend/models"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var globalClient *db.Client

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		logging.Log.Fatal("Error loading .env-File")
	}

	logging.Log.Info("Connecting to API...")
	router := gin.Default()
	router.GET("/profiles", getAllUsers)
	router.POST("/register", signUpWithEmailPassword)
	router.POST("/login", signInWithEmailPassword)
	router.POST("/logout", logoutUser)
	router.GET("/entries", getPersonalEntries)
	router.POST("/entries", newPersonalEntry)

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Next()
	})

	router.Run(os.Getenv("SERVER_URL"))
}

func signUpWithEmailPassword(c *gin.Context) {
	var req RegisterRequest

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

	user, err := dbClient.SignUpWithEmailPassword(req.Email, req.Password)

	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}

	token, _ := dbClient.SignInWithEmailPassword(req.Email, req.Password)

	newUser := models.User{
		UserId: token.User.ID.String(),
		Name:   req.Username,
	}

	err = models.NewUser(*dbClient, newUser)
	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "Login successful",
		"session": user,
	})

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
	dbClient.UserID = session.User.ID

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

	globalClient.UserID = uuid.Nil

	c.JSON(200, "user logged out")

}

func getAllUsers(c *gin.Context) {
	logging.Log.Info("Received GET-Request for user entries")

	if globalClient == nil || globalClient.UserID.String() == "" {
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

func newPersonalEntry(c *gin.Context) {
	logging.Log.Debug("Received POST-Request to insert new personal entry")

	var req models.JournalEntry

	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	//Checking if User is logged in
	if globalClient == nil || globalClient.UserID.String() == "" {
		c.JSON(400, gin.H{"error": "user not logged in"})
		logging.Log.Error("user not logged in")
		return
	}

	req.UserId = globalClient.UserID.String()

	err := models.InsertPersonalEntry(*globalClient, req)
	if err != nil {
		logging.Log.Error("Error occured while inserting user entries: ", err.Error())
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, "OK")
}

/*
TODO commenting
*/
func getPersonalEntries(c *gin.Context) {
	logging.Log.Debug("Received GET-Request for user entries")

	if globalClient == nil || globalClient.UserID.String() == "" {
		c.JSON(400, gin.H{"error": "user not logged in"})
		logging.Log.Error("user not logged in")
		return
	}

	sSelectedIndex := c.Query("selected_index")
	selectedIndex, _ := strconv.Atoi(sSelectedIndex)

	personalEntries, err := models.FetchEntries(selectedIndex, *globalClient)
	if err != nil {
		logging.Log.Debug("Error occured while fetching user entries")
		c.JSON(500, gin.H{"error": err.Error()})
		logging.Log.Error("error: ", err.Error())
		return
	}

	logging.Log.Debug("Returned user entries")

	c.JSON(200, personalEntries)
}
