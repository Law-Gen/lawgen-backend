package domain

import "time"

type PasswordResetOTP struct {
	Email     string
	OTP     string
	ExpiresAt time.Time
}
type PasswordResetToken struct {
	Email     string
	Token     string
	ExpiresAt time.Time
}


type RefreshToken struct {
	UserID    string
	Token     string
	ExpiresAt time.Time
}