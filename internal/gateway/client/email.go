package client

import (
	"encoding/base64"
	"fmt"
	"github.com/Nzyazin/zadnik.store/internal/common"
	"net/smtp"
	"time"
	"regexp"
)

type SMTPEmailSender struct {
	Host string
	Port int
	From string
	Password string
	Logger common.Logger
}

func NewSMTPEmailSender(host string, port int, from string, password string, logger common.Logger) *SMTPEmailSender {
	return &SMTPEmailSender{
		Host: host,
		Port: port,
		From: from,
		Password: password,
		Logger: logger,
	}
}

func (s *SMTPEmailSender) SendOrder(name, phone string) error {
	subject := "Заказ задника для обуви"
	body := fmt.Sprintf(
		"Получена новая заявка:\n\n"+
		"Имя: %s\n"+
		"Телефон: %s\n"+
		"Дата: %s\n",
		name,
		phone,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	err := s.sendEmail(s.From, subject, body)
	if err != nil {
		return fmt.Errorf("failed to send order email: %v", err)
	}

	s.Logger.Infof("Order email sent to %s", s.From)
	return nil
}

func (s *SMTPEmailSender) sendEmail(email, subject, body string) error {
	var message []byte
	// Кодируем тему в Base64 для корректного отображения в почтовых клиентах
	encodedSubject := base64.StdEncoding.EncodeToString([]byte(subject))
	
	message = fmt.Appendf(message, "From: %s\r\n", s.From)
	message = fmt.Appendf(message, "To: %s\r\n", email)
	message = fmt.Appendf(message, "Subject: =?UTF-8?B?%s?=\r\n", encodedSubject)
	message = fmt.Appendf(message, "MIME-Version: 1.0\r\n")
	message = fmt.Appendf(message, "Content-Type: text/plain; charset=utf-8\r\n")
	message = fmt.Appendf(message, "Content-Transfer-Encoding: 8bit\r\n")
	message = fmt.Appendf(message, "\r\n%s\r\n", body)

	auth := smtp.PlainAuth("", s.From, s.Password, s.Host)

	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", s.Host, s.Port),
		auth,
		s.From,
		[]string{s.From},
		message,
	)

	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	s.Logger.Infof("Email sent to %s", email)
	return nil
}

func isValidPhone(phone string) bool {
	// Проверяем длину телефона
	if len(phone) < 6 || len(phone) > 15 {
		return false
	}

	// Проверяем соответствие регулярному выражению
	pattern := `^[+]?([0-9]+(\(|\)|\-|\s)?)+$`
	re := regexp.MustCompile(pattern)
	if !re.MatchString(phone) {
		return false
	}

	// Проверяем, что в телефоне есть хотя бы 6 цифр
	digitsRe := regexp.MustCompile(`\D`)
	digitsOnly := digitsRe.ReplaceAllString(phone, "")

	return len(digitsOnly) >= 6
}