package email

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

type EmailService struct {
	dialer    *gomail.Dialer
	fromEmail string
}

func NewEmailService() *EmailService {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	fromEmail := os.Getenv("FROM_EMAIL")

	port, err := strconv.Atoi(smtpPort)
	if err != nil {
		fmt.Printf("Invalid SMTP port: %v\n", err)
		return nil
	}
	dialer := gomail.NewDialer(smtpHost, port, smtpUser, smtpPass)
	return &EmailService{
		dialer:    dialer,
		fromEmail: fromEmail,
	}
}

func (s *EmailService) SendActivationEmail(toEmail, activationLink string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.fromEmail)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Activate Your Account")
	m.SetBody("text/html", fmt.Sprintf("Hello,<br><br>Please click the following link to activate your account:<br><br><a href='%s'>Activate My Account</a><br><br>Thank you!", activationLink))

	if err := s.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send activation email: %w", err)
	}

	return nil
}

func (s *EmailService) SendPasswordResetEmail(toEmail, resetToken string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.fromEmail)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Password Reset Request")
	m.SetBody("text/html", fmt.Sprintf("Hello,<br><br>You requested a password reset. Please use the following token to set a new password:<br><br><b>%s</b><br><br>This token will expire in 1 hour.<br><br>If you did not request this, please ignore this email.", resetToken))
	if err := s.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}
	return nil
}
