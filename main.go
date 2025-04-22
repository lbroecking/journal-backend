package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Profile struct {
	ID int `json:"id"`
	//UserID string `json:"user_id"`
	Name string `json:"username"`
}

func main() {
	connStr := "host=127.0.0.1 port=54322 user=postgres password=postgres dbname=postgres sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	router := gin.Default()
	router.POST("/login", loginHandler)
	router.POST("/register", registerHandler)

	router.GET("/profiles", func(c *gin.Context) {
		rows, err := db.Query("SELECT id, username FROM profiles")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var profiles []Profile
		for rows.Next() {
			var p Profile
			if err := rows.Scan(&p.ID, &p.Name); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			profiles = append(profiles, p)
		}
		c.JSON(http.StatusOK, profiles)
	})

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Next()
	})

	router.Run(":3000")
}
