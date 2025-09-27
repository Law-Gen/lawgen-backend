package domain

// SubscriptionPlan is the core domain model, free of any implementation details.
type SubscriptionPlan struct {
	ID    string
	Name  string
	Price int
}