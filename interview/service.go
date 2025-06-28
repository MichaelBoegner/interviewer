package interview

import (
	"fmt"
	"log"
	"time"

	"github.com/michaelboegner/interviewer/billing"
	"github.com/michaelboegner/interviewer/chatgpt"
	"github.com/michaelboegner/interviewer/user"
)

func StartInterview(
	interviewRepo InterviewRepo,
	userRepo user.UserRepo,
	billingRepo billing.BillingRepo,
	ai chatgpt.AIClient,
	user *user.User,
	length,
	numberQuestions int,
	difficulty string,
	jd string) (*Interview, error) {

	err := deductAndLogCredit(user, userRepo, billingRepo)
	if err != nil {
		log.Printf("checkCreditsLogTransaction failed: %v", err)
		return nil, err
	}

	now := time.Now().UTC()
	jdSummary := ""

	if jd != "" {
		jdInput, err := ai.ExtractJDInput(jd)
		if err != nil {
			fmt.Printf("ai.ExtractJDInput() failed: %v", err)
			return nil, err
		}
		jdSummary, err = ai.ExtractJDSummary(jdInput)
		if err != nil {
			fmt.Printf("ai.ExtractJDSummary() failed: %v", err)
			return nil, err
		}
	}

	prompt := chatgpt.BuildPrompt([]string{}, "Introduction", 1, jdSummary)

	chatGPTResponse, err := ai.GetChatGPTResponse(prompt)
	if err != nil {
		log.Printf("getChatGPTResponse err: %v\n", err)
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

	id, err := interviewRepo.CreateInterview(interview)
	if err != nil {
		log.Printf("CreateInterview err: %v", err)
		return nil, err
	}
	interview.Id = id

	return interview, nil
}

func LinkConversation(interviewRepo InterviewRepo, interviewID, conversationID int) error {
	err := interviewRepo.LinkConversation(interviewID, conversationID)
	if err != nil {
		log.Printf("interviewRepo.LinkConversation failed: %v", err)
		return err
	}

	return nil
}

func GetInterview(interviewRepo InterviewRepo, interviewID int) (*Interview, error) {
	interview, err := interviewRepo.GetInterview(interviewID)
	if err != nil {
		return nil, err
	}

	return interview, nil
}

func canUseCredit(user *user.User) (string, error) {
	now := time.Now()

	switch {
	case user.SubscriptionEndDate != nil &&
		user.SubscriptionEndDate.After(now) &&
		user.SubscriptionStatus != "expired" &&
		user.SubscriptionCredits > 0:
		return "subscription", nil
	case user.IndividualCredits > 0:
		return "individual", nil
	default:
		return "", ErrNoValidCredits
	}
}

func deductAndLogCredit(user *user.User, userRepo user.UserRepo, billingRepo billing.BillingRepo) error {
	creditType, err := canUseCredit(user)
	if err != nil {
		log.Print("canUseCredit failed", err)
		return err
	}
	if creditType != "" {

	}

	err = userRepo.AddCredits(user.ID, -1, creditType)
	if err != nil {
		log.Printf("AddCredits failed: %v", err)
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
		log.Printf("billingRepo.LogCreditTransaction failed: %v", err)
		return err
	}

	return nil
}
