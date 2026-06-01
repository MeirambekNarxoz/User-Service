package services

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"time"
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
		return fmt.Errorf("SMTP credentials not configured (SMTP_USER/SMTP_PASS), skipping email delivery")
	}

	auth := smtp.PlainAuth("", s.smtpUser, s.smtpPass, s.host)

	// Construct MIME formatting
	// Important: CRLF (\r\n) is required for SMTP headers
	header := make(map[string]string)
	header["From"] = fmt.Sprintf("%s <%s>", s.smtpSender, s.smtpUser)
	header["To"] = toEmail
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=\"UTF-8\""

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + htmlBody

	address := fmt.Sprintf("%s:%s", s.host, s.port)

	conn, err := (&net.Dialer{Timeout: 10 * time.Second}).Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP %s: %w", address, err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: s.host}); err != nil {
			return fmt.Errorf("SMTP STARTTLS failed: %w", err)
		}
	}

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth failed: %w", err)
	}
	if err := client.Mail(s.smtpUser); err != nil {
		return err
	}
	if err := client.Rcpt(toEmail); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(message)); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
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
