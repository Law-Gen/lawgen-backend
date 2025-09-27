package domain

import "time"

type Feedback struct {
    ID               string    `bson:"_id,omitempty" json:"id"`
    SubmitterUserID  string    `json:"submitter_user_id"`
    Type             string    `json:"type"`
    Description      string    `json:"description"`
    Severity         string    `json:"severity"`
    Timestamp        time.Time `json:"timestamp"`
}