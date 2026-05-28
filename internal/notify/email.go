package notify

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

type Mailer struct {
	host     string
	port     string
	user     string
	password string
	from     string
	to       []string
}

func New() *Mailer {
	return &Mailer{
		host:     os.Getenv("SMTP_HOST"),
		port:     os.Getenv("SMTP_PORT"),
		user:     os.Getenv("SMTP_USER"),
		password: os.Getenv("SMTP_PASSWORD"),
		from:     os.Getenv("SMTP_FROM"),
		to:       strings.Split(os.Getenv("SMTP_TO"), ","), //  if multiple recievers, change from .env
	}
}

func (m *Mailer) SendAlert(to string, metric string, operator string, threshold float64, value float64) error {
	addr := fmt.Sprintf("%s:%s", m.host, m.port)
	auth := smtp.PlainAuth("", m.user, m.password, m.host)

	subject := fmt.Sprintf("Subject: [ALERT] %s threshold breached\r\n", metric)
	body := fmt.Sprintf(
		"Alert triggered:\r\n\r\nMetric:    %s\r\nCondition: %s %.2f\r\nCurrent:   %.2f\r\n",
		metric, operator, threshold, value,
	)
	msg := []byte(subject + "\r\n" + body)

	return smtp.SendMail(addr, auth, m.from, []string{to}, msg) // main sender func
}
func (m *Mailer) SendVerification(to string, verifyURL string) error { // link sender to user's email
	addr := fmt.Sprintf("%s:%s", m.host, m.port)
	auth := smtp.PlainAuth("", m.user, m.password, m.host)

	subject := "Subject: Verify your system-monitor account\r\n"
	body := fmt.Sprintf("Click the link to verify your account:\r\n\r\n%s\r\n\r\nExpires in 24 hours.", verifyURL)
	msg := []byte(subject + "\r\n" + body)

	return smtp.SendMail(addr, auth, m.from, []string{to}, msg)
}
