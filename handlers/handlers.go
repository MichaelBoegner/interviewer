package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/michaelboegner/interviewer/billing"
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

func (h *Handler) CreateUsersHandler(w http.ResponseWriter, r *http.Request) {
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
		RespondWithError(w, http.StatusInternalServerError, "Internal server error")
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

	jwToken, userID, err := user.LoginUser(h.UserRepo, params.Email, params.Password)
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
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value(middleware.ContextKeyTokenParams).(int)
	if !ok {
		RespondWithError(w, http.StatusBadRequest, "Invalid userID")
		return
	}

	userReturned, err := user.GetUser(h.UserRepo, userID)
	if err != nil {
		log.Printf("GetUser error: %v", err)
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
		"easy")
	if err != nil {
		log.Printf("Interview failed to start: %v", err)
		if errors.Is(err, interview.ErrNoValidCredits) {
			RespondWithError(w, http.StatusPaymentRequired, "You do not have enough credits to start a new interview.")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to start interview.")
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

	url, err := h.Billing.CreateCheckoutSession(user.Email, priceIDInt)
	if err != nil {
		log.Printf("billing.CreateCheckoutSession failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Could not start checkout")
		return
	}

	RespondWithJSON(w, http.StatusOK, CheckoutResponse{CheckoutURL: url})
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

	var webhookPayload billing.BillingWebhookPayload
	err = json.Unmarshal(body, &webhookPayload)
	if err != nil {
		log.Printf("json.Unmarshal failed: %v", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

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
	switch eventType {
	case "order_created":
		var orderCreatedAttributes billing.OrderCreatedAttributes
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &orderCreatedAttributes); err != nil {
			log.Printf("Unmarshal order_created failed: %v", err)
			RespondWithError(w, http.StatusBadRequest, "Invalid order_created payload")
			return
		}

		err = h.Billing.ApplyCredits(h.UserRepo, h.BillingRepo, orderCreatedAttributes.UserEmail, orderCreatedAttributes.FirstOrderItem.VariantID)
	case "subscription_created":
		var SubCreatedAttrs billing.SubscriptionAttributes
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &SubCreatedAttrs); err != nil {
			log.Printf("Unmarshal subscription_created failed: %v", err)
			RespondWithError(w, http.StatusBadRequest, "Invalid subscription_created payload")
			return
		}

		err = h.Billing.CreateSubscription(h.UserRepo, SubCreatedAttrs)
	case "subscription_cancelled":
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &emailAttribute); err != nil {
			log.Printf("Unmarshal subscription_cancelled failed: %v", err)
			RespondWithError(w, http.StatusBadRequest, "Invalid subscription_cancelled payload")
			return
		}

		err = h.Billing.CancelSubscription(h.UserRepo, emailAttribute.UserEmail)
	case "subscription_resumed":
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &emailAttribute); err != nil {
			log.Printf("Unmarshal subscription_resumed failed: %v", err)
			RespondWithError(w, http.StatusBadRequest, "Invalid subscription_resumed payload")
			return
		}

		err = h.Billing.ResumeSubscription(h.UserRepo, emailAttribute.UserEmail)
	case "subscription_expired":
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &emailAttribute); err != nil {
			log.Printf("Unmarshal subscription_expired failed: %v", err)
			RespondWithError(w, http.StatusBadRequest, "Invalid subscription_expired payload")
			return
		}

		err = h.Billing.ExpireSubscription(h.UserRepo, emailAttribute.UserEmail)
	case "subscription_payment_success":
		var SubRenewAttrs billing.SubscriptionRenewAttributes
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &SubRenewAttrs); err != nil {
			log.Printf("Unmarshal subscription_payment_success failed: %v", err)
			RespondWithError(w, http.StatusBadRequest, "Invalid subscription_payment_success payload")
			return
		}

		err = h.Billing.RenewSubscription(h.UserRepo, h.BillingRepo, SubRenewAttrs)
	case "subscription_plan_changed":
		var SubChangedAttrs billing.SubscriptionAttributes
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &SubChangedAttrs); err != nil {
			log.Printf("Unmarshal subscription_plan_changed failed: %v", err)
			RespondWithError(w, http.StatusBadRequest, "Invalid subscription_plan_changed payload")
			return
		}

		err = h.Billing.ChangeSubscription(h.UserRepo, h.BillingRepo, SubChangedAttrs)
	case "subscription_updated":
		var SubChangedAttrs billing.SubscriptionAttributes
		if err := json.Unmarshal(webhookPayload.Data.Attributes, &SubChangedAttrs); err != nil {
			log.Printf("Unmarshal subscription_updated failed: %v", err)
			RespondWithError(w, http.StatusBadRequest, "Invalid subscription_updated payload")
			return
		}

		err = h.Billing.UpdateSubscription(h.UserRepo, SubChangedAttrs)
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
		log.Printf("dashboardService.GetDashboardData failed: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Could not load dashboard")
		return
	}

	RespondWithJSON(w, http.StatusOK, dashboardData)
}
