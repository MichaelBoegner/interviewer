package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/michaelboegner/interviewer/conversation"
	"github.com/michaelboegner/interviewer/interview"
	"github.com/michaelboegner/interviewer/middleware"
	"github.com/michaelboegner/interviewer/token"
	"github.com/michaelboegner/interviewer/user"
)

func (h *Handler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("healthy"))
	return
}

func (h *Handler) CreateUsersHandler(w http.ResponseWriter, r *http.Request) {
	params := &middleware.AcceptedVals{}
	err := json.NewDecoder(r.Body).Decode(params)
	if err != nil {
		log.Printf("Decoding params failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
	}

	if params.Username == "" || params.Email == "" || params.Password == "" {
		log.Printf("Missing params in usersHandler.")
		RespondWithError(w, http.StatusBadRequest, "Username, Email, and Password required")
		return
	}
	userCreated, err := user.CreateUser(h.UserRepo, params.Username, params.Email, params.Password)
	if err != nil {
		log.Printf("CreateUser error: %v", err)
		if errors.Is(err, user.ErrDuplicateEmail) || errors.Is(err, user.ErrDuplicateUsername) || errors.Is(err, user.ErrDuplicateUser) {
			RespondWithError(w, http.StatusConflict, "Email or username already exists")
			return
		}
		// For preventing user creation in frontend.
		// noNewUsers := fmt.Sprintf("%s", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	payload := &ReturnVals{
		UserID:   userCreated.ID,
		Username: userCreated.Username,
		Email:    userCreated.Email,
	}
	RespondWithJSON(w, http.StatusCreated, payload)
	return
}

func (h *Handler) GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.ContextKeyTokenParams).(int)
	if !ok {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	userIDParam, err := GetPathID(r, "/api/users/")
	if err != nil {
		log.Printf("GetPathID error: %v\n", err)
		RespondWithError(w, http.StatusBadRequest, "UserID required")
		return
	}

	if userID != userIDParam {
		log.Printf("UserID mismatch: %v vs. %v", userID, userIDParam)
		RespondWithError(w, http.StatusUnauthorized, "Invalid ID")
		return
	}

	userReturned, err := user.GetUser(h.UserRepo, userID)
	if err != nil {
		log.Printf("GetUsers error: %v", err)
		return
	}

	payload := &ReturnVals{
		UserID:   userReturned.ID,
		Username: userReturned.Username,
		Email:    userReturned.Email,
	}

	RespondWithJSON(w, http.StatusOK, payload)
	return
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	params := &middleware.AcceptedVals{}
	err := json.NewDecoder(r.Body).Decode(params)
	if err != nil {
		log.Printf("Decoding params failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
	}

	if params.Username == "" || params.Password == "" {
		log.Printf("Invalid username or password.")
		RespondWithError(w, http.StatusBadRequest, "Invalid username or password.")
		return

	}

	jwToken, userID, err := user.LoginUser(h.UserRepo, params.Username, params.Password)
	if err != nil {
		log.Printf("LoginUser error: %v", err)
		RespondWithError(w, http.StatusUnauthorized, "Invalid username or password.")
		return
	}

	refreshToken, err := token.CreateRefreshToken(h.TokenRepo, userID)
	if err != nil {
		log.Printf("RefreshToken error: %v", err)
		RespondWithError(w, http.StatusUnauthorized, "")
		return
	}

	payload := ReturnVals{
		UserID:       userID,
		Username:     params.Username,
		JWToken:      jwToken,
		RefreshToken: refreshToken,
	}

	RespondWithJSON(w, http.StatusOK, payload)
	return
}

func (h *Handler) RefreshTokensHandler(w http.ResponseWriter, r *http.Request) {
	providedToken := r.Context().Value(middleware.ContextKeyTokenKey).(string)
	params := &middleware.AcceptedVals{}

	err := json.NewDecoder(r.Body).Decode(params)
	if err != nil {
		log.Printf("Decoding params failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
	}

	if params.UserID == 0 {
		log.Printf("Invalid userID")
		RespondWithError(w, http.StatusBadRequest, "Invalid username or password")
		return
	}

	storedToken, err := token.GetStoredRefreshToken(h.TokenRepo, params.UserID)
	if err != nil {
		log.Printf("GetStoredRefreshToken error: %v", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	ok := token.VerifyRefreshToken(storedToken, providedToken)
	if !ok {
		log.Printf("VerifyRefreshToken error")
		RespondWithError(w, http.StatusUnauthorized, "Refresh token is invalid")
		return
	}

	refreshToken, err := token.CreateRefreshToken(h.TokenRepo, params.UserID)
	if err != nil {
		log.Printf("CreateRefreshToken error: %v", err)
		RespondWithError(w, http.StatusUnauthorized, "")
		return
	}

	jwToken, err := token.CreateJWT(strconv.Itoa(params.UserID), 0)
	if err != nil {
		log.Printf("JWT creation failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "")
		return
	}

	payload := &ReturnVals{
		ID:           params.UserID,
		JWToken:      jwToken,
		RefreshToken: refreshToken,
	}
	RespondWithJSON(w, http.StatusOK, payload)
	return
}

func (h *Handler) InterviewsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.ContextKeyTokenParams).(int)
	if !ok {
		RespondWithError(w, http.StatusBadRequest, "Invalid userID")
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

	RespondWithJSON(w, http.StatusCreated, payload)
	return
}

func (h *Handler) CreateConversationsHandler(w http.ResponseWriter, r *http.Request) {
	params := &middleware.AcceptedVals{}
	err := json.NewDecoder(r.Body).Decode(params)
	if err != nil {
		log.Printf("Decoding params failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
	}

	interviewID, err := GetPathID(r, "/api/conversations/create/")

	if err != nil {
		log.Printf("PathID error: %v\n", err)
		RespondWithError(w, http.StatusBadRequest, "Missing ID")
		return
	}

	interviewReturned, err := interview.GetInterview(h.InterviewRepo, interviewID)
	if err != nil {
		log.Printf("GetInterview error: %v\n", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
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
		RespondWithError(w, http.StatusBadRequest, "Invalid interview_id")
		return
	}

	payload := &ReturnVals{
		Conversation: conversationReturned,
	}
	RespondWithJSON(w, http.StatusCreated, payload)
}

func (h *Handler) AppendConversationsHandler(w http.ResponseWriter, r *http.Request) {
	params := &middleware.AcceptedVals{}
	err := json.NewDecoder(r.Body).Decode(params)
	if err != nil {
		log.Printf("Decoding params failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
	}

	if params.Message == "" {
		log.Printf("messageUserResponse is nil")
		RespondWithError(w, http.StatusBadRequest, "Missing message")
	}

	interviewID, err := GetPathID(r, "/api/conversations/append/")
	if err != nil {
		log.Printf("PathID error: %v\n", err)
		RespondWithError(w, http.StatusBadRequest, "Missing ID")
		return
	}

	interviewReturned, err := interview.GetInterview(h.InterviewRepo, interviewID)
	if err != nil {
		log.Printf("GetInterview error: %v\n", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	conversationReturned, err := conversation.GetConversation(h.ConversationRepo, params.ConversationID)
	if err != nil {
		log.Printf("GetConversation error: %v", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid ID.")
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
		RespondWithError(w, http.StatusBadRequest, "Invalid ID.")
		return
	}

	payload := &ReturnVals{
		Conversation: conversationReturned,
	}
	RespondWithJSON(w, http.StatusCreated, payload)
}

func (h *Handler) RequestResetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var params PasswordResetRequest
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		log.Printf("Decoding request failed: %v", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resetJWT, err := user.RequestPasswordReset(h.UserRepo, params.Email)
	if err != nil {
		log.Printf("Error generating reset token for email %s: %v", params.Email, err)
		w.WriteHeader(http.StatusOK)
		return
	}

	resetURL := "https://yourapp.com/reset-password?token=" + resetJWT

	err = h.Mailer.SendPasswordReset(params.Email, resetURL)
	if err != nil {
		log.Printf("SendPasswordReset error: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to send email")
		return
	}

	payload := ReturnVals{}
	RespondWithJSON(w, http.StatusOK, payload)
}

func (h *Handler) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var params PasswordResetPayload
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		log.Printf("Decoding payload failed: %v", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := user.ResetPassword(h.UserRepo, params.Token, params.NewPassword)
	if err != nil {
		log.Printf("ResetPasswordHandler failed: %v", err)
		RespondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	payload := ReturnVals{}
	RespondWithJSON(w, http.StatusOK, payload)
}
