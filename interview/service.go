package interview

import (
	"fmt"
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

	err := i.deductAndLogCredit(user)
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

func (i *InterviewService) deductAndLogCredit(user *user.User) error {
	creditType, err := i.canUseCredit(user)
	if err != nil {
		i.Logger.Error("canUseCredit failed", "error", err)
		return err
	}
	if creditType == "" {
		i.Logger.Info("user doesn't have a valid plan or credits")
		return fmt.Errorf("user doesn't have a valid plan or credits")
	}

	err = i.UserRepo.AddCredits(user.ID, -1, creditType)
	if err != nil {
		i.Logger.Error("AddCredits failed", "error", err)
		return err
	}

	reason := "Interview started"
	tx := billing.CreditTransaction{
		UserID:     user.ID,
		Amount:     -1,
		CreditType: creditType,
		Reason:     reason,
	}
	if err := i.BillingRepo.LogCreditTransaction(tx); err != nil {
		i.Logger.Error("billingRepo.LogCreditTransaction failed", "error", err)
		return err
	}

	return nil
}

func (i *InterviewService) canUseCredit(user *user.User) (string, error) {
	now := time.Now()

	switch {
	case user.SubscriptionEndDate != nil &&
		user.SubscriptionEndDate.After(now) &&
		user.SubscriptionStatus != "expired" &&
		user.SubscriptionCredits > 0:
		i.Logger.Info("subscription plan in canUseCredit check")
		return "subscription", nil
	case user.IndividualCredits > 0:
		i.Logger.Info("individual plan in canUseCredit check")
		return "individual", nil
	default:
		i.Logger.Info("no valid credits in canUseCredit check")
		return "", ErrNoValidCredits
	}
}
