package main

import (
	"encoding/json"
	"journal-backend/db"
	"journal-backend/helpers"
	"journal-backend/logging"
	"journal-backend/models"
	"net/http"
	"os"
	"strconv"
	"time"

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

type DeleteRequest struct {
	Table string `json:"table"`
	Id    int8   `json:"id"`
}

type InsertEntry struct {
	Table string `json:"table"`
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
	router.GET("/entries", getEntries)
	router.POST("/entries", newEntry)
	router.PUT("/entries", updateEntry)
	router.DELETE("/delete", deleteEntry)

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

	if !checkUserAuth() {
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

func newEntry(c *gin.Context) {
	logging.Log.Debug("Received POST-Request to insert new personal entry")

	var raw map[string]interface{}
	if err := c.BindJSON(&raw); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	table, ok := raw["table"].(string)
	if !ok || table == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid 'table' key"})
		return
	}

	if !checkUserAuth() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not logged in"})
		return
	}

	var err error
	var entry map[string]any

	userID := globalClient.UserID.String()
	createdAt := time.Now().Format("2006-01-02")

	switch table {
	case "journal_entries":
		var persEntry models.PersonalEntry

		jsonBytes, _ := json.Marshal(raw)
		if err := json.Unmarshal(jsonBytes, &persEntry); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid journal entry structure"})
			return
		}

		persEntry.UserId = userID
		persEntry.CreatedAt = createdAt
		entry = helpers.ToMap(persEntry)

	case "moon_entries":
		var moonEntry models.MoonEntry

		jsonBytes, _ := json.Marshal(raw)
		if err := json.Unmarshal(jsonBytes, &moonEntry); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid moon entry structure"})
			return
		}
		moonEntry.UserId = userID
		moonEntry.CreatedAt = createdAt
		entry = helpers.ToMap(moonEntry)

	case "relationship_check":
		var relEntry models.RelationshipCheckEntry

		jsonBytes, _ := json.Marshal(raw)
		if err := json.Unmarshal(jsonBytes, &relEntry); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid relationship check structure"})
			return
		}
		relEntry.UserId = userID
		relEntry.CreatedAt = createdAt
		entry = helpers.ToMap(relEntry)

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown table"})
		return
	}

	logging.Log.Infof("Inserting entry into table '%s': %+v", table, entry)

	if err = models.InsertEntry(*globalClient, entry, table); err != nil {
		logging.Log.Errorf("Error occurred while inserting user entry: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert entry"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func checkUserAuth() bool {
	if globalClient == nil || globalClient.UserID.String() == "" {
		logging.Log.Error("user not logged in")
		return false
	}

	return true
}

func updateEntry(c *gin.Context) {
	logging.Log.Debug("Received POST-Request to insert new personal entry")

	var raw map[string]interface{}
	if err := c.BindJSON(&raw); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	table, ok := raw["table"].(string)
	if !ok || table == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid 'table' key"})
		return
	}

	if !checkUserAuth() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not logged in"})
		return
	}

	var err error
	var entry map[string]any
	var entryId int

	userID := globalClient.UserID.String()

	switch table {
	case "journal_entries":
		var persEntry models.PersonalEntry

		jsonBytes, _ := json.Marshal(raw)
		if err := json.Unmarshal(jsonBytes, &persEntry); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid journal entry structure"})
			return
		}

		persEntry.UserId = userID
		entryId = persEntry.EntryID

		entry = helpers.ToMap(persEntry)

	case "moon_entries":
		var moonEntry models.MoonEntry

		jsonBytes, _ := json.Marshal(raw)
		if err := json.Unmarshal(jsonBytes, &moonEntry); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid moon entry structure"})
			return
		}
		moonEntry.UserId = userID
		entryId = moonEntry.EntryID

		entry = helpers.ToMap(moonEntry)

	case "relationship_check":
		var relEntry models.RelationshipCheckEntry

		jsonBytes, _ := json.Marshal(raw)
		if err := json.Unmarshal(jsonBytes, &relEntry); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid relationship check structure"})
			return
		}
		relEntry.UserId = userID
		entryId = relEntry.EntryID

		entry = helpers.ToMap(relEntry)

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown table"})
		return
	}

	logging.Log.Infof("Inserting entry into table '%s': %+v", table, entry)

	if err = models.UpdateEntry(*globalClient, entry, table, entryId); err != nil {
		logging.Log.Errorf("Error occurred while inserting user entry: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert entry"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func deleteEntry(c *gin.Context) {
	logging.Log.Debug("Received DELETE-Request to delete one entry")

	var req DeleteRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	//Checking if User is logged in
	if !checkUserAuth() {
		c.JSON(400, gin.H{"error": "user not logged in"})
		return
	}

	logging.Log.Debug("Delete from ", req.Table, " where id= ", req.Id)

	err := models.DeleteEntry(*globalClient, req.Table, req.Id)
	if err != nil {
		logging.Log.Error("Error occured while deleting entry: ", err.Error())
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, "OK")
}

/*
TODO commenting
*/
func getEntries(c *gin.Context) {
	logging.Log.Debug("Received GET-Request for user entries")

	if !checkUserAuth() {
		c.JSON(400, gin.H{"error": "user not logged in"})
		return
	}

	sSelectedIndex := c.Query("selected_index")
	selectedIndex, _ := strconv.Atoi(sSelectedIndex)

	entries, err := models.FetchEntries(selectedIndex, *globalClient)
	if err != nil {
		logging.Log.Debug("Error occured while fetching user entries")
		c.JSON(500, gin.H{"error": err.Error()})
		logging.Log.Error("error: ", err.Error())
		return
	}

	logging.Log.Debug("Returned user entries")

	c.JSON(200, entries)
}
