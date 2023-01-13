package main

import (
	"encoding/json"
	"net/http"
)

type JsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"string"`
	Data    any    `json:"data,omitempty"`
}

func (app *application) writeJson(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}

	return nil
}
