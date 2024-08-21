package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerUsers(t *testing.T) {
	// Mock DB connection
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Mock instance of &apiConfig{}
	apiCfgMock := &apiConfig{
		DB: db,
	}

	// Expect the INSERT query to be called
	mock.ExpectExec("INSERT INTO users").
		WithArgs("testuser").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Create a new HTTP request
	testBody := map[string]string{
		"username": "testuser",
		"password": "test1234",
	}
	body, err := json.Marshal(testBody)
	require.NoError(t, err)

	// Create a new HTTP request with the JSON body
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Call the handler
	apiCfgMock.handlerUsers(w, req)

	// Expected payload
	expectedPayload := `{"username":"testuser"}`

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, expectedPayload, w.Body.String())
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Ensure that the SQL expectations were met
	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}
