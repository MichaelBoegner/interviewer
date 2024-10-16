package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/michaelboegner/interviewer/interview"
	"github.com/michaelboegner/interviewer/middleware"
	"github.com/michaelboegner/interviewer/token"
	"github.com/michaelboegner/interviewer/user"
)

type returnVals struct {
	Error        string            `json:"error,omitempty"`
	ID           int               `json:"id,omitempty"`
	UserID       int               `json:"user_id,omitempty"`
	Body         string            `json:"body,omitempty"`
	Username     string            `json:"username,omitempty"`
	Email        string            `json:"email,omitempty"`
	Token        string            `json:"token,omitempty"`
	Users        map[int]user.User `json:"users,omitempty"`
	Questions    map[int]string    `json:"firstQuestion,omitempty"`
	JWToken      string            `json:"jwtoken,omitempty"`
	RefreshToken string            `json:"refresh_token,omitempty"`
}

func (apiCfg *apiConfig) usersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// POST create a user
	case http.MethodPost:
		params, ok := r.Context().Value("params").(middleware.AcceptedVals)
		if !ok {
			respondWithError(w, http.StatusBadRequest, "Invalid request parameters")
		}

		if params.Username == "" || params.Email == "" || params.Password == "" {
			respondWithError(w, http.StatusBadRequest, "Username, Email, and Password required")
		} else {
			user, err := user.CreateUser(apiCfg.UserRepo, params.Username, params.Email, params.Password)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
			}

			payload := &returnVals{
				Username: user.Username,
				Email:    user.Email,
			}
			respondWithJSON(w, 200, payload)
		}

	case http.MethodGet:
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 4 || pathParts[3] == "" {
			respondWithError(w, http.StatusBadRequest, "Missing user ID")
		}

		userID, err := strconv.Atoi(pathParts[3])
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "ID passed is not an int convertable string")
		}

		user, err := user.GetUser(apiCfg.UserRepo, userID)
		if err != nil {
			log.Printf("GetUsers failed due to: %v", err)
		}

		payload := &returnVals{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		}

		respondWithJSON(w, http.StatusOK, payload)
	}
}

func (apiCfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// POST login a user
	case http.MethodPost:
		params, ok := r.Context().Value("params").(middleware.AcceptedVals)
		if !ok {
			respondWithError(w, http.StatusBadRequest, "Invalid request parameters")
		}
		if params.Username == "" || params.Password == "" {
			log.Printf("Invalid username or password.")
			respondWithError(w, http.StatusUnauthorized, "Invalid username or password.")
		} else {
			jwToken, userID, err := user.LoginUser(apiCfg.UserRepo, params.Username, params.Password)
			if err != nil {
				log.Printf("LoginUser error: %v", err)
				respondWithError(w, http.StatusUnauthorized, "Invalid username or password.")
			}

			refreshToken, err := token.CreateRefreshToken(apiCfg.TokenRepo, userID)
			if err != nil {
				log.Printf("RefreshToken error: %v", err)
				respondWithError(w, http.StatusUnauthorized, "")
			}

			payload := returnVals{
				UserID:       userID,
				Username:     params.Username,
				JWToken:      jwToken,
				RefreshToken: refreshToken,
			}

			respondWithJSON(w, http.StatusOK, payload)
		}

	}
}

func (apiCfg *apiConfig) interviewsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// POST start a resource instance of an interview and return the first question
	case http.MethodPost:
		interviewStarted, err := interview.StartInterview(apiCfg.InterviewRepo, 1, 30, 3, "easy")
		if err != nil {
			log.Printf("Interview failed to start: %v", err)
		}

		payload := returnVals{
			Questions: interviewStarted.Questions,
		}

		respondWithJSON(w, http.StatusOK, payload)
	}
}

func (apiCfg *apiConfig) refreshTokensHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// POST generate and return userID and a refreshToken
	case http.MethodPost:
		providedToken := r.Context().Value("token").(string)
		params, ok := r.Context().Value("params").(middleware.AcceptedVals)
		if !ok {
			respondWithError(w, http.StatusBadRequest, "Invalid request parameters")
		}

		storedToken, err := token.GetStoredRefreshToken(apiCfg.TokenRepo, params.UserID)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "User ID is invalid.")
		}

		ok = token.VerifyRefreshToken(storedToken, providedToken)
		if !ok {
			respondWithError(w, http.StatusUnauthorized, "Refresh token is invalid.")
		}

		refreshToken, err := token.CreateRefreshToken(apiCfg.TokenRepo, params.UserID)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "")
		}

		jwToken, err := token.CreateJWT(params.UserID, 0)
		if err != nil {
			log.Printf("JWT creation failed: %v", err)
			respondWithError(w, http.StatusInternalServerError, "")
		}

		payload := &returnVals{
			ID:           params.UserID,
			JWToken:      jwToken,
			RefreshToken: refreshToken,
		}
		respondWithJSON(w, http.StatusAccepted, payload)
	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}

	if code != 0 {
		w.WriteHeader(code)
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}

	if code != 0 {
		w.WriteHeader(code)
	}

	respBody := returnVals{
		Error: msg,
	}

	data, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		return
	}

	w.Write(data)
}
