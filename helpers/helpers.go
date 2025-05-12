package helpers

import (
	"encoding/json"
	"fmt"
	"journal-backend/logging"
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
