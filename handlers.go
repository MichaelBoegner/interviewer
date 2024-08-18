package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type returnVals struct {
	Error    string `json:"error,omitempty"`
	Id       int    `json:"id,omitempty"`
	Body     string `json:"body,omitempty"`
	Username string `json:"username,omitempty"`
}

type acceptedVals struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func handlerUsers(w http.ResponseWriter, r *http.Request) {
	// Unmarshal body data and return params
	params, err := getParams(r, w)
	if err != nil {
		log.Printf("\nError: %v", err)
	}

	switch r.Method {
	// POST create a user
	case http.MethodPost:
		payload := &returnVals{
			Username: params.Username,
		}
		respondWithJSON(w, 200, payload)
	}
}

func getParams(r *http.Request, w http.ResponseWriter) (acceptedVals, error) {
	decoder := json.NewDecoder(r.Body)
	params := acceptedVals{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Error decoding params")
		return params, err
	}

	return params, nil
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)
	w.Write(data)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.Header().Add("Content-Type", "application/json")
	respBody := returnVals{
		Error: msg,
	}

	data, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		return
	}

	w.WriteHeader(code)
	w.Write(data)
}
