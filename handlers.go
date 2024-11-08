package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/michaelboegner/interviewer/conversation"
	"github.com/michaelboegner/interviewer/interview"
	"github.com/michaelboegner/interviewer/middleware"
	"github.com/michaelboegner/interviewer/token"
	"github.com/michaelboegner/interviewer/user"
)

type returnVals struct {
	ID             int                        `json:"id,omitempty"`
	UserID         int                        `json:"user_id,omitempty"`
	InterviewID    int                        `json:"interview_id,omitempty"`
	ConversationID int                        `json:"conversation_id,omitempty"`
	Body           string                     `json:"body,omitempty"`
	Username       string                     `json:"username,omitempty"`
	Email          string                     `json:"email,omitempty"`
	Token          string                     `json:"token,omitempty"`
	FirstQuestion  string                     `json:"first_question,omitempty"`
	NextQuestion   string                     `json:"next_question,omitempty"`
	JWToken        string                     `json:"jwtoken,omitempty"`
	RefreshToken   string                     `json:"refresh_token,omitempty"`
	Error          string                     `json:"error,omitempty"`
	Users          map[int]user.User          `json:"users,omitempty"`
	Conversation   *conversation.Conversation `json:"conversation,omitempty"`
}

func (apiCfg *apiConfig) usersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// POST create a user
	case http.MethodPost:
		params, ok := r.Context().Value("params").(middleware.AcceptedVals)
		if !ok {
			respondWithError(w, http.StatusBadRequest, "Invalid request parameters")
			return
		}

		if params.Username == "" || params.Email == "" || params.Password == "" {
			log.Printf("Missing params in usersHandler.")
			respondWithError(w, http.StatusBadRequest, "Username, Email, and Password required")
			return
		}
		user, err := user.CreateUser(apiCfg.UserRepo, params.Username, params.Email, params.Password)
		if err != nil {
			log.Printf("CreateUser error: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		payload := &returnVals{
			Username: user.Username,
			Email:    user.Email,
		}
		respondWithJSON(w, http.StatusOK, payload)
		return

	// GET a user by {id}
	case http.MethodGet:
		userID, err := getPathID(r)
		if err != nil {
			log.Printf("PathID error: %v\n", err)
			respondWithError(w, http.StatusBadRequest, "Invalid ID.")
		}

		user, err := user.GetUser(apiCfg.UserRepo, userID)
		if err != nil {
			log.Printf("GetUsers error: %v", err)
			return
		}

		payload := &returnVals{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		}

		respondWithJSON(w, http.StatusOK, payload)
		return
	}
}

func (apiCfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// POST login a user
	case http.MethodPost:
		params, ok := r.Context().Value("params").(middleware.AcceptedVals)
		if !ok {
			respondWithError(w, http.StatusBadRequest, "Invalid request parameters")
			return
		}

		if params.Username == "" || params.Password == "" {
			log.Printf("Invalid username or password.")
			respondWithError(w, http.StatusUnauthorized, "Invalid username or password.")
			return

		}

		jwToken, userID, err := user.LoginUser(apiCfg.UserRepo, params.Username, params.Password)
		if err != nil {
			log.Printf("LoginUser error: %v", err)
			respondWithError(w, http.StatusUnauthorized, "Invalid username or password.")
			return
		}

		refreshToken, err := token.CreateRefreshToken(apiCfg.TokenRepo, userID)
		if err != nil {
			log.Printf("RefreshToken error: %v", err)
			respondWithError(w, http.StatusUnauthorized, "")
			return
		}

		payload := returnVals{
			UserID:       userID,
			Username:     params.Username,
			JWToken:      jwToken,
			RefreshToken: refreshToken,
		}

		respondWithJSON(w, http.StatusOK, payload)
		return

	}
}

func (apiCfg *apiConfig) interviewsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// POST start a resource instance of an interview and return the first question
	case http.MethodPost:
		token, ok := r.Context().Value("tokenKey").(string)
		if !ok {
			respondWithError(w, http.StatusBadRequest, "Invalid request parameters")
			return
		}

		userID, err := middleware.VerifyToken(token)
		if err != nil {
			log.Printf("Supplied token returns error: %v", err)
			respondWithError(w, http.StatusUnauthorized, "Unauthorized.")
			return
		}

		interviewStarted, err := interview.StartInterview(apiCfg.InterviewRepo, userID, 30, 3, "easy")
		if err != nil {
			log.Printf("Interview failed to start: %v", err)
			return
		}

		payload := returnVals{
			InterviewID:   interviewStarted.Id,
			FirstQuestion: interviewStarted.FirstQuestion,
		}

		respondWithJSON(w, http.StatusOK, payload)
		return
	}
}

func (apiCfg *apiConfig) conversationsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// POST start an instance of conversations and return next question
	case http.MethodPost:
		params, ok := r.Context().Value("params").(middleware.UpdateConversation)
		if !ok {
			respondWithError(w, http.StatusBadRequest, "Invalid request parameters")
			return
		}

		InterviewID, err := getPathID(r)
		if err != nil {
			log.Printf("PathID error: %v\n", err)
			respondWithError(w, http.StatusBadRequest, "Invalid ID.")
		}

		var conversationFromDatabase *conversation.Conversation
		exists := conversation.CheckForConversation(apiCfg.ConversationRepo, InterviewID)
		if !exists {
			interview, err := interview.GetInterview(apiCfg.InterviewRepo, InterviewID)
			if err != nil {
				log.Printf("GetInterview error: %v\n", err)
				respondWithError(w, http.StatusBadRequest, "Invalid interview_id")
				return
			}

			conversationFromDatabase, err = conversation.CreateConversation(apiCfg.ConversationRepo, InterviewID, interview.FirstQuestion, params.Message)
			if err != nil {
				log.Printf("CreateConversation error: %v", err)
				respondWithError(w, http.StatusBadRequest, "Invalid interview_id")
				return
			}
		} else {
			conversationFromDatabase, err = conversation.GetConversation(apiCfg.ConversationRepo, InterviewID)
			if err != nil {
				log.Printf("GetConversation error: %v", err)
				respondWithError(w, http.StatusBadRequest, "Invalid ID.")
				return
			}

			conversationFromDatabase, err = conversation.AppendConversation(apiCfg.ConversationRepo, conversationFromDatabase, params.Message, params.ConversationID, params.TopicID, params.QuestionID)
			if err != nil {
				log.Printf("AppendConversation error: %v", err)
				respondWithError(w, http.StatusBadRequest, "Invalid ID.")
				return
			}
		}

		payload := &returnVals{
			Conversation: conversationFromDatabase,
		}
		respondWithJSON(w, http.StatusOK, payload)
	}
}

func (apiCfg *apiConfig) refreshTokensHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// POST generate and return userID and a refreshToken
	case http.MethodPost:
		providedToken := r.Context().Value("tokenKey").(string)
		params, ok := r.Context().Value("params").(middleware.AcceptedVals)
		if !ok {
			respondWithError(w, http.StatusBadRequest, "Invalid request parameters")
			return
		}

		storedToken, err := token.GetStoredRefreshToken(apiCfg.TokenRepo, params.UserID)
		if err != nil {
			log.Printf("GetStoredRefreshToken error: %v", err)
			respondWithError(w, http.StatusUnauthorized, "User ID is invalid.")
			return
		}

		ok = token.VerifyRefreshToken(storedToken, providedToken)
		if !ok {
			log.Printf("VerifyRefreshToken error.")
			respondWithError(w, http.StatusUnauthorized, "Refresh token is invalid.")
			return
		}

		refreshToken, err := token.CreateRefreshToken(apiCfg.TokenRepo, params.UserID)
		if err != nil {
			log.Printf("CreateRefreshToken error: %v", err)
			respondWithError(w, http.StatusUnauthorized, "")
			return
		}

		jwToken, err := token.CreateJWT(params.UserID, 0)
		if err != nil {
			log.Printf("JWT creation failed: %v", err)
			respondWithError(w, http.StatusInternalServerError, "")
			return
		}

		payload := &returnVals{
			ID:           params.UserID,
			JWToken:      jwToken,
			RefreshToken: refreshToken,
		}
		respondWithJSON(w, http.StatusOK, payload)
		return
	}
}

func getPathID(r *http.Request) (int, error) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 || pathParts[3] == "" {
		err := errors.New("missing user ID")
		return 0, err
	}

	id, err := strconv.Atoi(pathParts[3])
	if err != nil {
		return 0, err
	}

	return id, nil
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
