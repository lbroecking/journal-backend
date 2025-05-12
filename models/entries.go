package models

import (
	"encoding/json"
	"journal-backend/db"
	"journal-backend/helpers"
	"journal-backend/logging"
	"strconv"

	"github.com/supabase-community/postgrest-go"
)

// wird gerade nicht genutzt, da entries als []map[string]interface{} zurückgegeben werden, ohne struct
type PersonalEntry struct {
	EntryID         int    `json:"id,omitempty"`
	UserId          string `json:"user_id"`
	Content         string `json:"content"`
	ContentGrateful string `json:"content_grateful"`
	ContentProud    string `json:"content_proud"`
	EmotionColor    string `json:"emotion_color"`
	CreatedAt       string `json:"created_at"`
}

type MoonEntry struct {
	EntryID   int             `json:"id,omitempty"`
	UserId    string          `json:"user_id"`
	LetGo     json.RawMessage `json:"let_go"`
	Want      json.RawMessage `json:"want"`
	MoonSign  string          `json:"moon_sign"`
	CreatedAt string          `json:"created_at"`
}

type RelationshipCheckEntry struct {
	EntryID   int    `json:"id,omitempty"`
	UserId    string `json:"user_id"`
	Question  string `json:"question"`
	Answer    string `json:"answer"`
	CreatedAt string `json:"created_at"`
}

func FetchEntries(selectedIndex int, dbClient db.Client) ([]map[string]interface{}, error) {
	var table, selectFields string
	var result []map[string]interface{}

	switch selectedIndex {
	case 0:
		table = "journal_entries"
		selectFields = "id, content,content_grateful,content_proud,emotion_color,created_at"
	case 1:
		table = "moon_entries"
		selectFields = "id, let_go,want,created_at,moon_sign"
	case 2:
		table = "relationship_check"
		selectFields = "id, question,answer,created_at"
	default:
		table = "journal_entries"
		selectFields = "*"
	}

	_, err := dbClient.
		From(table).
		Select(selectFields, "", false).
		Eq("user_id", dbClient.UserID.String()).
		Order("created_at", &postgrest.OrderOpts{Ascending: false}).
		ExecuteTo(&result)

	if err != nil {
		logging.Log.Error("error: ", err.Error())
		return nil, err
	}
	return result, nil
}

func InsertEntry(dbClient db.Client, entry interface{}, table string) error {

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

func UpdateEntry(dbClient db.Client, entry map[string]interface{}, table string, entryId int) error {

	sID := strconv.FormatInt(int64(entryId), 10)
	logging.Log.Debug("Update entry in ", table, " where id= ", entryId)

	filtered := helpers.FilterEmptyFields(entry)

	_, _, err := dbClient.
		From(table).
		Update(filtered, "", "").
		Eq("id", sID).
		Eq("user_id", dbClient.UserID.String()).
		Execute()

	if err != nil {
		return err
	}

	return nil
}

func DeleteEntry(dbClient db.Client, table string, entryId int8) error {

	sID := strconv.FormatInt(int64(entryId), 10)
	logging.Log.Debug("Delete from ", table, " where id= ", entryId)

	_, _, err := dbClient.
		From(table).
		Delete("", "exact").
		Eq("id", sID).
		Execute()

	if err != nil {
		return err
	}

	return nil
}
