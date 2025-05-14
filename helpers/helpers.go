package helpers

import (
	"encoding/json"
	"fmt"
	"journal-backend/db"
	"journal-backend/logging"
	"time"
)

func ToMap(v interface{}) map[string]interface{} {
	bytes, err := json.Marshal(v)
	if err != nil {
		logging.Log.Errorf("Failed to marshal struct: %v", err)
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bytes, &result); err != nil {
		logging.Log.Errorf("Failed to unmarshal into map: %v", err)
		return nil
	}

	return result
}

func FilterEmptyFields(input map[string]interface{}) map[string]interface{} {
	clean := make(map[string]interface{})
	for k, v := range input {
		switch val := v.(type) {
		case string:
			if val != "" {
				clean[k] = val
			}
		case int, int32, int64, float32, float64:
			if fmt.Sprintf("%v", val) != "0" {
				clean[k] = val
			}
		case bool:
			if val {
				clean[k] = val
			}
		case []interface{}:
			if len(val) > 0 {
				clean[k] = val
			}
		case map[string]interface{}:
			if len(val) > 0 {
				clean[k] = val
			}
		default:
			if val != nil {
				clean[k] = val
			}
		}
	}
	return clean
}

func ClearOldLetGoEntries(dbClient db.Client) error {
	// Berechne den Zeitpunkt, der 24 Stunden in der Vergangenheit liegt
	filterTime := time.Now().Add(-24 * time.Hour).Format("2006-01-02 15:04:05")

	// Führe das UPDATE aus: Setze `let_go` auf NULL für Einträge, die älter als 24 Stunden sind
	_, _, err := dbClient.
		From("moon_entries").
		Update(map[string]interface{}{"let_go": nil}, "", "").
		Gt("created_at", "1970-01-01"). // optional, um sicherzustellen, dass `created_at` gültig ist
		Lt("created_at", filterTime).   // Einträge älter als 24 Stunden
		Not("let_go", "is", "NULL").    // nur Einträge mit nicht-NULL `let_go`
		Execute()

	return err
}
