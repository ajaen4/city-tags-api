package server

import (
	"city-tags-api/internal/api_errors"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestNewHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	okHandler := func(w http.ResponseWriter, r *http.Request) error {
		return nil
	}
	okWriter := httptest.NewRecorder()
	NewHandler(okHandler).ServeHTTP(okWriter, req)

	if okWriter.Code != http.StatusOK || okWriter.Body.Len() != 0 {
		t.Errorf(
			"Handler returned wrong error response: got %v wanted no response",
			okWriter.Body,
		)
	}

	clientErr := &api_errors.ClientErr{
		HttpCode: http.StatusBadRequest,
		Message:  "Parameters not present or invalid",
		LogMess:  "",
		Errors: map[string]string{
			"offset": "Not present or invalid",
		},
	}
	handler := func(w http.ResponseWriter, r *http.Request) error {
		return clientErr
	}
	clientErrW := httptest.NewRecorder()
	NewHandler(handler).ServeHTTP(clientErrW, req)

	respErr := &api_errors.ClientErr{}
	err = json.Unmarshal(clientErrW.Body.Bytes(), respErr)
	if err != nil {
		t.Fatal(err)
	}

	if respErr.HttpCode != clientErr.HttpCode {
		t.Errorf(
			"Handler returned wrong HttpCode: got %v wanted %v",
			respErr.HttpCode,
			clientErr.HttpCode,
		)
	}

	if respErr.Message != clientErr.Message {
		t.Errorf(
			"Handler returned wrong Message: got %v wanted %v",
			respErr.Message,
			clientErr.Message,
		)
	}

	if respErr.LogMess != clientErr.LogMess {
		t.Errorf(
			"Handler returned wrong LogMess: got %v wanted %v",
			respErr.LogMess,
			clientErr.LogMess,
		)
	}

	if !reflect.DeepEqual(respErr.Errors, clientErr.Errors) {
		t.Errorf(
			"Handler returned wrong Errors: got %v wanted %v",
			respErr.Errors,
			clientErr.Errors,
		)
	}
}
