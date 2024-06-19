package types

import (
	"encoding/json"
	"log"
)

func ValidateJSON(jsonStr string) string {
	var js map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &js)
	if err != nil {
		log.Print(jsonStr)
		log.Fatal(err)
	}
	return jsonStr
}
