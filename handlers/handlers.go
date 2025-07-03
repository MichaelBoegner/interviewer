package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/michaelboegner/interviewer/billing"
	"github.com/michaelboegner/interviewer/chatgpt"
	"github.com/michaelboegner/interviewer/conversation"
	"github.com/michaelboegner/interviewer/dashboard"
	"github.com/michaelboegner/interviewer/interview"
	"github.com/michaelboegner/interviewer/middleware"
	"github.com/michaelboegner/interviewer/token"
	"github.com/michaelboegner/interviewer/user"
)

func (h *Handler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("healthy"))
	return
}

func (h *Handler) RequestVerificationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	verificationJWT, err := user.VerificationToken(req.Email, req.Username, req.Password)
	if err != nil {
		log.Printf("GenerateEmailVerificationToken failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to create token")
		return
	}

	verifyURL := os.Getenv("FRONTEND_URL") + "verify-email?token=" + verificationJWT
	go func() {
		if err := h.Mailer.SendVerificationEmail(req.Email, verifyURL); err != nil {
			log.Printf("SendVerificationEmail failed: %v", err)
		}
	}()

	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Verification email sent"})
}

func (h *Handler) CheckEmailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Decoding check-email body failed: %v", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Email == "" {
		RespondWithError(w, http.StatusBadRequest, "Email required")
		return
	}

	err := user.GetUserByEmail(h.UserRepo, req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			RespondWithJSON(w, http.StatusOK, map[string]bool{"exists": false})
			return
		}
		log.Printf("CheckEmailHandler internal error: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]bool{"exists": true})
}

func (h *Handler) CreateUsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Token string `json:"token"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("Decoding params failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	userCreated, err := user.CreateUser(h.UserRepo, req.Token)
	if err != nil {
		log.Printf("CreateUser error: %v", err)
		if errors.Is(err, user.ErrDuplicateEmail) || errors.Is(err, user.ErrDuplicateUsername) || errors.Is(err, user.ErrDuplicateUser) {
			RespondWithError(w, http.StatusConflict, "Email already exists")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	err = h.Mailer.SendWelcome(userCreated.Email)
	if err != nil {
		log.Printf("h.Mailer.SendWelcome failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	payload := &ReturnVals{
		UserID:   userCreated.ID,
		Username: userCreated.Username,
		Email:    userCreated.Email,
	}
	RespondWithJSON(w, http.StatusCreated, payload)
}

func (h *Handler) GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

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

func (h *Handler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value(middleware.ContextKeyTokenParams).(int)
	if !ok {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	userIDParam, err := GetPathID(r, "/api/users/delete/")
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
		log.Printf("GetUser error: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to find user")
		return
	}

	err = h.Billing.CancelSubscription(h.UserRepo, userReturned.Email)
	if err != nil {
		log.Printf("h.Billing.CancelSubscription failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	err = user.MarkUserDeleted(h.UserRepo, userID)
	if err != nil {
		log.Printf("DeleteUser failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	err = token.DeleteRefreshToken(h.TokenRepo, userID)
	if err != nil {
		log.Printf("DeleteRefreshTokensForUser failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	err = h.Mailer.SendDeletionConfirmation(userReturned.Email)
	if err != nil {
		log.Printf("h.Mailer.SendDeletionConfirmation failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	params := &middleware.AcceptedVals{}
	err := json.NewDecoder(r.Body).Decode(params)
	if err != nil {
		log.Printf("Decoding params failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
	}

	if params.Email == "" || params.Password == "" {
		log.Printf("Invalid username or password.")
		RespondWithError(w, http.StatusBadRequest, "Invalid username or password.")
		return

	}

	jwToken, username, userID, err := user.LoginUser(h.UserRepo, params.Email, params.Password)
	if err != nil {
		log.Printf("LoginUser error: %v", err)
		if errors.Is(err, user.ErrAccountDeleted) {
			RespondWithError(w, http.StatusUnauthorized, user.ErrAccountDeleted.Error())
			return
		}
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
		Username:     username,
		JWToken:      jwToken,
		RefreshToken: refreshToken,
	}

	RespondWithJSON(w, http.StatusOK, payload)
}

func (h *Handler) GithubLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	data := url.Values{
		"client_id":     {os.Getenv("GITHUB_CLIENT_ID")},
		"client_secret": {os.Getenv("GITHUB_CLIENT_SECRET")},
		"code":          {body.Code},
	}

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create request")
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "GitHub token exchange failed")
		return
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Token parse failed")
		return
	}

	client = &http.Client{}
	req, err = http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		log.Printf("http.NewRequest failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	githubResp, err := client.Do(req)
	if err != nil {
		log.Printf("GET api.github.com/user failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer githubResp.Body.Close()

	var githubUser struct {
		Email string `json:"email"`
		Login string `json:"login"`
	}
	json.NewDecoder(githubResp.Body).Decode(&githubUser)

	if githubUser.Email == "" {
		req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
		if err != nil {
			log.Printf("http.NewRequest failed: %v", err)
			RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
		emailResp, err := client.Do(req)
		if err != nil {
			log.Printf("GET api.github.com/user/emails failed: %v", err)
			RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		defer emailResp.Body.Close()

		var emails []struct {
			Email    string `json:"email"`
			Primary  bool   `json:"primary"`
			Verified bool   `json:"verified"`
		}
		json.NewDecoder(emailResp.Body).Decode(&emails)

		for _, e := range emails {
			if e.Primary && e.Verified {
				githubUser.Email = e.Email
				break
			}
		}
	}

	if githubUser.Email == "" {
		log.Printf("GitHub login failed: no verified email found for user %s", githubUser.Login)
		RespondWithError(w, http.StatusUnauthorized, "We couldn’t retrieve a valid email address from GitHub. Please check your GitHub email settings and try again.")
		return
	}

	user, err := user.GetOrCreateByEmail(h.UserRepo, githubUser.Email, githubUser.Login)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "User creation failed")
		return
	}

	jwt, err := token.CreateJWT(strconv.Itoa(user.ID), 0)
	if err != nil {
		log.Printf("token.CreateJWT failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	refreshToken, err := token.CreateRefreshToken(h.TokenRepo, user.ID)
	if err != nil {
		log.Printf("token.CreateRefreshToken failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]any{
		"userID":       user.ID,
		"username":     user.Username,
		"jwt":          jwt,
		"refreshToken": refreshToken,
	})
}

func (h *Handler) RefreshTokensHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

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

	user, err := h.UserRepo.GetUser(params.UserID)
	if err != nil || user.AccountStatus == "deleted" {
		log.Printf("Refresh attempt for deleted account ID %d", params.UserID)
		RespondWithError(w, http.StatusUnauthorized, "Account deactivated")
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
}

func (h *Handler) InterviewsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value(middleware.ContextKeyTokenParams).(int)
	if !ok {
		RespondWithError(w, http.StatusBadRequest, "Invalid userID")
		return
	}

	params := &middleware.AcceptedVals{}
	err := json.NewDecoder(r.Body).Decode(params)
	if err != nil {
		log.Printf("Decoding params failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	userReturned, err := user.GetUser(h.UserRepo, userID)
	if err != nil {
		log.Printf("GetUser error: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to find user")
		return
	}

	interviewStarted, err := interview.StartInterview(
		h.InterviewRepo,
		h.UserRepo,
		h.BillingRepo,
		h.OpenAI,
		userReturned,
		30,
		3,
		"easy",
		params.JD)
	if err != nil {
		var openaiErr *chatgpt.OpenAIError
		if errors.As(err, &openaiErr) {
			log.Printf("OpenAI error: %v", openaiErr)
			RespondWithError(w, openaiErr.StatusCode, openaiErr.Message)
			return
		}
		if errors.Is(err, interview.ErrNoValidCredits) {
			RespondWithError(w, http.StatusPaymentRequired, "You do not have enough credits to start a new interview or your subscription has expired.")
			return
		}
		log.Printf("Interview failed to start: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to start interview.")
		return
	}

	conversationID, err := conversation.CreateEmptyConversation(h.ConversationRepo, interviewStarted.Id, interviewStarted.Subtopic)
	if err != nil {
		log.Printf("conversation.CreateEmptyConversation failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal server error")
	}

	err = interview.LinkConversation(h.InterviewRepo, interviewStarted.Id, conversationID)
	if err != nil {
		log.Printf("interview.LinkConversation failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal server error")
	}

	payload := ReturnVals{
		InterviewID:    interviewStarted.Id,
		FirstQuestion:  interviewStarted.FirstQuestion,
		ConversationID: conversationID,
	}

	RespondWithJSON(w, http.StatusCreated, payload)
}

func (h *Handler) GetInterviewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value(middleware.ContextKeyTokenParams).(int)
	if !ok {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	interviewID, err := GetPathID(r, "/api/interviews/")
	if err != nil {
		log.Printf("GetPathID failed: %v", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid interview ID")
		return
	}

	interviewReturned, err := interview.GetInterview(h.InterviewRepo, interviewID)
	if err != nil {
		log.Printf("GetInterview failed: %v", err)
		RespondWithError(w, http.StatusNotFound, "Interview not found")
		return
	}
	if interviewReturned.UserId != userID {
		log.Printf("User ID mismatch on interview fetch")
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	payload := ReturnVals{
		Interview: interviewReturned,
	}

	RespondWithJSON(w, http.StatusOK, payload)
}

func (h *Handler) UpdateInterviewStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value(middleware.ContextKeyTokenParams).(int)
	if !ok {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	interviewID, err := GetPathID(r, "/api/interviews/")
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid interview ID")
		return
	}

	var payload struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || (payload.Status != "paused" && payload.Status != "active") {
		RespondWithError(w, http.StatusBadRequest, "Invalid status")
		return
	}

	interviewReturned, err := interview.GetInterview(h.InterviewRepo, interviewID)
	if err != nil {
		log.Printf("GetInterview error: %v\n", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	err = ValidateInterviewStatusTransition(interviewReturned.Status, payload.Status)
	if err != nil {
		log.Printf("ValidateInterviewStatusTransition failed: %v", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid status transition")
		return
	}

	err = h.InterviewRepo.UpdateStatus(interviewID, userID, payload.Status)
	if err != nil {
		log.Printf("UpdateInterviewStatus failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Could not update status")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) CreateConversationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value(middleware.ContextKeyTokenParams).(int)
	if !ok {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

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
	if interviewReturned.UserId != userID {
		log.Printf("interview.userid != token user_id")
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if interviewReturned.Status != "active" {
		RespondWithError(w, http.StatusConflict, "Interview is not active")
		return
	}

	conversationReturned, err := conversation.GetConversation(h.ConversationRepo, interviewID)
	if err != nil {
		log.Printf("conversation.GetConversation failed: %v", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
	}

	conversationCreated, err := conversation.CreateConversation(
		h.ConversationRepo,
		h.InterviewRepo,
		h.OpenAI,
		conversationReturned,
		interviewID,
		interviewReturned.Prompt,
		interviewReturned.FirstQuestion,
		interviewReturned.Subtopic,
		params.Message)
	if err != nil {
		var openaiErr *chatgpt.OpenAIError
		if errors.As(err, &openaiErr) {
			log.Printf("OpenAI error: %v", openaiErr)
			RespondWithError(w, openaiErr.StatusCode, openaiErr.Message)
			return
		}
		log.Printf("CreateConversation error: %v", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid interview_id")
		return
	}

	payload := &ReturnVals{
		Conversation: conversationCreated,
	}
	RespondWithJSON(w, http.StatusCreated, payload)
}

func (h *Handler) AppendConversationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value(middleware.ContextKeyTokenParams).(int)
	if !ok {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

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
	if interviewReturned.UserId != userID {
		log.Printf("interview.userid != token user_id")
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if interviewReturned.Status != "active" {
		RespondWithError(w, http.StatusConflict, "Interview is not active")
		return
	}

	conversationReturned, err := conversation.GetConversation(h.ConversationRepo, interviewID)
	if err != nil {
		log.Printf("GetConversation error: %v", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid ID.")
		return
	}

	conversationReturned, err = conversation.AppendConversation(
		h.ConversationRepo,
		h.InterviewRepo,
		h.OpenAI,
		interviewID,
		userID,
		conversationReturned,
		params.Message,
		interviewReturned.Prompt)
	if err != nil {
		var openaiErr *chatgpt.OpenAIError
		if errors.As(err, &openaiErr) {
			log.Printf("OpenAI error: %v", openaiErr)
			RespondWithError(w, openaiErr.StatusCode, openaiErr.Message)
			return
		}
		log.Printf("AppendConversation error: %v", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid ID.")
		return
	}

	payload := &ReturnVals{
		Conversation: conversationReturned,
	}
	RespondWithJSON(w, http.StatusCreated, payload)
}

func (h *Handler) GetConversationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value(middleware.ContextKeyTokenParams).(int)
	if !ok {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	interviewID, err := GetPathID(r, "/api/conversations/")
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
	if interviewReturned.UserId != userID {
		log.Printf("interview.userid != token user_id")
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	conversationReturned, err := conversation.GetConversation(h.ConversationRepo, interviewID)
	if err != nil {
		log.Printf("GetConversation error: %v", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid ID.")
		return
	}

	payload := &ReturnVals{
		Conversation: conversationReturned,
	}
	RespondWithJSON(w, http.StatusOK, payload)
}

func (h *Handler) RequestResetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
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

	frontendURL := os.Getenv("FRONTEND_URL")
	resetURL := frontendURL + "reset-password?token=" + resetJWT

	go func(email, resetURL string) {
		err := h.Mailer.SendPasswordReset(email, resetURL)
		if err != nil {
			log.Printf("SendPasswordReset error: %v", err)
			return
		}
	}(params.Email, resetURL)

	payload := ReturnVals{}
	RespondWithJSON(w, http.StatusOK, payload)
}

func (h *Handler) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var params PasswordResetPayload
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		log.Printf("Decoding payload failed: %v", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	err := user.ResetPassword(h.UserRepo, params.NewPassword, params.Token)
	if err != nil {
		log.Printf("ResetPasswordHandler failed: %v", err)
		RespondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	payload := ReturnVals{}
	RespondWithJSON(w, http.StatusOK, payload)
}

func (h *Handler) CreateCheckoutSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value(middleware.ContextKeyTokenParams).(int)
	if !ok {
		log.Printf("r.Context().Value() failed")
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var params CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil || params.Tier == "" {
		log.Printf("jsonNewDecoder failed: %v", err)
		RespondWithError(w, http.StatusBadRequest, "Missing or invalid tier")
		return
	}

	user, err := user.GetUser(h.UserRepo, userID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Could not find user")
		return
	}

	var priceID string
	switch params.Tier {
	case "individual":
		priceID = os.Getenv("LEMON_VARIANT_ID_INDIVIDUAL")
	case "pro":
		priceID = os.Getenv("LEMON_VARIANT_ID_PRO")
	case "premium":
		priceID = os.Getenv("LEMON_VARIANT_ID_PREMIUM")
	default:
		RespondWithError(w, http.StatusBadRequest, "Invalid tier selected")
		return
	}

	priceIDInt, err := strconv.Atoi(priceID)
	if err != nil {
		log.Printf("strconv.Atoi() failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	url, err := h.Billing.RequestCheckoutSession(user.Email, priceIDInt)
	if err != nil {
		log.Printf("billing.CreateCheckoutSession failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Could not start checkout")
		return
	}

	RespondWithJSON(w, http.StatusOK, CheckoutResponse{CheckoutURL: url})
}

func (h *Handler) CancelSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value(middleware.ContextKeyTokenParams).(int)
	if !ok {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	userReturned, err := user.GetUser(h.UserRepo, userID)
	if err != nil {
		log.Printf("GetUser failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Could not retrieve user")
		return
	}

	err = h.Billing.RequestDeleteSubscription(userReturned.SubscriptionID)
	if err != nil {
		log.Printf("DeleteSubscription failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Could not cancel subscription")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ResumeSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value(middleware.ContextKeyTokenParams).(int)
	if !ok {
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	userReturned, err := user.GetUser(h.UserRepo, userID)
	if err != nil {
		log.Printf("GetUser failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Could not retrieve user")
		return
	}

	err = h.Billing.RequestResumeSubscription(userReturned.SubscriptionID)
	if err != nil {
		log.Printf("DeleteSubscription failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Could not cancel subscription")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ChangePlanHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value(middleware.ContextKeyTokenParams).(int)
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	var params CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil || params.Tier == "" {
		log.Printf("jsonNewDecoder failed: %v", err)
		RespondWithError(w, http.StatusBadRequest, "Missing or invalid tier")
		return
	}

	user, err := user.GetUser(h.UserRepo, userID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Could not find user")
		return
	}

	var priceID string
	switch params.Tier {
	case "pro":
		priceID = os.Getenv("LEMON_VARIANT_ID_PRO")
	case "premium":
		priceID = os.Getenv("LEMON_VARIANT_ID_PREMIUM")
	default:
		RespondWithError(w, http.StatusBadRequest, "Invalid tier")
		return
	}

	priceIDInt, err := strconv.Atoi(priceID)
	if err != nil {
		log.Printf("strconv.Atoi() failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if err := h.Billing.RequestUpdateSubscriptionVariant(user.SubscriptionID, priceIDInt); err != nil {
		log.Printf("UpdateLemonSubscriptionVariant failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to update subscription")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) BillingWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("io.ReadAll failed: %v", err)
		RespondWithError(w, http.StatusBadRequest, "Bad Request")
		return
	}
	defer r.Body.Close()

	signature := r.Header.Get("X-Signature")
	if !h.Billing.VerifyBillingSignature(signature, body, os.Getenv("LEMON_WEBHOOK_SECRET")) {
		log.Printf("Invalid billing event signature")
		RespondWithError(w, http.StatusUnauthorized, "Invalid signature")
		return
	}

	var webhookPayload billing.BillingWebhookPayload
	err = json.Unmarshal(body, &webhookPayload)
	if err != nil {
		log.Printf("json.Unmarshal failed: %v", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}
	subscriptionID := webhookPayload.Data.SubscriptionID
	webhookID := webhookPayload.Meta.WebhookID
	exists, err := h.BillingRepo.HasWebhookBeenProcessed(webhookID)
	if err != nil {
		log.Printf("h.BillingRepo.HasWebhookBeenProcessed failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Error checking webhook")
		return
	}
	if exists {
		log.Printf("Webhook %s already processed", webhookID)
		w.WriteHeader(http.StatusOK)
		return
	}

	var emailAttribute struct {
		UserEmail string `json:"user_email"`
	}

	eventType := webhookPayload.Meta.EventName

	log.Printf("Received webhook: eventType=%q, webhookID=%s, subscriptionID=%s", eventType, webhookID, subscriptionID)

	switch eventType {
	case "order_created":
		var orderAttrs billing.OrderAttributes
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &orderAttrs); err != nil {
			log.Printf("Unmarshal order_created failed: %v", err)
			RespondWithError(w, http.StatusBadRequest, "Invalid order_created payload")
			return
		}

		err = h.Billing.ApplyCredits(h.UserRepo, h.BillingRepo, orderAttrs.UserEmail, orderAttrs.FirstOrderItem.VariantID)
		if err != nil {
			log.Printf("h.Billing.ApplyCredits failed: %v", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
			return
		}
	case "subscription_created":
		var SubCreatedAttrs billing.SubscriptionAttributes
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &SubCreatedAttrs); err != nil {
			log.Printf("Unmarshal subscription_created failed: %v", err)
			RespondWithError(w, http.StatusBadRequest, "Invalid subscription_created payload")
			return
		}

		exists, err := h.UserRepo.HasActiveOrCancelledSubscription(SubCreatedAttrs.UserEmail)
		if err != nil {
			log.Printf("Subscription duplicate check failed: %v", err)
			RespondWithError(w, http.StatusInternalServerError, "Subscription check failed")
			return
		}
		if exists {
			log.Printf("Duplicate subscription attempt blocked for %s", SubCreatedAttrs.UserEmail)
			return
		}

		err = h.Billing.CreateSubscription(h.UserRepo, SubCreatedAttrs, subscriptionID)
		if err != nil {
			log.Printf("h.Billing.CreateSubscription failed: %v", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
			return
		}
	case "subscription_cancelled":
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &emailAttribute); err != nil {
			log.Printf("Unmarshal subscription_cancelled failed: %v", err)
			RespondWithError(w, http.StatusBadRequest, "Invalid subscription_cancelled payload")
			return
		}

		err = h.Billing.CancelSubscription(h.UserRepo, emailAttribute.UserEmail)
		if err != nil {
			log.Printf("h.Billing.CancelSubscription failed: %v", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
			return
		}
	case "subscription_resumed":
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &emailAttribute); err != nil {
			log.Printf("Unmarshal subscription_resumed failed: %v", err)
			RespondWithError(w, http.StatusBadRequest, "Invalid subscription_resumed payload")
			return
		}

		err = h.Billing.ResumeSubscription(h.UserRepo, emailAttribute.UserEmail)
		if err != nil {
			log.Printf("h.Billing.ResumeSubscription failed: %v", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
			return
		}
	case "subscription_expired":
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &emailAttribute); err != nil {
			log.Printf("Unmarshal subscription_expired failed: %v", err)
			RespondWithError(w, http.StatusBadRequest, "Invalid subscription_expired payload")
			return
		}

		err = h.Billing.ExpireSubscription(h.UserRepo, h.BillingRepo, emailAttribute.UserEmail)
		if err != nil {
			log.Printf("h.Billing.ExpireSubscription failed: %v", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
			return
		}
	case "subscription_payment_success":
		var SubRenewAttrs billing.SubscriptionRenewAttributes
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &SubRenewAttrs); err != nil {
			log.Printf("Unmarshal subscription_payment_success failed: %v", err)
			RespondWithError(w, http.StatusBadRequest, "Invalid subscription_payment_success payload")
			return
		}

		if SubRenewAttrs.BillingReason == "initial" {
			log.Println("Skipping credits on initial charge (already granted via order_created)")
			return
		}

		err = h.Billing.RenewSubscription(h.UserRepo, h.BillingRepo, SubRenewAttrs)
		if err != nil {
			log.Printf("h.Billing.RenewSubscription failed: %v", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
			return
		}
	case "subscription_plan_changed":
		var SubChangedAttrs billing.SubscriptionAttributes
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &SubChangedAttrs); err != nil {
			log.Printf("Unmarshal subscription_plan_changed failed: %v", err)
			RespondWithError(w, http.StatusBadRequest, "Invalid subscription_plan_changed payload")
			return
		}

		err = h.Billing.ChangeSubscription(h.UserRepo, h.BillingRepo, SubChangedAttrs)
		if err != nil {
			log.Printf("h.Billing.ChangeSubscription failed: %v", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
			return
		}
	case "subscription_updated":
		var SubChangedAttrs billing.SubscriptionAttributes
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &SubChangedAttrs); err != nil {
			log.Printf("Unmarshal subscription_updated failed: %v", err)
			RespondWithError(w, http.StatusBadRequest, "Invalid subscription_updated payload")
			return
		}

		err = h.Billing.UpdateSubscription(h.UserRepo, SubChangedAttrs, subscriptionID)
		if err != nil {
			log.Printf("h.Billing.UpdateSubscription failed: %v", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
			return
		}
	case "order_refunded":
		var orderAttrs billing.OrderAttributes
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &orderAttrs); err != nil {
			log.Printf("Unmarshal order_created failed: %v", err)
			RespondWithError(w, http.StatusBadRequest, "Invalid order_created payload")
			return
		}

		err = h.Billing.DeductCredits(h.UserRepo, h.BillingRepo, orderAttrs)
		if err != nil {
			log.Printf("h.Billing.DeductCredits failed: %v", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
			return
		}
	case "subscription_payment_failed", "subscription_payment_recovered":
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &emailAttribute); err != nil {
			log.Printf("Unmarshal %s failed: %v", eventType, err)
			RespondWithError(w, http.StatusBadRequest, "Invalid payment status payload")
			return
		}
		log.Printf("Payment event: %s for user %s", eventType, emailAttribute.UserEmail)
	default:
		log.Printf("Unhandled event type: %s", eventType)
		RespondWithError(w, http.StatusNotImplemented, "Unhandled event type")
		return
	}

	if err != nil {
		log.Printf("eventType switch func failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	err = h.BillingRepo.MarkWebhookProcessed(webhookID, eventType)
	if err != nil {
		log.Printf("MarkWebhookProcessed failed: %v", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value(middleware.ContextKeyTokenParams).(int)
	if !ok {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	dashboardData, err := dashboard.GetDashboardData(userID, h.UserRepo, h.InterviewRepo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			RespondWithError(w, http.StatusUnauthorized, "User not found")
			return
		}
		log.Printf("dashboard.GetDashboardData failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Could not load dashboard")
		return
	}

	RespondWithJSON(w, http.StatusOK, dashboardData)
}

func (h *Handler) JDInputHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var input struct {
		JobDescription string `json:"job_description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.JobDescription == "" {
		RespondWithError(w, http.StatusBadRequest, "Invalid job description")
		return
	}

	jdInput, err := h.OpenAI.ExtractJDInput(input.JobDescription)
	if err != nil {
		var openaiErr *chatgpt.OpenAIError
		if errors.As(err, &openaiErr) {
			log.Printf("OpenAI error: %v", openaiErr)
			RespondWithError(w, openaiErr.StatusCode, openaiErr.Message)
			return
		}
		log.Printf("chatgpt.ExtractJDInput failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to process job description")
		return
	}

	jdSummary, err := h.OpenAI.ExtractJDSummary(jdInput)
	if err != nil {
		var openaiErr *chatgpt.OpenAIError
		if errors.As(err, &openaiErr) {
			log.Printf("OpenAI error: %v", openaiErr)
			RespondWithError(w, openaiErr.StatusCode, openaiErr.Message)
			return
		}
		log.Printf("chatgpt.ExtractJDSummary failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to process job description")
		return

	}

	RespondWithJSON(w, http.StatusOK, jdSummary)
}
