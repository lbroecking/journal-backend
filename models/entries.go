package models

import (
	"journal-backend/db"
)

// wird gerade nicht genutzt, da entries als []map[string]interface{} zurückgegeben werden, ohne struct
type JournalEntry struct {
	//ID              int    `json:"id"` //not really important for Frontend
	Content         string `json:"content"`
	ContentGrateful string `json:"content_grateful"`
	ContentProud    string `json:"content_proud"`
	EmotionColor    string `json:"emotion_color"`
	CreatedAt       string `json:"created_at"`
	Profiles        struct {
		ID int `json:"id"` //important for Frontend?
		//UserID string `json:"user_id"`
		Username string `json:"username"`
	} `json:"profiles"`
}

func FetchEntries(selectedIndex int, dbClient db.Client) ([]map[string]interface{}, error) {
	var table, selectFields string
	var result []map[string]interface{}

	switch selectedIndex {
	case 0:
		table = "journal_entries"
		selectFields = "id,content,content_grateful,content_proud,emotion_color,created_at,profiles(username)"
	case 1:
		table = "moon_entries"
		selectFields = "id,let_go,want,created_at,moon_sign,profiles(username)"
	case 2:
		table = "relationship_check"
		selectFields = "id,question,answer,created_at,profiles(username)"
	default:
		table = "journal_entries"
		selectFields = "*"
	}

	_, err := dbClient.
		From(table).
		Select(selectFields, "", false).
		Eq("user_id", dbClient.UserID).
		ExecuteTo(&result)

	if err != nil {
		return nil, err
	}
	return result, nil
}
