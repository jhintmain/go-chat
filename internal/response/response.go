package response

import (
	"encoding/json"
	"net/http"
)

func Success(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK
	json.NewEncoder(w).Encode(data)
}

func Error(w http.ResponseWriter, httpCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	data := map[string]string{
		"status":  "error",
		"message": message,
	}
	json.NewEncoder(w).Encode(data)
}
