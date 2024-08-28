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
	Email    string `json:"email,omitempty"`
	Token    string `json:"token,omitempty"`
}

type acceptedVals struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (apiCfg *apiConfig) handlerUsers(w http.ResponseWriter, r *http.Request) {
	// Unmarshal body data and return params
	_, err := getParams(r, w)
	if err != nil {
		log.Printf("Error: %v\n", err)
	}

	// switch r.Method {
	// // POST create a user
	// case http.MethodPost:
	// 	_, err := apiCfg.DB.Exec("INSERT INTO users (username) VALUES ($1)", params.Username)
	// 	if err != nil {
	// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	// 		return
	// 	}

	// 	payload := &returnVals{
	// 		Username: params.Username,
	// 	}
	// 	respondWithJSON(w, 200, payload)
	// }
}

// func (apiCfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
// 	// Unmarshal body data and return params
// 	params, err := getParams(r, w)
// 	if err != nil {
// 		log.Printf("Error: %v\n", err)
// 	}

// 	switch r.Method {
// 	// POST login a user
// 	case http.MethodPost:
// 		user, id, token, err := apiCfg.DB.LoginUser(params.Email, params.Password, jwtSecret, params.ExpiresInSeconds)
// 		if err != nil {
// 			respondWithError(w, 401, "Unauthorized")
// 		}

// 		payload := &returnVals{
// 			Id:    id,
// 			Email: user.Email,
// 			Token: token,
// 		}

// 		respondWithJSON(w, 200, payload)
// 	}
// }

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
