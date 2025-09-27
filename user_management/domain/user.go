package domain

import (
	"time"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
)


// User represents the core user entity in the system
type User struct {
	ID           string
	FullName     string
	Email        string
	Password     string
	Role         string
	Activated    bool
	Profile      UserProfile
	SubscriptionStatus string 
	CreatedAt    time.Time
	UpdatedAt    time.Time 
}

// UserProfile represents embedded profile data
type UserProfile struct {
	Gender 			string
	BirthDate 		time.Time			
	ProfilePictureURL string 
	LanguagePreference string
}

// UnactivatedUser represents a user who has not yet activated their account
type UnactivatedUser struct {
	ID                    string 
	FullName              string
	Email                 string
	Password              string 
	Activated			  bool 
	ActivationToken       string  
	ActivationTokenExpiry *time.Time  
	CreatedAt             time.Time 
	UpdatedAt             time.Time 
}