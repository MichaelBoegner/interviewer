package dashboard

import (
	"log"
)

func (d *DashboardService) GetDashboardData(userID int) (*DashboardData, error) {
	user, err := d.UserRepo.GetUser(userID)
	if err != nil {
		log.Printf("GetUser failed for userID %d: %v", userID, err)
		return nil, err
	}

	interviews, err := d.InterviewRepo.GetInterviewSummariesByUserID(userID)
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
