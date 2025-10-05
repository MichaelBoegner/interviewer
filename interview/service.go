package interview

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/michaelboegner/interviewer/billing"
	"github.com/michaelboegner/interviewer/chatgpt"
	"github.com/michaelboegner/interviewer/user"
)

func (i *InterviewService) StartInterview(
	user *user.User,
	length,
	numberQuestions int,
	difficulty string,
	jd string) (*Interview, error) {

	err := deductAndLogCredit(user, i.UserRepo, i.BillingRepo, i.Logger)
	if err != nil {
		i.Logger.Error("checkCreditsLogTransaction failed", "error", err)
		return nil, err
	}

	now := time.Now().UTC()
	jdSummary := ""

	if jd != "" {
		jdInput, err := i.AI.ExtractJDInput(jd)
		if err != nil {
			i.Logger.Error("ai.ExtractJDInput() failed", "error", err)
			return nil, err
		}
		jdSummary, err = i.AI.ExtractJDSummary(jdInput)
		if err != nil {
			i.Logger.Error("ai.ExtractJDSummary() failed", "error", err)
			return nil, err
		}
	}

	prompt := chatgpt.BuildPrompt([]string{}, "Introduction", 1, jdSummary)

	chatGPTResponse, err := i.AI.GetChatGPTResponse(prompt)
	if err != nil {
		i.Logger.Error("getChatGPTResponse err", "error", err)
		return nil, err
	}

	interview := &Interview{
		UserId:          user.ID,
		Length:          length,
		NumberQuestions: numberQuestions,
		Difficulty:      difficulty,
		Status:          "active",
		Score:           100,
		Language:        "Python",
		Prompt:          prompt,
		JDSummary:       jdSummary,
		FirstQuestion:   chatGPTResponse.NextQuestion,
		Subtopic:        chatGPTResponse.Subtopic,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	id, err := i.InterviewRepo.CreateInterview(interview)
	if err != nil {
		i.Logger.Error("CreateInterview err", "error", err)
		return nil, err
	}
	interview.Id = id

	return interview, nil
}

func (i *InterviewService) LinkConversation(interviewID, conversationID int) error {
	err := i.InterviewRepo.LinkConversation(interviewID, conversationID)
	if err != nil {
		i.Logger.Error("interviewRepo.LinkConversation failed", "error", err)
		return err
	}

	return nil
}

func (i *InterviewService) GetInterview(interviewID int) (*Interview, error) {
	interview, err := i.InterviewRepo.GetInterview(interviewID)
	if err != nil {
		i.Logger.Error("interviewRepo.GetInterview failed", "error", err)
		return nil, err
	}

	return interview, nil
}

func canUseCredit(user *user.User, logger *slog.Logger) (string, error) {
	now := time.Now()

	switch {
	case user.SubscriptionEndDate != nil &&
		user.SubscriptionEndDate.After(now) &&
		user.SubscriptionStatus != "expired" &&
		user.SubscriptionCredits > 0:
		logger.Info("subscrtipion plan in canUseCredit check")
		return "subscription", nil
	case user.IndividualCredits > 0:
		logger.Info("individual plan in canUseCredit check")
		return "individual", nil
	default:
		logger.Info("no valid credits in canUseCredit check")
		return "", ErrNoValidCredits
	}
}

func deductAndLogCredit(user *user.User, userRepo user.UserRepo, billingRepo billing.BillingRepo, logger *slog.Logger) error {
	creditType, err := canUseCredit(user, logger)
	if err != nil {
		logger.Error("canUseCredit failed", "error", err)
		return err
	}
	if creditType == "" {
		logger.Info("user doesn't have a valid plan or credits")
		return fmt.Errorf("user doesn't have a valid plan or credits")
	}

	err = userRepo.AddCredits(user.ID, -1, creditType)
	if err != nil {
		logger.Error("AddCredits failed", "error", err)
		return err
	}

	reason := "Interview started"
	tx := billing.CreditTransaction{
		UserID:     user.ID,
		Amount:     -1,
		CreditType: creditType,
		Reason:     reason,
	}
	if err := billingRepo.LogCreditTransaction(tx); err != nil {
		logger.Error("billingRepo.LogCreditTransaction failed", "error", err)
		return err
	}

	return nil
}
