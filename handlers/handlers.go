package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/michaelboegner/interviewer/chatgpt"
	"github.com/michaelboegner/interviewer/conversation"
	"github.com/michaelboegner/interviewer/interview"
	"github.com/michaelboegner/interviewer/middleware"
	"github.com/michaelboegner/interviewer/token"
	"github.com/michaelboegner/interviewer/user"
)

type Handler struct {
	UserRepo         user.UserRepo
	InterviewRepo    interview.InterviewRepo
	ConversationRepo conversation.ConversationRepo
	TokenRepo        token.TokenRepo
	OpenAI           chatgpt.AIClient
	DB               *sql.DB
}

func NewHandler(
	interviewRepo interview.InterviewRepo,
	userRepo user.UserRepo,
	tokenRepo token.TokenRepo,
	conversationRepo conversation.ConversationRepo,
	openAI chatgpt.AIClient,
	db *sql.DB) *Handler {
	return &Handler{
		InterviewRepo:    interviewRepo,
		UserRepo:         userRepo,
		TokenRepo:        tokenRepo,
		ConversationRepo: conversationRepo,
		OpenAI:           openAI,
		DB:               db,
	}
}

type ReturnVals struct {
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
	User           *user.User                 `json:"user,omitempty"`
}

func (h *Handler) CreateUsersHandler(w http.ResponseWriter, r *http.Request) {
	params := &middleware.AcceptedVals{}
	err := json.NewDecoder(r.Body).Decode(params)
	if err != nil {
		log.Printf("Decoding params failed: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
	}

	if params.Username == "" || params.Email == "" || params.Password == "" {
		log.Printf("Missing params in usersHandler.")
		respondWithError(w, http.StatusBadRequest, "Username, Email, and Password required")
		return
	}
	userCreated, err := user.CreateUser(h.UserRepo, params.Username, params.Email, params.Password)
	if err != nil {
		log.Printf("CreateUser error: %v", err)
		if errors.Is(err, user.ErrDuplicateEmail) || errors.Is(err, user.ErrDuplicateUsername) || errors.Is(err, user.ErrDuplicateUser) {
			respondWithError(w, http.StatusConflict, "Email or username already exists")
			return
		}
		// For preventing user creation in frontend.
		// noNewUsers := fmt.Sprintf("%s", err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	payload := &ReturnVals{
		UserID:   userCreated.ID,
		Username: userCreated.Username,
		Email:    userCreated.Email,
	}
	respondWithJSON(w, http.StatusCreated, payload)
	return
}

func (h *Handler) GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.ContextKeyTokenParams).(int)
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	userIDParam, err := getPathID(r, "/api/users/")
	if err != nil {
		log.Printf("PathID error: %v\n", err)
		respondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	if userID != userIDParam {
		log.Printf("UserID mismatch: %v vs. %v", userID, userIDParam)
		respondWithError(w, http.StatusUnauthorized, "Invalid ID")
		return
	}

	user, err := user.GetUser(h.UserRepo, userID)
	if err != nil {
		log.Printf("GetUsers error: %v", err)
		return
	}

	payload := &ReturnVals{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
	}

	respondWithJSON(w, http.StatusOK, payload)
	return
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	params := &middleware.AcceptedVals{}
	err := json.NewDecoder(r.Body).Decode(params)
	if err != nil {
		log.Printf("Decoding params failed: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
	}

	if params.Username == "" || params.Password == "" {
		log.Printf("Invalid username or password.")
		respondWithError(w, http.StatusUnauthorized, "Invalid username or password.")
		return

	}

	jwToken, userID, err := user.LoginUser(h.UserRepo, params.Username, params.Password)
	if err != nil {
		log.Printf("LoginUser error: %v", err)
		respondWithError(w, http.StatusUnauthorized, "Invalid username or password.")
		return
	}

	refreshToken, err := token.CreateRefreshToken(h.TokenRepo, userID)
	if err != nil {
		log.Printf("RefreshToken error: %v", err)
		respondWithError(w, http.StatusUnauthorized, "")
		return
	}

	payload := ReturnVals{
		UserID:       userID,
		Username:     params.Username,
		JWToken:      jwToken,
		RefreshToken: refreshToken,
	}

	respondWithJSON(w, http.StatusOK, payload)
	return
}

func (h *Handler) InterviewsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.ContextKeyTokenParams).(int)
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Invalid userID")
		return
	}

	interviewStarted, err := interview.StartInterview(h.InterviewRepo, h.OpenAI, userID, 30, 3, "easy")
	if err != nil {
		log.Printf("Interview failed to start: %v", err)
		return
	}

	payload := ReturnVals{
		InterviewID:   interviewStarted.Id,
		FirstQuestion: interviewStarted.FirstQuestion,
	}

	respondWithJSON(w, http.StatusCreated, payload)
	return
}

func (h *Handler) CreateConversationsHandler(w http.ResponseWriter, r *http.Request) {
	params := &middleware.AcceptedVals{}
	err := json.NewDecoder(r.Body).Decode(params)
	if err != nil {
		log.Printf("Decoding params failed: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
	}

	interviewID, err := getPathID(r, "/api/conversations/create/")

	if err != nil {
		log.Printf("PathID error: %v\n", err)
		respondWithError(w, http.StatusBadRequest, "Missing ID")
		return
	}

	interviewReturned, err := interview.GetInterview(h.InterviewRepo, interviewID)
	if err != nil {
		log.Printf("GetInterview error: %v\n", err)
		respondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	conversationReturned, err := conversation.CreateConversation(
		h.ConversationRepo,
		h.OpenAI,
		interviewID,
		interviewReturned.Prompt,
		interviewReturned.FirstQuestion,
		interviewReturned.Subtopic,
		params.Message)
	if err != nil {
		log.Printf("CreateConversation error: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid interview_id")
		return
	}

	payload := &ReturnVals{
		Conversation: conversationReturned,
	}
	respondWithJSON(w, http.StatusCreated, payload)
}

func (h *Handler) AppendConversationsHandler(w http.ResponseWriter, r *http.Request) {
	params := &middleware.AcceptedVals{}
	err := json.NewDecoder(r.Body).Decode(params)
	if err != nil {
		log.Printf("Decoding params failed: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
	}

	if params.Message == "" {
		log.Printf("messageUserResponse is nil")
		respondWithError(w, http.StatusBadRequest, "Missing message")
	}

	interviewID, err := getPathID(r, "/api/conversations/append/")
	if err != nil {
		log.Printf("PathID error: %v\n", err)
		respondWithError(w, http.StatusBadRequest, "Missing ID")
		return
	}

	interviewReturned, err := interview.GetInterview(h.InterviewRepo, interviewID)
	if err != nil {
		log.Printf("GetInterview error: %v\n", err)
		respondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	conversationReturned, err := conversation.GetConversation(h.ConversationRepo, params.ConversationID)
	if err != nil {
		log.Printf("GetConversation error: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid ID.")
		return
	}

	conversationReturned, err = conversation.AppendConversation(
		h.ConversationRepo,
		h.OpenAI,
		conversationReturned,
		params.Message,
		interviewReturned.Prompt)
	if err != nil {
		log.Printf("AppendConversation error: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid ID.")
		return
	}

	payload := &ReturnVals{
		Conversation: conversationReturned,
	}
	respondWithJSON(w, http.StatusCreated, payload)
}

func (h *Handler) RefreshTokensHandler(w http.ResponseWriter, r *http.Request) {
	providedToken := r.Context().Value(middleware.ContextKeyTokenKey).(string)
	params := &middleware.AcceptedVals{}
	err := json.NewDecoder(r.Body).Decode(params)
	if err != nil {
		log.Printf("Decoding params failed: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
	}

	storedToken, err := token.GetStoredRefreshToken(h.TokenRepo, params.UserID)
	if err != nil {
		log.Printf("GetStoredRefreshToken error: %v", err)
		respondWithError(w, http.StatusUnauthorized, "User ID is invalid.")
		return
	}

	ok := token.VerifyRefreshToken(storedToken, providedToken)
	if !ok {
		log.Printf("VerifyRefreshToken error.")
		respondWithError(w, http.StatusUnauthorized, "Refresh token is invalid.")
		return
	}

	refreshToken, err := token.CreateRefreshToken(h.TokenRepo, params.UserID)
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

	payload := &ReturnVals{
		ID:           params.UserID,
		JWToken:      jwToken,
		RefreshToken: refreshToken,
	}
	respondWithJSON(w, http.StatusOK, payload)
	return
}

func getPathID(r *http.Request, prefix string) (int, error) {
	path := strings.TrimPrefix(r.URL.Path, prefix)
	path = strings.Trim(path, "/")

	if path == "" {
		log.Printf("getPathID returned empty string")
		err := errors.New("Missing or invalid url param")
		return 0, err
	}

	id, err := strconv.Atoi(path)
	if err != nil {
		log.Printf("getPathID failed: %v", err)
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

	respBody := ReturnVals{
		Error: msg,
	}

	data, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		return
	}

	w.Write(data)
}
