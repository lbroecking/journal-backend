package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	supabaseURL     = "https://bopabjxclatablmbnwia.supabase.co"
	supabaseAnonKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImJvcGFianhjbGF0YWJsbWJud2lhIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NDMxNjk1ODksImV4cCI6MjA1ODc0NTU4OX0.Idkx_4ehN72Y34NtMv0BUR9ZP3vYOekLd46LgRWGwoA"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func loginHandler(c *gin.Context) {
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

	url := fmt.Sprintf("%s/auth/v1/token?grant_type=password", supabaseURL)
	supabaseReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		log.Println("Fehler beim Erstellen der Anfrage:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "interner Fehler"})
		return
	}

	supabaseReq.Header.Set("apikey", supabaseAnonKey)
	supabaseReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(supabaseReq)
	if err != nil {
		log.Println("Fehler beim Senden:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "interner Fehler"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		log.Println("Supabase Login fehlgeschlagen:", string(respBody))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Login fehlgeschlagen"})
		return
	}

	var authData map[string]interface{}
	json.Unmarshal(respBody, &authData)

	c.JSON(http.StatusOK, authData)
}

func registerHandler(c *gin.Context) {
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

	url := fmt.Sprintf("%s/auth/v1/signup", supabaseURL)
	supabaseReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		log.Println("Fehler beim Erstellen der Anfrage:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "interner Fehler"})
		return
	}

	supabaseReq.Header.Set("apikey", supabaseAnonKey)
	supabaseReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(supabaseReq)
	if err != nil {
		log.Println("Fehler beim Senden:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "interner Fehler"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		log.Println("Registrierung fehlgeschlagen:", string(respBody))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Registrierung fehlgeschlagen"})
		return
	}

	var authData map[string]interface{}
	json.Unmarshal(respBody, &authData)

	c.JSON(http.StatusOK, authData)
}
