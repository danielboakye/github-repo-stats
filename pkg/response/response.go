package response

import (
	"encoding/json"
	"net/http"
)

// ErrorMessage is the JSON response format for errors.
type ErrorMessage struct {
	Message string `json:"message"`
}

// Error returns string format of ErrorMessage
func (e ErrorMessage) Error() string {
	return e.Message
}

// JSON sends a generic response as JSON.
func JSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	if err, ok := data.(error); ok {
		data = ErrorMessage{Message: err.Error()}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "\t") // setting indent to make responses easier to read during testing
	return encoder.Encode(data)
}

// InternalError sends a response for when there is internal error
func InternalError(w http.ResponseWriter) {
	JSON(w, http.StatusInternalServerError, ErrorMessage{"Internal Server Error"})
}

// InvalidRequest sends a response for when a request contains errors.
func InvalidRequest(w http.ResponseWriter, message string) {
	JSON(w, http.StatusBadRequest, ErrorMessage{message})
}
