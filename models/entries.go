package models

import (
	"journal-backend/db"
	"journal-backend/logging"
)

type JournalEntry struct {
	ID              int    `json:"id"`
	Content         string `json:"content"`
	ContentGrateful string `json:"content_grateful"`
	ContentProud    string `json:"content_proud"`
	EmotionColor    string `json:"emotion_color"`
	CreatedAt       string `json:"created_at"`
	Profiles        struct {
		ID int `json:"id"`
		//UserID string `json:"user_id"`
		Username string `json:"username"`
	} `json:"profiles"`
}

func PersonalEntries(dbClient db.Client) ([]JournalEntry, error) {
	//personalEntries aus DB lesen || GET
	logging.Log.Info("Selecting entry-details from database...")
	var entries []JournalEntry

	logging.Log.Info(dbClient.UserID)
	logging.Log.Info(entries)

	selectFields := "id,content,content_grateful,content_proud,emotion_color,created_at,profiles(username)"
	_, err := dbClient.From("journal_entries").
		Select(selectFields, "", false).
		Eq("user_id", dbClient.UserID).
		ExecuteTo(&entries)

	logging.Log.Info(entries)

	if err != nil {
		return nil, err
	}

	logging.Log.Info(entries)
	return entries, nil

}

func FetchEntries(selectedIndex int, userID string, dbClient *db.Client) ([]map[string]interface{}, error) {
	var table, selectFields string

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

	var result []map[string]interface{}
	_, err := dbClient.
		From(table).
		Select(selectFields, "", false).
		Eq("user_id", userID).
		ExecuteTo(&result)

	if err != nil {
		return nil, err
	}
	return result, nil
}
