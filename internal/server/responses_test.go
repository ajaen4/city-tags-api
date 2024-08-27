package server

import (
	"city-tags-api/internal/api_errors"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondWithJSON(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		payload any
	}{
		{
			"normal response",
			http.StatusOK,
			CityData{CityId: 000, CityName: "testName", Continent: "testCont", Country3Code: "testCode"},
		},
		{
			"client error response",
			http.StatusBadRequest,
			&api_errors.ClientErr{
				HttpCode: http.StatusBadRequest,
				Message:  "Parameters not present or invalid",
				LogMess:  "",
				Errors: map[string]string{
					"offset": "Not present or invalid",
				},
			},
		},
		{
			"client error response",
			http.StatusInternalServerError,
			&api_errors.InternalErr{
				HttpCode: http.StatusInternalServerError,
				Message:  "Internal error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := httptest.NewRecorder()
			respondWithJSON(writer, tt.code, tt.payload)

			expecContentType := "application/json"
			contentType := writer.Header().Get("Content-Type")
			if contentType != expecContentType {
				t.Errorf("Wrong Content-Type header, got %s wanted %s", contentType, expecContentType)
			}

			if writer.Code != tt.code {
				t.Errorf("Wrong code response, got %d wanted %d", writer.Code, tt.code)
			}

			jsonPayload, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatal(err)
			}
			body := writer.Body.String()
			if string(jsonPayload) != body {
				t.Errorf("Wrong body response, got %s wanted %s", body, jsonPayload)
			}
		})
	}
}
