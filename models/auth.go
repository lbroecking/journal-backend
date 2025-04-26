package models

import (
	"bytes"
	"encoding/json"
	"io"
	"journal-backend/logging"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func LoginHandler(c *gin.Context) {
	url := os.Getenv("SUPABASE_URL")
	apiKey := os.Getenv("SUPABASE_KEY")

	var req LoginRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ungültige Daten"})
		return
	}

	payload := map[string]string{
		"email":    req.Email,
		"password": req.Password,
	}

	body, _ := json.Marshal(payload)

	supabaseReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		logging.Log.Info("Fehler beim Erstellen der Anfrage:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "interner Fehler"})
		return
	}

	supabaseReq.Header.Set("apikey", apiKey)
	supabaseReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(supabaseReq)
	if err != nil {
		logging.Log.Info("error while sending:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "interner Fehler"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		logging.Log.Info("Supabase Login fehlgeschlagen:", string(respBody))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Login fehlgeschlagen"})
		return
	}

	var authData map[string]interface{}
	json.Unmarshal(respBody, &authData)

	c.JSON(http.StatusOK, authData)
}

func RegisterHandler(c *gin.Context) {
	url := os.Getenv("SUPABASE_URL")
	apiKey := os.Getenv("SUPABASE_KEY")

	var req LoginRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ungültige Daten"})
		return
	}

	payload := map[string]string{
		"email":    req.Email,
		"password": req.Password,
	}
	body, _ := json.Marshal(payload)

	supabaseReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		logging.Log.Info("Fehler beim Erstellen der Anfrage:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "interner Fehler"})
		return
	}

	supabaseReq.Header.Set("apikey", apiKey)
	supabaseReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(supabaseReq)
	if err != nil {
		logging.Log.Info("Fehler beim Senden:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "interner Fehler"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		logging.Log.Info("registration failed:", string(respBody))
		c.JSON(http.StatusBadRequest, gin.H{"error": "registration failed"})
		return
	}

	var authData map[string]interface{}
	json.Unmarshal(respBody, &authData)

	c.JSON(http.StatusOK, authData)
}
