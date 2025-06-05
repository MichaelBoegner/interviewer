package dashboard

import (
	"time"

	"github.com/michaelboegner/interviewer/interview"
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
