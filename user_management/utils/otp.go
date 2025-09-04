// utils/otp.go
package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// GenerateOTP generates a 6-digit numeric OTP
func GenerateOTP() string {
	const otpLength = 6
	const max int64 = 9 // To generate digits 0-9

	otp := ""
	for i := 0; i < otpLength; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(max+1))
		if err != nil {
			// Fallback to simpler random if cryptographically secure fails
			// In a real app, you might want more robust error handling
			fmt.Printf("Warning: Failed to generate cryptographically secure digit, falling back. Error: %v\n", err)
			otp += fmt.Sprintf("%d", time.Now().Nanosecond()%10)
		} else {
			otp += fmt.Sprintf("%d", num.Int64())
		}
	}
	return otp
}

// GenerateOTPSecret generates a longer, more complex secret for things like JWT, not the actual user-facing OTP
func GenerateOTPSecret() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32) // 32 characters long
	_, err := rand.Read(b)
	if err != nil {
		// handle error
		panic(err)
	}
	for i := 0; i < 32; i++ {
		b[i] = chars[b[i]%byte(len(chars))]
	}
	return string(b)
}