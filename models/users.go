package models

import (
	"journal-backend/db"
	"journal-backend/logging"
)

type User struct {
	//ID     int    `json:"id"`
	UserId string `json:"user_id"`
	Name   string `json:"username"`
}

type UserModify struct {
	UserId   string `json:"user_id"`
	UserName string `json:"username"`
	Picture  int    `json:"avatar_url"`
}

func GetAllUsers(dbClient db.Client) ([]map[string]interface{}, error) {
	logging.Log.Info("Received GET-Request")
	logging.Log.Info("Selecting UserId + username of all users stored in database...")

	var result []map[string]interface{}

	selectFields := "username"
	table := "profiles"

	_, err := dbClient.
		From(table).
		Select(selectFields, "", false).
		ExecuteTo(&result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func NewUser(dbClient db.Client, user interface{}) error {

	table := "profiles"

	if dbClient.UserID.String() == "" {
		logging.Log.Error("no user ID available in client")
	}

	_, _, err := dbClient.
		From(table).
		Insert(user, false, "", "*", "").
		Execute()

	if err != nil {
		return err
	}

	return nil
}
