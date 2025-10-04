package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
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

	if req.Email == "" || req.Username == "" || req.Password == "" {
		RespondWithError(w, http.StatusBadRequest, "Username, Email, and Password required")
		return
	}

	verificationJWT, err := user.VerificationToken(req.Email, req.Username, req.Password)
	if err != nil {
		h.Logger.Error("GenerateEmailVerificationToken failed", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to create token")
		return
	}

	verifyURL := os.Getenv("FRONTEND_URL") + "verify-email?token=" + verificationJWT

	go func(email, url string) {
		if err := h.Mailer.SendVerificationEmail(email, url); err != nil {
			h.Logger.Error("SendVerificationEmail failed", "error", err)
		}
	}(req.Email, verifyURL)

	payload := &ReturnVals{
		Message: "Verification email sent",
	}

	RespondWithJSON(w, http.StatusOK, payload)
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
		h.Logger.Error("Decoding check-email body failed", "error", err)
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
		h.Logger.Error("CheckEmailHandler internal error", "error", err)
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
		h.Logger.Error("Decoding params failed", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	userCreated, err := user.CreateUser(h.UserRepo, req.Token)
	if err != nil {
		h.Logger.Error("CreateUser error", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	jwt, err := token.CreateJWT(strconv.Itoa(userCreated.ID), 0)
	if err != nil {
		h.Logger.Error("token.CreateJWT failed", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	go func(email string) {
		if err := h.Mailer.SendWelcome(email); err != nil {
			h.Logger.Error("SendWelcome failed", "error", err)
		}
	}(userCreated.Email)

	payload := &ReturnVals{
		UserID:   userCreated.ID,
		Username: userCreated.Username,
		Email:    userCreated.Email,
		JWToken:  jwt,
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

	userIDParam, err := GetPathID(r, "/api/users/", h.Logger)
	if err != nil {
		h.Logger.Error("GetPathID error", "error", err)
		RespondWithError(w, http.StatusBadRequest, "UserID required")
		return
	}

	if userID != userIDParam {
		h.Logger.Error("UserID mismatch", "got", userIDParam, "want", userID)
		RespondWithError(w, http.StatusUnauthorized, "Invalid ID")
		return
	}

	userReturned, err := user.GetUser(h.UserRepo, userID)
	if err != nil {
		h.Logger.Error("GetUsers error", "error", err)
		return
	}

	payload := &ReturnVals{
		UserID:   userReturned.ID,
		Username: userReturned.Username,
		Email:    userReturned.Email,
	}

	RespondWithJSON(w, http.StatusOK, payload)
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

	userIDParam, err := GetPathID(r, "/api/users/delete/", h.Logger)
	if err != nil {
		h.Logger.Error("GetPathID error", "error", err)
		RespondWithError(w, http.StatusBadRequest, "UserID required")
		return
	}

	if userID != userIDParam {
		h.Logger.Error("UserID mismatch", "got", userIDParam, "want", userID)
		RespondWithError(w, http.StatusUnauthorized, "Invalid ID")
		return
	}

	userReturned, err := user.GetUser(h.UserRepo, userID)
	if err != nil {
		h.Logger.Error("GetUser error", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to find user")
		return
	}

	err = h.Billing.CancelSubscription(h.UserRepo, userReturned.Email)
	if err != nil {
		h.Logger.Error("h.Billing.CancelSubscription failed", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	err = user.MarkUserDeleted(h.UserRepo, userID)
	if err != nil {
		h.Logger.Error("DeleteUser failed", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	err = token.DeleteRefreshToken(h.TokenRepo, userID)
	if err != nil {
		h.Logger.Error("DeleteRefreshTokensForUser failed", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	err = h.Mailer.SendDeletionConfirmation(userReturned.Email)
	if err != nil {
		h.Logger.Error("h.Mailer.SendDeletionConfirmation failed", "error", err)
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
		h.Logger.Error("Decoding params failed", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	if params.Email == "" || params.Password == "" {
		h.Logger.Error("Invalid username or password.")
		RespondWithError(w, http.StatusBadRequest, "Authentication failed.")
		return

	}

	jwToken, username, userID, err := user.LoginUser(h.UserRepo, params.Email, params.Password)
	if err != nil {
		h.Logger.Error("LoginUser error", "error", err)
		if errors.Is(err, user.ErrAccountDeleted) {
			RespondWithError(w, http.StatusUnauthorized, user.ErrAccountDeleted.Error())
			return
		}
		RespondWithError(w, http.StatusUnauthorized, "Authentication failed.")
		return
	}

	refreshToken, err := token.CreateRefreshToken(h.TokenRepo, userID)
	if err != nil {
		h.Logger.Error("RefreshToken error", "error", err)
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
		h.Logger.Error("http.NewRequest failed", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	githubResp, err := client.Do(req)
	if err != nil {
		h.Logger.Error("GET api.github.com/user failed", "error", err)
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
			h.Logger.Error("http.NewRequest failed", "error", err)
			RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
		emailResp, err := client.Do(req)
		if err != nil {
			h.Logger.Error("GET api.github.com/user/emails failed", "error", err)
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
		h.Logger.Error("GitHub login failed: no verified email found for user", "userEmail", githubUser.Login)
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
		h.Logger.Error("token.CreateJWT failed", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	refreshToken, err := token.CreateRefreshToken(h.TokenRepo, user.ID)
	if err != nil {
		h.Logger.Error("token.CreateRefreshToken failed", "error", err)
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

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		RespondWithError(w, http.StatusUnauthorized, "Missing or invalid Authorization header")
		return
	}
	providedToken := strings.TrimPrefix(authHeader, "Bearer ")
	params := &middleware.AcceptedVals{}

	err := json.NewDecoder(r.Body).Decode(params)
	if err != nil {
		h.Logger.Error("Decoding params failed", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	if params.UserID == 0 {
		h.Logger.Error("Invalid userID")
		RespondWithError(w, http.StatusBadRequest, "Authentication failed")
		return
	}

	storedToken, err := token.GetStoredRefreshToken(h.TokenRepo, params.UserID)
	if err != nil {
		h.Logger.Error("GetStoredRefreshToken error", "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	ok := token.VerifyRefreshToken(storedToken, providedToken)
	if !ok {
		h.Logger.Error("VerifyRefreshToken error")
		RespondWithError(w, http.StatusUnauthorized, "Refresh token is invalid")
		return
	}

	refreshToken, err := token.CreateRefreshToken(h.TokenRepo, params.UserID)
	if err != nil {
		h.Logger.Error("CreateRefreshToken error", "error", err)
		RespondWithError(w, http.StatusUnauthorized, "")
		return
	}

	user, err := h.UserRepo.GetUser(params.UserID)
	if err != nil {
		h.Logger.Error("h.UserRepo.GetUser error", "error", err)
		RespondWithError(w, http.StatusUnauthorized, "Account deactivated")
		return
	}
	if user.AccountStatus == "deleted" {
		h.Logger.Error("Refresh attempt for deleted account ID", "userID", params.UserID)
		RespondWithError(w, http.StatusUnauthorized, "Account deactivated")
		return
	}

	jwToken, err := token.CreateJWT(strconv.Itoa(params.UserID), 0)
	if err != nil {
		h.Logger.Error("JWT creation failed", "error", err)
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
		h.Logger.Error("Decoding params failed", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	userReturned, err := user.GetUser(h.UserRepo, userID)
	if err != nil {
		h.Logger.Error("GetUser error", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to find user")
		return
	}

	interviewStarted, err := h.InterviewService.StartInterview(
		userReturned,
		30,
		3,
		"easy",
		params.JD)
	if err != nil {
		var openaiErr *chatgpt.OpenAIError
		if errors.As(err, &openaiErr) {
			h.Logger.Error("OpenAI error", "error", openaiErr)
			RespondWithError(w, openaiErr.StatusCode, openaiErr.Message)
			return
		}
		if errors.Is(err, interview.ErrNoValidCredits) {
			RespondWithError(w, http.StatusPaymentRequired, "You do not have enough credits to start a new interview or your subscription has expired.")
			return
		}
		h.Logger.Error("Interview failed to start", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to start interview.")
		return
	}

	conversationID, err := conversation.CreateEmptyConversation(h.ConversationRepo, interviewStarted.Id, interviewStarted.Subtopic)
	if err != nil {
		h.Logger.Error("conversation.CreateEmptyConversation failed", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	err = h.InterviewService.LinkConversation(interviewStarted.Id, conversationID)
	if err != nil {
		h.Logger.Error("interview.LinkConversation failed", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
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

	interviewID, err := GetPathID(r, "/api/interviews/", h.Logger)
	if err != nil {
		h.Logger.Error("GetPathID failed", "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid interview ID")
		return
	}

	interviewReturned, err := h.InterviewService.GetInterview(interviewID)
	if err != nil {
		h.Logger.Error("GetInterview failed", "error", err)
		RespondWithError(w, http.StatusNotFound, "Interview not found")
		return
	}
	if interviewReturned.UserId != userID {
		h.Logger.Error("User ID mismatch on interview fetch", "got", userID, "want", interviewReturned.UserId)
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

	interviewID, err := GetPathID(r, "/api/interviews/", h.Logger)
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

	interviewReturned, err := h.InterviewService.GetInterview(interviewID)
	if err != nil {
		h.Logger.Error("GetInterview error", "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	err = ValidateInterviewStatusTransition(interviewReturned.Status, payload.Status)
	if err != nil {
		h.Logger.Error("ValidateInterviewStatusTransition failed", "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid status transition")
		return
	}

	err = h.InterviewService.InterviewRepo.UpdateStatus(interviewID, userID, payload.Status)
	if err != nil {
		h.Logger.Error("UpdateInterviewStatus failed", "error", err)
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
		h.Logger.Error("Decoding params failed", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	interviewID, err := GetPathID(r, "/api/conversations/create/", h.Logger)

	if err != nil {
		h.Logger.Error("PathID error", "error", err)
		RespondWithError(w, http.StatusBadRequest, "Missing ID")
		return
	}

	interviewReturned, err := h.InterviewService.GetInterview(interviewID)
	if err != nil {
		h.Logger.Error("GetInterview error", "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	if interviewReturned.UserId != userID {
		h.Logger.Error("userID does not exist", "got", userID, "want", interviewReturned.UserId)
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if interviewReturned.Status != "active" {
		RespondWithError(w, http.StatusConflict, "Interview is not active")
		return
	}

	conversationReturned, err := conversation.GetConversation(h.ConversationRepo, interviewID)
	if err != nil {
		h.Logger.Error("conversation.GetConversation failed", "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
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
			h.Logger.Error("OpenAI error", "error", openaiErr)
			RespondWithError(w, openaiErr.StatusCode, openaiErr.Message)
			return
		}
		h.Logger.Error("CreateConversation error", "error", err)
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
		h.Logger.Error("Decoding params failed", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	if params.Message == "" {
		h.Logger.Error("messageUserResponse is nil")
		RespondWithError(w, http.StatusBadRequest, "Missing message")
		return
	}

	interviewID, err := GetPathID(r, "/api/conversations/append/", h.Logger)
	if err != nil {
		h.Logger.Error("PathID error", "error", err)
		RespondWithError(w, http.StatusBadRequest, "Missing ID")
		return
	}

	interviewReturned, err := h.InterviewService.GetInterview(interviewID)
	if err != nil {
		h.Logger.Error("GetInterview error", "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	if interviewReturned.UserId != userID {
		h.Logger.Error("incorrect userID", "got", userID, "want", interviewReturned.UserId)
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if interviewReturned.Status != "active" {
		RespondWithError(w, http.StatusConflict, "Interview is not active")
		return
	}

	conversationReturned, err := conversation.GetConversation(h.ConversationRepo, interviewID)
	if err != nil {
		h.Logger.Error("GetConversation error", "error", err)
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
			h.Logger.Error("OpenAI error", "error", openaiErr)
			RespondWithError(w, openaiErr.StatusCode, openaiErr.Message)
			return
		}
		h.Logger.Error("AppendConversation error", "error", err)
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

	interviewID, err := GetPathID(r, "/api/conversations/", h.Logger)
	if err != nil {
		h.Logger.Error("PathID error", "error", err)
		RespondWithError(w, http.StatusBadRequest, "Missing ID")
		return
	}

	interviewReturned, err := h.InterviewService.GetInterview(interviewID)
	if err != nil {
		h.Logger.Error("GetInterview error", "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	if interviewReturned.UserId != userID {
		h.Logger.Error("incorrect userID", "got", userID, "want", interviewReturned.UserId)
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	conversationReturned, err := conversation.GetConversation(h.ConversationRepo, interviewID)
	if err != nil {
		h.Logger.Error("GetConversation error", "error", err)
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
		h.Logger.Error("Decoding request failed", "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resetJWT, err := user.RequestPasswordReset(h.UserRepo, params.Email)
	if err != nil {
		h.Logger.Error("Error generating reset token for email", "error", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	resetURL := frontendURL + "reset-password?token=" + resetJWT

	go func(email, resetURL string) {
		err := h.Mailer.SendPasswordReset(email, resetURL)
		if err != nil {
			h.Logger.Error("SendPasswordReset error", "error", err)
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
		h.Logger.Error("Decoding payload failed", "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	err := user.ResetPassword(h.UserRepo, params.NewPassword, params.Token)
	if err != nil {
		h.Logger.Error("ResetPasswordHandler failed", "error", err)
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
		h.Logger.Error("r.Context().Value() error")
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var params CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil || params.Tier == "" {
		h.Logger.Error("jsonNewDecoder failed", "error", err)
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
		h.Logger.Error("strconv.Atoi() failed", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	url, err := h.Billing.RequestCheckoutSession(user.Email, priceIDInt)
	if err != nil {
		h.Logger.Error("billing.CreateCheckoutSession failed", "error", err)
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
		h.Logger.Error("GetUser failed", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Could not retrieve user")
		return
	}

	err = h.Billing.RequestDeleteSubscription(userReturned.SubscriptionID)
	if err != nil {
		h.Logger.Error("DeleteSubscription failed", "error", err)
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
		h.Logger.Error("GetUser failed", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Could not retrieve user")
		return
	}

	err = h.Billing.RequestResumeSubscription(userReturned.SubscriptionID)
	if err != nil {
		h.Logger.Error("DeleteSubscription failed", "error", err)
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
		h.Logger.Error("jsonNewDecoder failed", "error", err)
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
		h.Logger.Error("strconv.Atoi() failed", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if err := h.Billing.RequestUpdateSubscriptionVariant(user.SubscriptionID, priceIDInt); err != nil {
		h.Logger.Error("UpdateLemonSubscriptionVariant failed", "error", err)
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
		h.Logger.Error("io.ReadAll failed", "error", err)
		RespondWithError(w, http.StatusBadRequest, "Bad Request")
		return
	}
	defer r.Body.Close()

	signature := r.Header.Get("X-Signature")
	if !h.Billing.VerifyBillingSignature(signature, body, os.Getenv("LEMON_WEBHOOK_SECRET")) {
		h.Logger.Error("Invalid billing event signature")
		RespondWithError(w, http.StatusUnauthorized, "Invalid signature")
		return
	}

	var webhookPayload billing.BillingWebhookPayload
	err = json.Unmarshal(body, &webhookPayload)
	if err != nil {
		h.Logger.Error("json.Unmarshal failed", "error", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}
	subscriptionID := webhookPayload.Data.SubscriptionID
	webhookID := webhookPayload.Meta.WebhookID
	exists, err := h.BillingRepo.HasWebhookBeenProcessed(webhookID)
	if err != nil {
		h.Logger.Error("h.BillingRepo.HasWebhookBeenProcessed failed", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Error checking webhook")
		return
	}
	if exists {
		h.Logger.Info("Webhook already processed", "webhookID", webhookID)
		w.WriteHeader(http.StatusOK)
		return
	}

	var emailAttribute struct {
		UserEmail string `json:"user_email"`
	}

	eventType := webhookPayload.Meta.EventName

	h.Logger.Info("Received webhook", "eventType", eventType, "webhookID", webhookID, "subscriptionID", subscriptionID)

	switch eventType {
	case "order_created":
		var orderAttrs billing.OrderAttributes
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &orderAttrs); err != nil {
			h.Logger.Error("Unmarshal order_created failed", "error", err)
			RespondWithError(w, http.StatusBadRequest, "Invalid order_created payload")
			return
		}

		err = h.Billing.ApplyCredits(h.UserRepo, h.BillingRepo, orderAttrs.UserEmail, orderAttrs.FirstOrderItem.VariantID)
		if err != nil {
			h.Logger.Error("h.Billing.ApplyCredits failed", "error", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
			return
		}
	case "subscription_created":
		var SubCreatedAttrs billing.SubscriptionAttributes
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &SubCreatedAttrs); err != nil {
			h.Logger.Error("Unmarshal subscription_created failed", "error", err)
			RespondWithError(w, http.StatusBadRequest, "Invalid subscription_created payload")
			return
		}

		exists, err := h.UserRepo.HasActiveOrCancelledSubscription(SubCreatedAttrs.UserEmail)
		if err != nil {
			h.Logger.Error("Subscription duplicate check failed", "error", err)
			RespondWithError(w, http.StatusInternalServerError, "Subscription check failed")
			return
		}
		if exists {
			h.Logger.Info("Duplicate subscription attempt blocked", "userEmail", SubCreatedAttrs.UserEmail)
			return
		}

		err = h.Billing.CreateSubscription(h.UserRepo, SubCreatedAttrs, subscriptionID)
		if err != nil {
			h.Logger.Error("h.Billing.CreateSubscription failed", "error", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
			return
		}
	case "subscription_cancelled":
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &emailAttribute); err != nil {
			h.Logger.Error("Unmarshal subscription_cancelled failed", "error", err)
			RespondWithError(w, http.StatusBadRequest, "Invalid subscription_cancelled payload")
			return
		}

		err = h.Billing.CancelSubscription(h.UserRepo, emailAttribute.UserEmail)
		if err != nil {
			h.Logger.Error("h.Billing.CancelSubscription failed", "error", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
			return
		}
	case "subscription_resumed":
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &emailAttribute); err != nil {
			h.Logger.Error("Unmarshal subscription_resumed failed", "error", err)
			RespondWithError(w, http.StatusBadRequest, "Invalid subscription_resumed payload")
			return
		}

		err = h.Billing.ResumeSubscription(h.UserRepo, emailAttribute.UserEmail)
		if err != nil {
			h.Logger.Error("h.Billing.ResumeSubscription failed", "error", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
			return
		}
	case "subscription_expired":
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &emailAttribute); err != nil {
			h.Logger.Error("Unmarshal subscription_expired failed", "error", err)
			RespondWithError(w, http.StatusBadRequest, "Invalid subscription_expired payload")
			return
		}

		err = h.Billing.ExpireSubscription(h.UserRepo, h.BillingRepo, emailAttribute.UserEmail)
		if err != nil {
			h.Logger.Error("h.Billing.ExpireSubscription failed", "error", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
			return
		}
	case "subscription_payment_success":
		var SubRenewAttrs billing.SubscriptionRenewAttributes
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &SubRenewAttrs); err != nil {
			h.Logger.Error("Unmarshal subscription_payment_success failed", "error", err)
			RespondWithError(w, http.StatusBadRequest, "Invalid subscription_payment_success payload")
			return
		}

		if SubRenewAttrs.BillingReason == "initial" {
			h.Logger.Info("Skipping credits on initial charge (already granted via order_created)")
			return
		}

		err = h.Billing.RenewSubscription(h.UserRepo, h.BillingRepo, SubRenewAttrs)
		if err != nil {
			h.Logger.Error("h.Billing.RenewSubscription failed", "error", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
			return
		}
	case "subscription_plan_changed":
		var SubChangedAttrs billing.SubscriptionAttributes
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &SubChangedAttrs); err != nil {
			h.Logger.Error("Unmarshal subscription_plan_changed failed", "error", err)
			RespondWithError(w, http.StatusBadRequest, "Invalid subscription_plan_changed payload")
			return
		}

		err = h.Billing.ChangeSubscription(h.UserRepo, h.BillingRepo, SubChangedAttrs)
		if err != nil {
			h.Logger.Error("h.Billing.ChangeSubscription failed", "error", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
			return
		}
	case "subscription_updated":
		var SubChangedAttrs billing.SubscriptionAttributes
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &SubChangedAttrs); err != nil {
			h.Logger.Error("Unmarshal subscription_updated failed", "error", err)
			RespondWithError(w, http.StatusBadRequest, "Invalid subscription_updated payload")
			return
		}

		err = h.Billing.UpdateSubscription(h.UserRepo, SubChangedAttrs, subscriptionID)
		if err != nil {
			h.Logger.Error("h.Billing.UpdateSubscription failed", "error", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
			return
		}
	case "order_refunded":
		var orderAttrs billing.OrderAttributes
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &orderAttrs); err != nil {
			h.Logger.Error("Unmarshal order_created failed", "error", err)
			RespondWithError(w, http.StatusBadRequest, "Invalid order_created payload")
			return
		}

		err = h.Billing.DeductCredits(h.UserRepo, h.BillingRepo, orderAttrs)
		if err != nil {
			h.Logger.Error("h.Billing.DeductCredits failed", "error", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
			return
		}
	case "subscription_payment_failed", "subscription_payment_recovered":
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &emailAttribute); err != nil {
			h.Logger.Error("Unmarshal failed", "eventType", eventType, "error", err)
			RespondWithError(w, http.StatusBadRequest, "Invalid payment status payload")
			return
		}
		h.Logger.Info("Payment event", "eventType", eventType, "user", emailAttribute.UserEmail)
	default:
		h.Logger.Info("Unhandled event type", "eventType", eventType)
		RespondWithError(w, http.StatusNotImplemented, "Unhandled event type")
		return
	}

	if err != nil {
		h.Logger.Error("eventType switch func failed", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	err = h.BillingRepo.MarkWebhookProcessed(webhookID, eventType)
	if err != nil {
		h.Logger.Error("MarkWebhookProcessed failed", "error", err)
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
		h.Logger.Error("dashboard.GetDashboardData failed", "error", err)
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
			h.Logger.Error("OpenAI error", "error", openaiErr)
			RespondWithError(w, openaiErr.StatusCode, openaiErr.Message)
			return
		}
		h.Logger.Error("chatgpt.ExtractJDInput failed", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to process job description")
		return
	}

	jdSummary, err := h.OpenAI.ExtractJDSummary(jdInput)
	if err != nil {
		var openaiErr *chatgpt.OpenAIError
		if errors.As(err, &openaiErr) {
			h.Logger.Error("OpenAI error", "error", openaiErr)
			RespondWithError(w, openaiErr.StatusCode, openaiErr.Message)
			return
		}
		h.Logger.Error("chatgpt.ExtractJDSummary failed", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to process job description")
		return

	}

	RespondWithJSON(w, http.StatusOK, jdSummary)
}
