package dashboard

import (
	"log/slog"
	"time"

	"github.com/michaelboegner/interviewer/interview"
	"github.com/michaelboegner/interviewer/user"
)

type DashboardData struct {
	Email                 string              `json:"email"`
	Plan                  string              `json:"plan"`
	Status                string              `json:"status"`
	SubscriptionStartDate *time.Time          `json:"subscription_start_date"`
	SubscriptionEndDate   *time.Time          `json:"subscription_end_date"`
	IndividualCredits     int                 `json:"individual_credits"`
	SubscriptionCredits   int                 `json:"subscription_credits"`
	PastInterviews        []interview.Summary `json:"past_interviews"`
}

type DashboardService struct {
	UserRepo      user.UserRepo
	InterviewRepo interview.InterviewRepo
	Logger        *slog.Logger
}

func NewDashboardService(userRepo user.UserRepo, interviewRepo interview.InterviewRepo, logger *slog.Logger) *DashboardService {
	return &DashboardService{
		UserRepo:      userRepo,
		InterviewRepo: interviewRepo,
		Logger:        logger,
	}
}
