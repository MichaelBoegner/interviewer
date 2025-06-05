package dashboard

import (
	"log"

	"github.com/michaelboegner/interviewer/interview"
	"github.com/michaelboegner/interviewer/user"
)

func GetDashboardData(userID int, userRepo user.UserRepo, interviewRepo interview.InterviewRepo) (*DashboardData, error) {
	user, err := userRepo.GetUser(userID)
	if err != nil {
		log.Printf("GetUser failed for userID %d: %v", userID, err)
		return nil, err
	}

	interviews, err := interviewRepo.GetInterviewSummariesByUserID(userID)
	if err != nil {
		log.Printf("GetInterviewSummariesByUserID failed for userID %d: %v", userID, err)
		return nil, err
	}

	return &DashboardData{
		Email:                 user.Email,
		Plan:                  user.SubscriptionTier,
		Status:                user.SubscriptionStatus,
		SubscriptionStartDate: user.SubscriptionStartDate,
		SubscriptionEndDate:   user.SubscriptionEndDate,
		IndividualCredits:     user.IndividualCredits,
		SubscriptionCredits:   user.SubscriptionCredits,
		PastInterviews:        interviews,
	}, nil
}
