package services

import (
	"fmt"
	"net/smtp"
)

type EmailService struct {
	host       string
	port       string
	smtpUser   string
	smtpPass   string
	smtpSender string
}

func NewEmailService(host, port, user, pass, sender string) *EmailService {
	return &EmailService{
		host:       host,
		port:       port,
		smtpUser:   user,
		smtpPass:   pass,
		smtpSender: sender,
	}
}

func (s *EmailService) send(toEmail, subject, htmlBody string) error {
	// If credentials are not set, return error so the app uses the 000000 bypass
	if s.smtpUser == "" || s.smtpPass == "" {
		return fmt.Errorf("SMTP credentials not configured, skipping email delivery")
	}

	auth := smtp.PlainAuth("", s.smtpUser, s.smtpPass, s.host)

	// Construct MIME formatting
	header := fmt.Sprintf("From: %s <%s>\r\n", s.smtpSender, s.smtpUser)
	header += fmt.Sprintf("To: %s\r\n", toEmail)
	header += fmt.Sprintf("Subject: %s\r\n", subject)
	header += "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"

	msg := []byte(header + htmlBody)

	address := fmt.Sprintf("%s:%s", s.host, s.port)
	return smtp.SendMail(address, auth, s.smtpUser, []string{toEmail}, msg)
}

func (s *EmailService) SendVerificationCode(toEmail, code string) error {
	subject := "Код подтверждения регистрации SkillHub"
	body := fmt.Sprintf(
		"<h3>Добро пожаловать в SkillHub!</h3><p>Ваш код подтверждения: <strong style='font-size:28px;letter-spacing:6px'>%s</strong></p><p>Код действителен 10 минут.</p>",
		code,
	)
	return s.send(toEmail, subject, body)
}

func (s *EmailService) SendResetPasswordCode(toEmail, code string) error {
	subject := "Восстановление пароля SkillHub"
	body := fmt.Sprintf(
		"<h3>Сброс пароля</h3><p>Ваш код восстановления: <strong style='font-size:28px;letter-spacing:6px'>%s</strong></p><p>Код действителен 10 минут.</p>",
		code,
	)
	return s.send(toEmail, subject, body)
}
