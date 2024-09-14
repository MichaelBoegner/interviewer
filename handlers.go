package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type acceptedVals struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type Users struct {
	Users map[int]User
}

type User struct {
	Id       int
	Username string
	Email    string
}
type returnVals struct {
	Error    string       `json:"error,omitempty"`
	Id       int          `json:"id,omitempty"`
	Body     string       `json:"body,omitempty"`
	Username string       `json:"username,omitempty"`
	Email    string       `json:"email,omitempty"`
	Token    string       `json:"token,omitempty"`
	Users    map[int]User `json:"users,omitempty"`
}

func (apiCfg *apiConfig) handlerUsers(w http.ResponseWriter, r *http.Request) {
	// Unmarshal body data and return params
	params, err := getParams(r, w)
	if err != nil {
		log.Printf("Error: %v\n", err)
	}

	switch r.Method {
	// POST create a user
	case http.MethodPost:
		password, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.MinCost)
		if err != nil {
			log.Printf("Error: %v\n", err)
		}

		_, err = apiCfg.DB.Exec("INSERT INTO users (username, password, email) VALUES ($1, $2, $3)", params.Username, password, params.Email)
		if err != nil {
			log.Printf("Error: %v\n", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		payload := &returnVals{
			Username: params.Username,
			Email:    params.Email,
		}
		respondWithJSON(w, 200, payload)

	case http.MethodGet:
		rows, err := apiCfg.DB.Query("SELECT id, username, email FROM users")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		userMap := make(map[int]User)
		users := &Users{
			Users: userMap,
		}
		for rows.Next() {
			user := User{}
			rows.Scan(&user.Id, &user.Username, &user.Email)
			users.Users[user.Id] = user
		}

		payload := &returnVals{
			Users: users.Users,
		}

		respondWithJSON(w, http.StatusOK, payload)
	}
}

func (apiCfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	// Unmarshal body data and return params
	params, err := getParams(r, w)
	if err != nil {
		log.Printf("Error: %v\n", err)
	}

	switch r.Method {
	// POST login a user
	case http.MethodPost:
		var hashedPassword string
		err = apiCfg.DB.QueryRow("SELECT password from users WHERE username = $1", params.Username).Scan(&hashedPassword)
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		} else if err != nil {
			log.Printf("Error querying database: %v\n", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(params.Password))
		if err != nil {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}
	}

	payload := returnVals{
		Username: params.Username,
	}

	respondWithJSON(w, http.StatusOK, payload)
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
