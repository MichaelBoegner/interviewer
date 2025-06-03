package dashboard

import "github.com/michaelboegner/interviewer/interview"

type DashboardData struct {
	Email               string              `json:"email"`
	Plan                string              `json:"plan"`
	Status              string              `json:"status"`
	IndividualCredits   int                 `json:"individual_credits"`
	SubscriptionCredits int                 `json:"subscription_credits"`
	PastInterviews      []interview.Summary `json:"past_interviews"`
}
