package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerUsers(t *testing.T) {
	// Create a request to pass to handler
	payload := map[string]string{"username": "testuser", "password": "testpass"}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer((jsonPayload)))
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlerUsers)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	expected := `{"username":"testuser"}`
	result, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(result, []byte(expected)) {
		t.Errorf("handler returned unexpected body: \ngot %v \nwant %v\n", string(result), expected)
	}
}
