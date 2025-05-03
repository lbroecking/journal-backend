package models

import (
	"journal-backend/db"
	"journal-backend/logging"
)

// wird gerade nicht genutzt, da entries als []map[string]interface{} zurückgegeben werden, ohne struct
type JournalEntry struct {
	//ID              int    `json:"id"` //not really important for Frontend
	UserId          string `json:"user_id"`
	Content         string `json:"content"`
	ContentGrateful string `json:"content_grateful"`
	ContentProud    string `json:"content_proud"`
	EmotionColor    string `json:"emotion_color"`
	CreatedAt       string `json:"created_at"`
}

func FetchEntries(selectedIndex int, dbClient db.Client) ([]map[string]interface{}, error) {
	var table, selectFields string
	var result []map[string]interface{}

	switch selectedIndex {
	case 0:
		table = "journal_entries"
		selectFields = "content,content_grateful,content_proud,emotion_color,created_at"
	case 1:
		table = "moon_entries"
		selectFields = "let_go,want,created_at,moon_sign"
	case 2:
		table = "relationship_check"
		selectFields = "question,answer,created_at"
	default:
		table = "journal_entries"
		selectFields = "*"
	}

	_, err := dbClient.
		From(table).
		Select(selectFields, "", false).
		Eq("user_id", dbClient.UserID.String()).
		ExecuteTo(&result)

	if err != nil {
		logging.Log.Error("error: ", err.Error())
		return nil, err
	}
	return result, nil
}

func InsertPersonalEntry(dbClient db.Client, entry interface{}) error {
	table := "journal_entries"

	if dbClient.UserID.String() == "" {
		logging.Log.Error("no user ID available in client")
	}

	_, _, err := dbClient.
		From(table).
		Insert(entry, false, "", "*", "").
		Eq("user_id", dbClient.UserID.String()).
		Execute()

	if err != nil {
		return err
	}

	return nil
}
